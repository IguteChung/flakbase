package db

import (
	"context"
	"io"

	"github.com/IguteChung/flakbase/pkg/data"
	"github.com/IguteChung/flakbase/pkg/rules"
)

// DB defines the database interface for Flakbase.
type DB interface {
	// Connect prepares the DB client.
	Connect(ctx context.Context) (Client, error)
	// SetRules sets the security rules of DB.
	SetRules(r rules.Rules)
}

// Client defines the database client to set/get data.
type Client interface {
	io.Closer
	// Set inserts or updates the data to given reference.
	Set(ctx context.Context, ref string, data interface{}) error
	// Get retrieves the data from reference by given query.
	Get(ctx context.Context, ref string, query data.Query) (interface{}, error)
	// Reset cleans all data stored, for testing purpose.
	Reset(ctx context.Context) error
}
