package store

import (
	"context"

	"github.com/IguteChung/flakbase/pkg/data"
	"github.com/IguteChung/flakbase/pkg/db/memory"
)

// ListenResult defines the result of handling.
type ListenResult struct {
	NoIndex bool
}

// ListenChannel defines the channel for ListenEvent.
type ListenChannel chan (data.ListenMessage)

// Handler handles the client request.
type Handler interface {
	// HandleSet handles operation set.
	HandleSet(ctx context.Context, ref string, data interface{}) error
	// HandleUpdate handles operation update.
	HandleUpdate(ctx context.Context, ref string, data interface{}) error
	// HandleListen handles the subscription of listen.
	HandleListen(ctx context.Context, ref string, query data.Query, ch ListenChannel) (*ListenResult, error)
	// HandleUnlisten handles the unsubscription of listen.
	HandleUnlisten(ctx context.Context, ref string, query data.Query, ch ListenChannel) error
	// HandleGet handles the operation get.
	HandleGet(ctx context.Context, ref string, query data.Query) (interface{}, error)
	// Reset cleans all data stored, for testing purpose.
	Reset(ctx context.Context) error
}

// NewHandler creates a Handler.
func NewHandler() (Handler, error) {
	return &handler{
		l: &listeners{
			l: map[string]map[ListenChannel]map[data.Query]bool{},
		},
		db: memory.NewDB(),
	}, nil
}
