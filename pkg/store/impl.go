package store

import (
	"context"
	"fmt"
	"path"

	"github.com/IguteChung/flakbase/pkg/data"
	"github.com/IguteChung/flakbase/pkg/db"
)

type handler struct {
	l  *listeners
	db db.DB
}

func (s *handler) HandleSet(ctx context.Context, ref string, data interface{}) error {
	// connect to db.
	client, err := s.db.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %v", err)
	}
	defer client.Close()

	// set the data to DB.
	if err := client.Set(ctx, ref, data); err != nil {
		return fmt.Errorf("failed to set data to %s: %v", ref, err)
	}

	// callback the data.
	if err := s.callbackRef(ctx, client, ref); err != nil {
		return fmt.Errorf("failed to callback set %s: %v", ref, err)
	}
	return nil
}

func (s *handler) HandleUpdate(ctx context.Context, ref string, data interface{}) error {
	// connect to db.
	client, err := s.db.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %v", err)
	}
	defer client.Close()

	// if the data is a map, set the data sequentially.
	changedRefs := []string{}
	if m, ok := data.(map[string]interface{}); ok {
		// TODO: set entries in transaction.
		for k, v := range m {
			updatedRef := path.Join(ref, k)
			changedRefs = append(changedRefs, updatedRef)
			if err := client.Set(ctx, updatedRef, v); err != nil {
				return fmt.Errorf("failed to update data to %s: %v", ref, err)
			}
		}
	} else {
		// directly call set if data is not a map.
		changedRefs = []string{ref}
		if err := client.Set(ctx, ref, data); err != nil {
			return fmt.Errorf("failed to set data to %s: %v", ref, err)
		}
	}

	// callback the data.
	if err := s.callbackRef(ctx, client, changedRefs...); err != nil {
		return fmt.Errorf("failed to callback set %s: %v", ref, err)
	}

	return nil
}

func (s *handler) HandleListen(ctx context.Context, ref string, query data.Query, ch ListenChannel) (*ListenResult, error) {
	// register the listener.
	s.l.register(ref, ch, query)

	// connect to db.
	client, err := s.db.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %v", err)
	}
	defer client.Close()

	// read data once and callback.
	resp, err := client.Get(ctx, ref, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s: %v", ref, err)
	}
	ch <- data.ListenMessage{
		Ref:     ref,
		QueryID: query.ID,
		Data:    resp,
	}
	return &ListenResult{}, nil
}

func (s *handler) HandleUnlisten(ctx context.Context, ref string, query data.Query, ch ListenChannel) error {
	s.l.unregister(ref, ch, query)
	return nil
}

func (s *handler) HandleGet(ctx context.Context, ref string, query data.Query) (interface{}, error) {
	// connect to db.
	client, err := s.db.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %v", err)
	}
	defer client.Close()

	// get the data from DB.
	resp, err := client.Get(ctx, ref, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get data from %s: %v", ref, err)
	}
	return resp, nil
}

func (s *handler) Reset(ctx context.Context) error {
	// clean the listener.
	s.l.clean()

	// connect to db.
	client, err := s.db.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %v", err)
	}
	defer client.Close()

	// reset the db.
	return client.Reset(ctx)
}

func (s *handler) callbackRef(ctx context.Context, client db.Client, updatedRefs ...string) error {
	for _, ref := range s.l.find(updatedRefs...) {
		for ch, queries := range s.l.l[ref] {
			for query := range queries {
				// TODO: consider load changed data in parallel.
				resp, err := client.Get(ctx, ref, query)
				if err != nil {
					return fmt.Errorf("failed to callback %s: %v", ref, err)
				}
				ch <- data.ListenMessage{
					Ref:     ref,
					QueryID: query.ID,
					Data:    resp,
				}
			}
		}
	}
	return nil
}
