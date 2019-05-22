package memory

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/IguteChung/flakbase/pkg/data"
	"github.com/mohae/deepcopy"
)

type client struct {
	*memory
}

func (c *client) Close() error {
	return nil
}

func (c *client) Set(ctx context.Context, ref string, data interface{}) error {
	// lock the whole db for write.
	c.Lock()
	defer c.Unlock()

	// for each segment of path, append the data to the data tree.
	m := c.m
	paths := strings.Split(ref, "/")
	for i, p := range paths {
		if p == "" {
			// leading space.
			continue
		}

		// trailing branch.
		if i == len(paths)-1 {
			if data == nil {
				// for delete case.
				delete(m, p)
			} else {
				m[p] = deepcopy.Copy(data)
			}
			break
		}

		// handle if branch not exists.
		if _, ok := m[p].(map[string]interface{}); !ok {
			m[p] = map[string]interface{}{}
		}

		// move the pointer to child.
		m = m[p].(map[string]interface{})
	}

	return nil
}

func (c *client) Get(ctx context.Context, ref string, query data.Query) (interface{}, error) {
	// lock the db for read.
	c.RLock()
	defer c.RUnlock()

	// handle query on root.
	if ref == "/" {
		return queryOnData(c.m, query), nil
	}

	m := c.m
	paths := strings.Split(ref, "/")
	for i, p := range paths {
		if p == "" {
			// leading space.
			continue
		}

		// trailing branch.
		if i == len(paths)-1 {
			return queryOnData(m[p], query), nil
		}

		// move the pointer to child.
		if child, ok := m[p].(map[string]interface{}); ok {
			m = child
		} else {
			return nil, nil
		}
	}

	return nil, nil
}

func queryOnData(data interface{}, query data.Query) interface{} {
	// check if data is filterable.
	m, ok := data.(map[string]interface{})
	if !ok {
		return data
	}

	type node struct {
		key   string
		value interface{}
		index interface{}
	}
	updated := make([]*node, 0, len(m))

	// for each node in map, find the key index for ordering.
	for k, v := range m {
		var index interface{}
		switch query.OrderBy {
		case ".key", "$key":
			index = k
		case ".value", "$value":
			index = v
		case "":
			// do nothing
		default:
			// support nested query.
			ptr := v
			orderBys := strings.Split(query.OrderBy, ".")
			for i, orderBy := range orderBys {
				if orderBy == "" {
					// leading space.
					continue
				}

				if ptrMap, ok := ptr.(map[string]interface{}); ok {
					if i == len(orderBys)-1 {
						// pick the index from child.
						index = ptrMap[orderBy]
						break
					}
					// move ptr to child.
					ptr = ptrMap[orderBy]
				}
			}
		}

		// filter by startAt and endAt.
		// TODO: currently convert interface to string and do the comparison.
		if query.StartAt != nil && fmt.Sprint(index) < fmt.Sprint(query.StartAt) {
			continue
		}
		if query.EndAt != nil && fmt.Sprint(index) > fmt.Sprint(query.EndAt) {
			continue
		}

		// filter by startKey and endKey.
		if query.StartKey != "" && k < query.StartKey {
			continue
		}
		if query.EndKey != "" && k > query.EndKey {
			continue
		}

		updated = append(updated, &node{key: k, value: v, index: index})
	}

	// sort all nodes and filter by limit.
	if limit := query.Limit; limit != 0 {
		sort.Slice(updated, func(i int, j int) bool {
			if updated[i].index == updated[j].index {
				// if index equals, compare key.
				return updated[i].key < updated[j].key
			}
			// TODO: currently convert interface to string and do the comparison.
			return fmt.Sprint(updated[i].index) < fmt.Sprint(updated[j].index)
		})
		if limit > len(updated) {
			limit = len(updated)
		}
		if query.LimitOrder == "l" {
			// limitToFirst case.
			updated = updated[:limit]
		} else {
			// limitToLast case.
			updated = updated[len(updated)-limit:]
		}
	}

	// generate response map.
	updatedMap := map[string]interface{}{}
	hasPrimary := false
	for _, u := range updated {
		if _, ok := u.value.(map[string]interface{}); !ok {
			hasPrimary = true
		}
		updatedMap[u.key] = u.value
	}

	// handle shallow query is has no primary in map.
	if query.Shallow && !hasPrimary {
		for k := range updatedMap {
			updatedMap[k] = true
		}
	}

	return deepcopy.Copy(updatedMap)
}

func (c *client) Reset(ctx context.Context) error {
	c.Lock()
	defer c.Unlock()

	c.m = map[string]interface{}{}
	return nil
}
