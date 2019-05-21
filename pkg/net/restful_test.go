package net

import (
	"net/url"
	"testing"

	"github.com/IguteChung/flakbase/pkg/data"
	"github.com/stretchr/testify/assert"
)

func TestParseInvalidQuery(t *testing.T) {
	values := []url.Values{
		url.Values{"limitToFirst": []string{"3"}, "limitToLast": []string{"2"}},
		url.Values{"limitToFirst": []string{"A"}},
		url.Values{"limitToLast": []string{"A"}},
		url.Values{"equalTo": []string{"value"}, "startAt": []string{"value2"}},
		url.Values{"equalTo": []string{"value"}, "endAt": []string{"value2"}},
	}

	for _, v := range values {
		_, err := ParseQuery(v)
		assert.Error(t, err)
	}
}

func TestParseQuery(t *testing.T) {
	testCases := []struct {
		v url.Values
		q *data.Query
	}{
		{
			url.Values{"limitToFirst": []string{"3"}},
			&data.Query{Limit: 3, LimitOrder: "l"},
		},
		{
			url.Values{"limitToLast": []string{"2"}},
			&data.Query{Limit: 2, LimitOrder: "r"},
		},
		{
			url.Values{"orderBy": []string{"$key"}},
			&data.Query{OrderBy: "$key"},
		},
		{
			url.Values{"orderBy": []string{`"$key"`}},
			&data.Query{OrderBy: "$key"},
		},
		{
			url.Values{"equalTo": []string{"string"}},
			&data.Query{StartAt: "string", EndAt: "string"},
		},
		{
			url.Values{"startAt": []string{"string"}},
			&data.Query{StartAt: "string"},
		},
		{
			url.Values{"startAt": []string{`"string"`}},
			&data.Query{StartAt: "string"},
		},
		{
			url.Values{"startAt": []string{"3"}},
			&data.Query{StartAt: float64(3)},
		},
		{
			url.Values{"endAt": []string{"true"}},
			&data.Query{EndAt: true},
		},
		{
			url.Values{"shallow": []string{"true"}},
			&data.Query{Shallow: true},
		},
		{
			url.Values{"startKey": []string{"string1"}, "endKey": []string{"string2"}},
			&data.Query{StartKey: "string1", EndKey: "string2"},
		},
	}

	for _, tc := range testCases {
		query, err := ParseQuery(tc.v)
		assert.NoError(t, err)
		assert.EqualValues(t, tc.q, query)
	}
}
