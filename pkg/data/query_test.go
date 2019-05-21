package data

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalListenQuery(t *testing.T) {
	b := []byte(`{"t":"d","d":{"r":10,"a":"q","b":{"p":"/path","t":3,"q":{"sp":5,"sn":"startKey","ep":8,"en":"endKey","i":"child","l":3,"vf":"l"}}}}`)
	var r *Request
	assert.NoError(t, json.Unmarshal(b, &r))
	assert.EqualValues(t, &Request{
		Type:      TypeListen,
		Ref:       "/path",
		RequestID: 10,
		Query: &Query{
			ID:         3,
			StartAt:    float64(5),
			StartKey:   "startKey",
			EndAt:      float64(8),
			EndKey:     "endKey",
			OrderBy:    "child",
			Limit:      3,
			LimitOrder: "l",
		},
	}, r)
}

func TestUnmarshalUnlistenQuery(t *testing.T) {
	b := []byte(`{"t":"d","d":{"r":10,"a":"n","b":{"p":"/path","t":3,"q":{"sp":5,"sn":"startKey","ep":8,"en":"endKey","i":"child","l":3,"vf":"l"}}}}`)
	var r *Request
	assert.NoError(t, json.Unmarshal(b, &r))
	assert.EqualValues(t, &Request{
		Type:      TypeUnlisten,
		Ref:       "/path",
		RequestID: 10,
		Query: &Query{
			ID:         3,
			StartAt:    float64(5),
			StartKey:   "startKey",
			EndAt:      float64(8),
			EndKey:     "endKey",
			OrderBy:    "child",
			Limit:      3,
			LimitOrder: "l",
		},
	}, r)
}

func TestUnmarshalUpdateQuery(t *testing.T) {
	b := []byte(`{"t":"d","d":{"r":10,"a":"m","b":{"p":"/path","d":{"key1":"value1","key2":"value2"}}}}`)
	var r *Request
	assert.NoError(t, json.Unmarshal(b, &r))
	assert.EqualValues(t, &Request{
		Type:      TypeUpdate,
		Ref:       "/path",
		RequestID: 10,
		Data: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	}, r)
}

func TestUnmarshalSetQuery(t *testing.T) {
	b := []byte(`{"t":"d","d":{"r":10,"a":"p","b":{"p":"/path","d":{"key1":"value1","key2":"value2"}}}}`)
	var r *Request
	assert.NoError(t, json.Unmarshal(b, &r))
	assert.EqualValues(t, &Request{
		Type:      TypeSet,
		Ref:       "/path",
		RequestID: 10,
		Data: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	}, r)
}
