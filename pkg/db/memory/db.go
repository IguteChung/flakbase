package memory

import (
	"context"
	"sync"

	"github.com/IguteChung/flakbase/pkg/db"
	"github.com/IguteChung/flakbase/pkg/rules"
)

type memory struct {
	sync.RWMutex
	m map[string]interface{}
}

func (m *memory) Connect(ctx context.Context) (db.Client, error) {
	return &client{memory: m}, nil
}

func (m *memory) SetRules(r rules.Rules) {
	// do nothing now.
}

// NewDB creates a memory DB for Flakbase.
func NewDB() db.DB {
	return &memory{
		m: map[string]interface{}{},
	}
}
