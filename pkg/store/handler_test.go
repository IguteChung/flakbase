package store

import (
	"context"
	"testing"

	"github.com/IguteChung/flakbase/pkg/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func doc(id ...string) map[string]interface{} {
	data := map[string]interface{}{
		"id1": map[string]interface{}{
			"text":   "value1",
			"const":  "value",
			"number": float64(1),
			"map": map[string]interface{}{
				"key": "value",
			},
		},
		"id2": map[string]interface{}{
			"text":   "value2",
			"const":  "value",
			"number": float64(2),
			"map": map[string]interface{}{
				"key": "value",
			},
		},
		"id3": map[string]interface{}{
			"text":   "value3",
			"const":  "value",
			"number": float64(3),
			"map": map[string]interface{}{
				"key": "value",
			},
		},
		"id4": map[string]interface{}{
			"text":   "value4",
			"const":  "value",
			"number": float64(4),
			"map": map[string]interface{}{
				"key": "value",
			},
		},
	}

	// if id is not specified, return all documents.
	if len(id) == 0 {
		return data
	}
	return data[id[0]].(map[string]interface{})
}

type mockListenChannel struct {
	ch ListenChannel
	t  *testing.T
}

func (c *mockListenChannel) assertOccurs(msg data.ListenMessage) {
	select {
	case m := <-c.ch:
		assert.EqualValues(c.t, msg, m)
	default:
		c.t.FailNow()
	}
}

func (c *mockListenChannel) assertNotOccurs() {
	select {
	case <-c.ch:
		c.t.FailNow()
	default:
	}
}

func newMockListenChannel(t *testing.T) *mockListenChannel {
	return &mockListenChannel{
		t:  t,
		ch: make(ListenChannel, 1),
	}
}

type handlerSuite struct {
	suite.Suite
	handler Handler
}

func TestMemoryHandlerSuite(t *testing.T) {
	handler, err := NewHandler(&Config{})
	if err != nil {
		t.Fatalf("unable to new memory handler: %v", err)
	}
	suite.Run(t, &handlerSuite{
		handler: handler,
	})
}

func TestMongoHandlerSuite(t *testing.T) {
	handler, err := NewHandler(&Config{
		Mongo: "mongodb://localhost:27017",
	})
	if err != nil {
		t.Fatalf("unable to new mongo handler: %v", err)
	}
	suite.Run(t, &handlerSuite{
		handler: handler,
	})
}

func (s handlerSuite) SetupTest() {
	s.NoError(s.handler.Reset(context.Background()))
}

func (s *handlerSuite) TestInsertSingleDocument() {
	ctx := context.Background()
	c1 := newMockListenChannel(s.T())
	c2 := newMockListenChannel(s.T())
	_, err := s.handler.HandleListen(ctx, "/path", data.Query{}, c1.ch)
	s.NoError(err)
	c1.assertOccurs(data.ListenMessage{Ref: "/path"})
	_, err = s.handler.HandleListen(ctx, "/path/id1", data.Query{}, c2.ch)
	s.NoError(err)
	c2.assertOccurs(data.ListenMessage{Ref: "/path/id1"})
	s.NoError(s.handler.HandleSet(ctx, "/path/id1", doc("id1")))
	c1.assertOccurs(data.ListenMessage{Ref: "/path", Data: map[string]interface{}{"id1": doc("id1")}})
	c2.assertOccurs(data.ListenMessage{Ref: "/path/id1", Data: doc("id1")})
	c1.assertNotOccurs()
	c2.assertNotOccurs()

	resp, err := s.handler.HandleGet(ctx, "/path/id1", data.Query{})
	s.NoError(err)
	s.EqualValues(doc("id1"), resp)

	resp, err = s.handler.HandleGet(ctx, "/path/id1/text", data.Query{})
	s.NoError(err)
	s.EqualValues("value1", resp)
}

