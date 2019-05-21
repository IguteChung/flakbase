package net

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/IguteChung/flakbase/pkg/data"
)

// ParseQuery parses query string into client Query.
func ParseQuery(q url.Values) (*data.Query, error) {
	// parse the query strings.
	limitToFirst := q.Get("limitToFirst")
	limitToLast := q.Get("limitToLast")
	orderBy := q.Get("orderBy")
	startAt := q.Get("startAt")
	startKey := q.Get("startKey")
	endAt := q.Get("endAt")
	endKey := q.Get("endKey")
	equalTo := q.Get("equalTo")
	shallow := q.Get("shallow")

	// prepare the request to return.
	query := &data.Query{
		StartKey: startKey,
		EndKey:   endKey,
	}

	// validate the limits.
	if limitToFirst != "" && limitToLast != "" {
		return nil, fmt.Errorf("limitToFirst %s and limitToLast %s both exist", limitToFirst, limitToLast)
	} else if limitToFirst != "" {
		l, err := strconv.ParseInt(limitToFirst, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid limitToFirst %s: %v", limitToFirst, err)
		}
		query.Limit, query.LimitOrder = int(l), "l"
	} else if limitToLast != "" {
		l, err := strconv.ParseInt(limitToLast, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid limitToLast %s: %v", limitToLast, err)
		}
		query.Limit, query.LimitOrder = int(l), "r"
	}

	// validate orderBy, which can be a string w/ or wo/ quotes.
	if orderBy != "" {
		if err := json.Unmarshal([]byte(orderBy), &query.OrderBy); err != nil {
			query.OrderBy = orderBy
		}
	}

	// validate equalTo, startAt and endAt, which can be any types.
	if equalTo != "" {
		if startAt != "" || endAt != "" {
			return nil, fmt.Errorf("equalTo %s cannot be with startAt %s or endAt %s", equalTo, startAt, endAt)
		}
		startAt, endAt = equalTo, equalTo
	}
	if startAt != "" {
		if err := json.Unmarshal([]byte(startAt), &query.StartAt); err != nil {
			query.StartAt = startAt
		}
	}
	if endAt != "" {
		if err := json.Unmarshal([]byte(endAt), &query.EndAt); err != nil {
			query.EndAt = endAt
		}
	}

	// validate shallow query.
	query.Shallow = shallow == "true"

	return query, nil
}
