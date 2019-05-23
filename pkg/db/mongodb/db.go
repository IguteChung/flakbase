package mongodb

import (
	"context"
	"fmt"

	"github.com/IguteChung/flakbase/pkg/db"
	"github.com/IguteChung/flakbase/pkg/rules"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoDB struct {
	uri   string
	rules rules.Rules
}

func (m *mongoDB) Connect(ctx context.Context) (db.Client, error) {
	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(m.uri))
	if err != nil {
		return nil, fmt.Errorf("failed to new mongodb client %s: %v", m.uri, err)
	}
	if err := mongoClient.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb %s: %v", m.uri, err)
	}
	return &client{Client: mongoClient, rules: m.rules}, nil
}

func (m *mongoDB) SetRules(r rules.Rules) {
	m.rules = r
}

// NewDB creates a mongo DB for Flakbase.
func NewDB(uri string) db.DB {
	return &mongoDB{uri: uri}
}