func (s *handlerSuite) TestInsertMultipleDocuments() {
	ctx := context.Background()
	c1 := newMockListenChannel(s.T())
	c2 := newMockListenChannel(s.T())
	_, err := s.handler.HandleListen(ctx, "/path", data.Query{}, c1.ch)
	s.NoError(err)
	c1.assertOccurs(data.ListenMessage{Ref: "/path"})
	_, err = s.handler.HandleListen(ctx, "/path/id1", data.Query{}, c2.ch)
	s.NoError(err)
	c2.assertOccurs(data.ListenMessage{Ref: "/path/id1"})
	s.NoError(s.handler.HandleUpdate(ctx, "/path", doc()))
	c1.assertOccurs(data.ListenMessage{Ref: "/path", Data: doc()})
	c2.assertOccurs(data.ListenMessage{Ref: "/path/id1", Data: doc("id1")})
	c1.assertNotOccurs()
	c2.assertNotOccurs()

	resp, err := s.handler.HandleGet(ctx, "/path/id1", data.Query{})
	s.NoError(err)
	s.EqualValues(doc("id1"), resp)

	resp, err = s.handler.HandleGet(ctx, "/path/id1/text", data.Query{})
	s.NoError(err)
	s.EqualValues("value1", resp)
}

func (s *handlerSuite) TestInsertDocumentByUpdate() {
	ctx := context.Background()
	s.NoError(s.handler.HandleSet(ctx, "/path/id1", doc("id1")))
	s.NoError(s.handler.HandleUpdate(ctx, "/", map[string]interface{}{
		"/path/id1/text":   "revised",
		"/path/id1/number": nil,
		"/path/id2/text":   "value2",
		"/path/id2/const":  "value",
		"/path/id2/number": float64(2),
		"/path/id2/map": map[string]interface{}{
			"key": "value",
		},
	}))

	resp, err := s.handler.HandleGet(ctx, "/path/id1", data.Query{})
	s.NoError(err)
	s.EqualValues(map[string]interface{}{
		"text":  "revised",
		"const": "value",
		"map": map[string]interface{}{
			"key": "value",
		},
	}, resp)

	resp, err = s.handler.HandleGet(ctx, "/path/id2", data.Query{})
	s.NoError(err)
	s.EqualValues(doc("id2"), resp)
}

func (s *handlerSuite) TestReplaceDocument() {
	ctx := context.Background()
	c1 := newMockListenChannel(s.T())
	c2 := newMockListenChannel(s.T())
	_, err := s.handler.HandleListen(ctx, "/path", data.Query{}, c1.ch)
	s.NoError(err)
	c1.assertOccurs(data.ListenMessage{Ref: "/path"})
	_, err = s.handler.HandleListen(ctx, "/path/id1", data.Query{}, c2.ch)
	s.NoError(err)
	c2.assertOccurs(data.ListenMessage{Ref: "/path/id1"})
	s.NoError(s.handler.HandleSet(ctx, "/path/id1", doc("id1")))
	c1.assertOccurs(data.ListenMessage{Ref: "/path", Data: map[string]interface{}{"id1": doc("id1")}})
	c2.assertOccurs(data.ListenMessage{Ref: "/path/id1", Data: doc("id1")})
	s.NoError(s.handler.HandleSet(ctx, "/path/id1", doc("id2")))
	c1.assertOccurs(data.ListenMessage{Ref: "/path", Data: map[string]interface{}{"id1": doc("id2")}})
	c2.assertOccurs(data.ListenMessage{Ref: "/path/id1", Data: doc("id2")})
	c1.assertNotOccurs()
	c2.assertNotOccurs()

	resp, err := s.handler.HandleGet(ctx, "/path/id1", data.Query{})
	s.NoError(err)
	s.EqualValues(doc("id2"), resp)

	resp, err = s.handler.HandleGet(ctx, "/path/id1/text", data.Query{})
	s.NoError(err)
	s.EqualValues("value2", resp)
}

