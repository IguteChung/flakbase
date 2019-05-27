package store

import (
	"context"
	"fmt"

	"github.com/IguteChung/flakbase/pkg/data"
	"github.com/IguteChung/flakbase/pkg/db"
	"github.com/IguteChung/flakbase/pkg/db/memory"
	"github.com/IguteChung/flakbase/pkg/db/mongodb"
	"github.com/IguteChung/flakbase/pkg/rules"
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

// Config defines the config to create handler.
type Config struct {
	Mongo string
	Rule  string
}

// NewHandler creates a Handler.
func NewHandler(c *Config) (Handler, error) {
	// decide the db to use by config.
	var db db.DB
	if c.Mongo != "" {
		mongo, err := mongodb.NewDB(c.Mongo)
		if err != nil {
			return nil, fmt.Errorf("failed to create mongo db %s: %v", c.Mongo, err)
		}
		db = mongo
	} else {
		db = memory.NewDB()
	}

	// load security rules if specified.
	r, err := rules.Import(c.Rule)
	if err != nil {
		return nil, fmt.Errorf("failed to import security rule %s: %v", c.Rule, err)
	}

	// set the db rules.
	db.SetRules(r.Child("rules"))

	return &handler{
		l: &listeners{
			l: map[string]map[ListenChannel]map[data.Query]bool{},
		},
		db: db,
	}, nil
}
