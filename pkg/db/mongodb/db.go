package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/IguteChung/flakbase/pkg/db"
	"github.com/IguteChung/flakbase/pkg/rules"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoDB struct {
	*Config
	rules rules.Rules
}

func (m *mongoDB) Connect(ctx context.Context) (db.Client, error) {
	// connect to mongo.
	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(m.URI))
	if err != nil {
		return nil, fmt.Errorf("failed to new mongodb client %s: %v", m.URI, err)
	}
	if err := mongoClient.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb %s: %v", m.URI, err)
	}

	// set database and collection table.
	database, collTable := defaultDatabase, defaultCollTable
	if m.Database != "" {
		database = m.Database
	}
	if m.CollectionsTable != "" {
		collTable = m.CollectionsTable
	}

	return &client{
		Client:    mongoClient,
		rules:     m.rules,
		database:  database,
		collTable: collTable,
	}, nil
}

func (m *mongoDB) SetRules(r rules.Rules) {
	m.rules = r
}

// NewDB creates a mongo DB for Flakbase.
func NewDB(config string) (db.DB, error) {
	b, err := ioutil.ReadFile(config)
	if err != nil {
		return nil, fmt.Errorf("failed to read mongo config %s: %v", config, err)
	}
	var c *Config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mongo config %s: %v", config, err)
	}
	return &mongoDB{Config: c}, nil
}