func (s *handlerSuite) TestUpdateDocumentField() {
	ctx := context.Background()
	c1 := newMockListenChannel(s.T())
	c2 := newMockListenChannel(s.T())
	s.NoError(s.handler.HandleSet(ctx, "/path/id1", doc("id1")))
	_, err := s.handler.HandleListen(ctx, "/path", data.Query{}, c1.ch)
	s.NoError(err)
	c1.assertOccurs(data.ListenMessage{Ref: "/path", Data: map[string]interface{}{"id1": doc("id1")}})
	_, err = s.handler.HandleListen(ctx, "/path/id1", data.Query{}, c2.ch)
	s.NoError(err)
	c2.assertOccurs(data.ListenMessage{Ref: "/path/id1", Data: doc("id1")})
	s.NoError(s.handler.HandleUpdate(ctx, "/path/id1/text", "revised"))

	updatedDoc := map[string]interface{}{
		"id1": map[string]interface{}{
			"text":   "revised",
			"const":  "value",
			"number": float64(1),
			"map": map[string]interface{}{
				"key": "value",
			},
		},
	}
	c1.assertOccurs(data.ListenMessage{Ref: "/path", Data: updatedDoc})
	c2.assertOccurs(data.ListenMessage{Ref: "/path/id1", Data: updatedDoc["id1"]})
	c1.assertNotOccurs()
	c2.assertNotOccurs()

	s.NoError(s.handler.HandleUpdate(ctx, "/path/id1/map/key", "revised"))
	updatedDoc = map[string]interface{}{
		"id1": map[string]interface{}{
			"text":   "revised",
			"const":  "value",
			"number": float64(1),
			"map": map[string]interface{}{
				"key": "revised",
			},
		},
	}
	c1.assertOccurs(data.ListenMessage{Ref: "/path", Data: updatedDoc})
	c2.assertOccurs(data.ListenMessage{Ref: "/path/id1", Data: updatedDoc["id1"]})
	c1.assertNotOccurs()
	c2.assertNotOccurs()

	resp, err := s.handler.HandleGet(ctx, "/path/id1", data.Query{})
	s.NoError(err)
	s.EqualValues(updatedDoc["id1"], resp)

	resp, err = s.handler.HandleGet(ctx, "/path/id1/text", data.Query{})
	s.NoError(err)
	s.EqualValues("revised", resp)
}

func (s *handlerSuite) TestUpdateMultipleDocumentsFields() {
	ctx := context.Background()
	c1 := newMockListenChannel(s.T())
	c2 := newMockListenChannel(s.T())
	s.NoError(s.handler.HandleUpdate(ctx, "/path", doc()))
	_, err := s.handler.HandleListen(ctx, "/path", data.Query{}, c1.ch)
	s.NoError(err)
	c1.assertOccurs(data.ListenMessage{Ref: "/path", Data: doc()})
	_, err = s.handler.HandleListen(ctx, "/path/id1", data.Query{}, c2.ch)
	s.NoError(err)
	c2.assertOccurs(data.ListenMessage{Ref: "/path/id1", Data: doc("id1")})
	s.NoError(s.handler.HandleUpdate(ctx, "/", map[string]interface{}{
		"/path/id1/text": "revised",
		"/path/id2/text": nil,
		"/path/id3/text": map[string]interface{}{"value": "value3"},
	}))

	updatedDoc := map[string]interface{}{
		"id1": map[string]interface{}{
			"text":   "revised",
			"const":  "value",
			"number": float64(1),
			"map": map[string]interface{}{
				"key": "value",
			},
		},
		"id2": map[string]interface{}{
			"const":  "value",
			"number": float64(2),
			"map": map[string]interface{}{
				"key": "value",
			},
		},
		"id3": map[string]interface{}{
			"text":   map[string]interface{}{"value": "value3"},
			"const":  "value",
			"number": float64(3),
			"map": map[string]interface{}{
				"key": "value",
			},
		},
		"id4": map[string]interface{}{
			"text":   "value4",
			"const":  "value",
			"number": float64(4),
			"map": map[string]interface{}{
				"key": "value",
			},
		},
	}
	c1.assertOccurs(data.ListenMessage{Ref: "/path", Data: updatedDoc})
	c2.assertOccurs(data.ListenMessage{Ref: "/path/id1", Data: updatedDoc["id1"]})
	c1.assertNotOccurs()
	c2.assertNotOccurs()

	resp, err := s.handler.HandleGet(ctx, "/path/id1", data.Query{})
	s.NoError(err)
	s.EqualValues(updatedDoc["id1"], resp)

	resp, err = s.handler.HandleGet(ctx, "/path/id1/text", data.Query{})
	s.NoError(err)
	s.EqualValues("revised", resp)
}

