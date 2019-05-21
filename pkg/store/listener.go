package store

import (
	"strings"
	"sync"

	"github.com/IguteChung/flakbase/pkg/data"
)

type listeners struct {
	sync.Mutex
	l map[string]map[ListenChannel]map[data.Query]bool
}

func (l *listeners) register(ref string, ch ListenChannel, query data.Query) {
	l.Lock()
	defer l.Unlock()

	if _, ok := l.l[ref]; !ok {
		l.l[ref] = map[ListenChannel]map[data.Query]bool{}
	}
	if _, ok := l.l[ref][ch]; !ok {
		l.l[ref][ch] = map[data.Query]bool{}
	}

	l.l[ref][ch][query] = true
}

func (l *listeners) unregister(ref string, ch ListenChannel, query data.Query) {
	l.Lock()
	defer l.Unlock()

	if chs, ok := l.l[ref]; ok {
		if queries, ok := chs[ch]; ok {
			delete(queries, query)
		}
	}
}

func (l *listeners) clean() {
	l.Lock()
	defer l.Unlock()

	l.l = map[string]map[ListenChannel]map[data.Query]bool{}
}

// find matches the listeners and returns matched references.
func (l *listeners) find(updatedRefs ...string) []string {
	refs := make([]string, 0, len(l.l))
	for ref := range l.l {
		for _, updatedRef := range updatedRefs {
			refPath, updatedRefPath := ref+"/", updatedRef+"/"
			if ref == "/" || strings.HasPrefix(refPath, updatedRefPath) || strings.HasPrefix(updatedRefPath, refPath) {
				refs = append(refs, ref)
				break
			}
		}
	}
	return refs
}
