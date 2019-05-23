package esprima

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

// node defines the basic esprima node, used internally for unmarshalling.
type node struct {
	Type string `json:"type"`
}

// Parse parses a js string and return the esprima AST model.
func Parse(content string) (*Program, error) {
	// create a temp js file.
	tmpfile, err := ioutil.TempFile("/tmp", "esprima.*.js")
	if err != nil {
		return nil, fmt.Errorf("failed to create tmp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// write the content to temp file.
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		return nil, fmt.Errorf("failed to write the tmp file: %v", err)
	}
	tmpfile.Close()

	// run esprima parser.
	bytes, err := exec.Command("esparse", tmpfile.Name()).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to esparse %s: %v", content, err)
	}
	log.Println(string(bytes))

	// unmarshal the parsed content to a esprima node.
	var m *Program
	if err := json.Unmarshal(bytes, &m); err != nil {
		return nil, fmt.Errorf("failed to unmarshal esprima %s: %v", string(bytes), err)
	}
	return m, nil
}
