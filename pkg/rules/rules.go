package rules

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

// Rules defines the security rules model.
type Rules map[string]interface{}

// Import reads the security rule file and returns a parsed Rules.
func Import(filename string) (Rules, error) {
	// if no rule file specified.
	if filename == "" {
		return nil, nil
	}

	// read the rule file.
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read rule file %s: %v", filename, err)
	}

	// unmarshal to Rules.
	var r Rules
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rule: %v", err)
	}

	// cd to rules.
	return r.child("rules"), nil
}

// ContainsKey checks whether current Rules contains a specific key.
func (r Rules) ContainsKey(key string) bool {
	_, ok := r[key]
	return ok
}

// VariableKey return the variable key of current Rules, empty if no variable key found.
func (r Rules) VariableKey() string {
	for k := range r {
		if k != "$other" && strings.HasPrefix(k, "$") {
			return k
		}
	}
	return ""
}

// Child changes the rule path with given path.
func (r Rules) Child(path string) Rules {
	for _, p := range strings.Split(path, "/") {
		if p == "" {
			continue
		}
		r = r.child(p)
	}
	return r
}

// Indexes retrieves the indexes for current Rules.
func (r Rules) Indexes() []string {
	i, ok := r[".indexOn"]
	if !ok {
		return nil
	}
	indexes, ok := i.([]string)
	if !ok {
		return nil
	}
	return indexes
}

func (r Rules) child(name string) Rules {
	// change to variable key or child name.
	var child interface{}
	if v, ok := r[name]; ok {
		child = v
	} else if k := r.VariableKey(); k != "" {
		child = r[k]
	}

	// convert the child to a Rule.
	m, ok := child.(map[string]interface{})
	if !ok {
		return nil
	}
	return Rules(m)
}