func (s *handlerSuite) TestDeleteDocuments() {
	ctx := context.Background()
	c1 := newMockListenChannel(s.T())
	c2 := newMockListenChannel(s.T())
	s.NoError(s.handler.HandleUpdate(ctx, "/path", doc()))
	_, err := s.handler.HandleListen(ctx, "/path", data.Query{}, c1.ch)
	s.NoError(err)
	c1.assertOccurs(data.ListenMessage{Ref: "/path", Data: doc()})
	_, err = s.handler.HandleListen(ctx, "/path/id1", data.Query{}, c2.ch)
	s.NoError(err)
	c2.assertOccurs(data.ListenMessage{Ref: "/path/id1", Data: doc("id1")})
	s.NoError(s.handler.HandleSet(ctx, "/path/id1", nil))

	updatedMap := map[string]interface{}{
		"id2": doc("id2"),
		"id3": doc("id3"),
		"id4": doc("id4"),
	}
	c1.assertOccurs(data.ListenMessage{Ref: "/path", Data: updatedMap})
	c2.assertOccurs(data.ListenMessage{Ref: "/path/id1"})
	c1.assertNotOccurs()
	c2.assertNotOccurs()

	resp, err := s.handler.HandleGet(ctx, "/path/id1", data.Query{})
	s.NoError(err)
	s.EqualValues(nil, resp)

	resp, err = s.handler.HandleGet(ctx, "/path/id1/text", data.Query{})
	s.NoError(err)
	s.EqualValues(nil, resp)
}

func (s *handlerSuite) TestUnlisten() {
	ctx := context.Background()
	c1 := newMockListenChannel(s.T())
	c2 := newMockListenChannel(s.T())
	_, err := s.handler.HandleListen(ctx, "/path", data.Query{}, c1.ch)
	s.NoError(err)
	c1.assertOccurs(data.ListenMessage{Ref: "/path"})
	s.NoError(s.handler.HandleUpdate(ctx, "/path", doc()))
	c1.assertOccurs(data.ListenMessage{Ref: "/path", Data: doc()})
	_, err = s.handler.HandleListen(ctx, "/", data.Query{}, c2.ch)
	s.NoError(err)
	c2.assertOccurs(data.ListenMessage{Ref: "/", Data: map[string]interface{}{"path": doc()}})
	s.NoError(s.handler.HandleUnlisten(ctx, "/path", data.Query{}, c1.ch))
	s.NoError(s.handler.HandleUnlisten(ctx, "/", data.Query{}, c2.ch))
	s.NoError(s.handler.HandleSet(ctx, "/path/id1", nil))
	c1.assertNotOccurs()
	c2.assertNotOccurs()
}

func (s *handlerSuite) TestSetWithLongSize() {
	// max size in mongodb is 120 bytes.
	bytes := make([]byte, 300)
	for i := range bytes {
		bytes[i] = 'a'
	}
	bytes[0] = '/'
	s.NoError(s.handler.HandleUpdate(context.Background(), string(bytes), doc()))
}

func (s *handlerSuite) TestSetDuplicatedPath() {
	ctx := context.Background()
	c1 := newMockListenChannel(s.T())
	c2 := newMockListenChannel(s.T())
	_, err := s.handler.HandleListen(ctx, "/path", data.Query{}, c1.ch)
	s.NoError(err)
	c1.assertOccurs(data.ListenMessage{Ref: "/path"})
	_, err = s.handler.HandleListen(ctx, "/path/id1", data.Query{}, c2.ch)
	s.NoError(err)
	c2.assertOccurs(data.ListenMessage{Ref: "/path/id1"})
	s.NoError(s.handler.HandleSet(ctx, "/path/id1/id1", doc("id1")))
	c1.assertOccurs(data.ListenMessage{Ref: "/path", Data: map[string]interface{}{"id1": map[string]interface{}{"id1": doc("id1")}}})
	c2.assertOccurs(data.ListenMessage{Ref: "/path/id1", Data: map[string]interface{}{"id1": doc("id1")}})
	c1.assertNotOccurs()
	c2.assertNotOccurs()

	resp, err := s.handler.HandleGet(ctx, "/path/id1", data.Query{})
	s.NoError(err)
	s.EqualValues(map[string]interface{}{"id1": doc("id1")}, resp)

	resp, err = s.handler.HandleGet(ctx, "/path/id1/id1", data.Query{})
	s.NoError(err)
	s.EqualValues(doc("id1"), resp)
}

func (s *handlerSuite) testQuery(query data.Query, result map[string]interface{}) {
	ctx := context.Background()
	c := newMockListenChannel(s.T())
	_, err := s.handler.HandleListen(ctx, "/path", query, c.ch)
	s.NoError(err)
	c.assertOccurs(data.ListenMessage{Ref: "/path"})
	s.NoError(s.handler.HandleUpdate(ctx, "/path", doc()))
	c.assertOccurs(data.ListenMessage{Ref: "/path", Data: result})
	c.assertNotOccurs()

	data, err := s.handler.HandleGet(ctx, "/path", query)
	s.NoError(err)
	s.EqualValues(result, data)
}

