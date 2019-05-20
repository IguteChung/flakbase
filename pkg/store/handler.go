package store

import "github.com/IguteChung/flakbase/pkg/data"

// Handler handles client request and process the data.
type Handler struct {
}

func (s *Handler) Handle(q *data.Query) error {
	return nil
}
