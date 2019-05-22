package mongodb

import (
	"context"
	"fmt"

	"github.com/IguteChung/flakbase/pkg/db"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoDB struct {
	URI string
}

func (m *mongoDB) Connect(ctx context.Context) (db.Client, error) {
	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(m.URI))
	if err != nil {
		return nil, fmt.Errorf("failed to new mongodb client %s: %v", m.URI, err)
	}
	if err := mongoClient.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb %s: %v", m.URI, err)
	}
	return &client{Client: mongoClient}, nil
}

// NewDB creates a mongo DB for Flakbase.
func NewDB(uri string) db.DB {
	return &mongoDB{URI: uri}
}
