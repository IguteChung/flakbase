package store

import (
	"testing"

	"github.com/IguteChung/flakbase/pkg/data"
	"github.com/stretchr/testify/assert"
)

func assertContainsOnly(t *testing.T, s interface{}, contains ...interface{}) bool {
	if !assert.Len(t, s, len(contains)) {
		return false
	}
	for _, contain := range contains {
		if !assert.Contains(t, s, contain) {
			return false
		}
	}
	return true
}

func TestFindReferences(t *testing.T) {
	l := &listeners{
		l: map[string]map[ListenChannel]map[data.Query]bool{
			"/":                                 nil,
			"/path":                             nil,
			"/path/collection/document1":        nil,
			"/path/collection/document2":        nil,
			"/path/collection2/document1":       nil,
			"/path/collection2/document1/field": nil,
		},
	}

	assertContainsOnly(t, l.find("/path2"), "/")
	assertContainsOnly(t, l.find("/path/collection3"), "/", "/path")
	assertContainsOnly(t, l.find("/path"),
		"/",
		"/path",
		"/path/collection/document1",
		"/path/collection/document2",
		"/path/collection2/document1",
		"/path/collection2/document1/field",
	)
	assertContainsOnly(t, l.find("/path/collection"),
		"/",
		"/path",
		"/path/collection/document1",
		"/path/collection/document2",
	)
	assertContainsOnly(t, l.find("/path/collection/document1"),
		"/",
		"/path",
		"/path/collection/document1",
	)
	assertContainsOnly(t, l.find("/path/collection2/document1"),
		"/",
		"/path",
		"/path/collection2/document1",
		"/path/collection2/document1/field",
	)
	assertContainsOnly(t, l.find("/path/collection2/document1/field"),
		"/",
		"/path",
		"/path/collection2/document1",
		"/path/collection2/document1/field",
	)
	assertContainsOnly(t, l.find("/path/collection/document1", "/path/collection2/document1"),
		"/",
		"/path",
		"/path/collection/document1",
		"/path/collection2/document1",
		"/path/collection2/document1/field",
	)
}