func (s *handlerSuite) TestOrderByKeyQuery() {
	s.testQuery(data.Query{
		OrderBy: ".key",
		StartAt: "id2",
		EndAt:   "id3",
	}, map[string]interface{}{
		"id2": doc("id2"),
		"id3": doc("id3"),
	})
}

func (s *handlerSuite) TestOrderByKeyLimitQuery() {
	s.testQuery(data.Query{
		OrderBy:    ".key",
		StartAt:    "id2",
		Limit:      2,
		LimitOrder: "l",
	}, map[string]interface{}{
		"id2": doc("id2"),
		"id3": doc("id3"),
	})
}

func (s *handlerSuite) TestOrderByChildPageQuery() {
	s.testQuery(data.Query{
		OrderBy:    "text",
		StartKey:   "id2",
		Limit:      2,
		LimitOrder: "l",
	}, map[string]interface{}{
		"id2": doc("id2"),
		"id3": doc("id3"),
	})
}

func (s *handlerSuite) TestOrderByChildLimitQuery() {
	s.testQuery(data.Query{
		OrderBy:    "text",
		StartAt:    "value2",
		Limit:      2,
		LimitOrder: "r",
	}, map[string]interface{}{
		"id3": doc("id3"),
		"id4": doc("id4"),
	})
}

func (s *handlerSuite) TestShallowQuery() {
	s.testQuery(data.Query{
		Shallow: true,
	}, map[string]interface{}{
		"id1": true,
		"id2": true,
		"id3": true,
		"id4": true,
	})
}

func (s *handlerSuite) TestOrderByChildOnSameValueQuery() {
	s.testQuery(data.Query{
		OrderBy:    "const",
		StartAt:    "value",
		Limit:      2,
		LimitOrder: "l",
	}, map[string]interface{}{
		"id1": doc("id1"),
		"id2": doc("id2"),
	})
}

func (s *handlerSuite) TestLimitQuery() {
	s.testQuery(data.Query{
		Limit:      2,
		LimitOrder: "l",
	}, map[string]interface{}{
		"id1": doc("id1"),
		"id2": doc("id2"),
	})
}

func (s *handlerSuite) TestOrderByNumberQuery() {
	s.testQuery(data.Query{
		OrderBy: "number",
		StartAt: 2,
		EndAt:   3,
	}, map[string]interface{}{
		"id2": doc("id2"),
		"id3": doc("id3"),
	})
}

func (s *handlerSuite) TestOrderByWithoutStartQuery() {
	s.testQuery(data.Query{
		OrderBy:    "number",
		Limit:      2,
		LimitOrder: "l",
	}, map[string]interface{}{
		"id1": doc("id1"),
		"id2": doc("id2"),
	})
}

func (s *handlerSuite) TestOrderPreviousPageQuery() {
	s.testQuery(data.Query{
		OrderBy:    "const",
		EndAt:      "value",
		EndKey:     "id3",
		Limit:      2,
		LimitOrder: "r",
	}, map[string]interface{}{
		"id2": doc("id2"),
		"id3": doc("id3"),
	})
}

func (s *handlerSuite) TestShallowRoot() {
	ctx := context.Background()
	s.NoError(s.handler.HandleUpdate(ctx, "/path1/path2", doc()))

	resp, err := s.handler.HandleGet(ctx, "/", data.Query{Shallow: true})
	s.NoError(err)
	s.EqualValues(map[string]interface{}{"path1": true}, resp)

	resp, err = s.handler.HandleGet(ctx, "/path1", data.Query{Shallow: true})
	s.NoError(err)
	s.EqualValues(map[string]interface{}{"path2": true}, resp)

	resp, err = s.handler.HandleGet(ctx, "/path1/path2", data.Query{Shallow: true})
	s.NoError(err)
	s.EqualValues(map[string]interface{}{
		"id1": true,
		"id2": true,
		"id3": true,
		"id4": true,
	}, resp)

	resp, err = s.handler.HandleGet(ctx, "/path1/path2/id1", data.Query{Shallow: true})
	s.NoError(err)
	s.EqualValues(doc("id1"), resp)
}
