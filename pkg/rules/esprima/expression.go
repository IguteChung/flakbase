package esprima

import (
	"encoding/json"
	"fmt"
)

// Expression defines the AST model for espirma.
type Expression struct {
	Type              string
	BinaryExpression  *BinaryExpression
	CallExpression    *CallExpression
	MemberExpression  *MemberExpression
	LogicalExpression *LogicalExpression
	ArrayExpression   *ArrayExpression
	Identifier        *Identifier
	Literal           *Literal
}

// BinaryExpression defines the AST model for espirma.
type BinaryExpression struct {
	Operator string      `json:"operator"`
	Left     *Expression `json:"left"`
	Right    *Expression `json:"right"`
}

// CallExpression defines the AST model for espirma.
type CallExpression struct {
	Callee    *Expression   `json:"callee"`
	Arguments []*Expression `json:"arguments"`
}

// MemberExpression defines the AST model for espirma.
type MemberExpression struct {
	Computed bool        `json:"computed"`
	Object   *Expression `json:"object"`
	Property *Expression `json:"property"`
}

// LogicalExpression defines the AST model for espirma.
type LogicalExpression struct {
	Operator string      `json:"operator"`
	Left     *Expression `json:"left"`
	Right    *Expression `json:"right"`
}

// ArrayExpression defines the AST model for espirma.
type ArrayExpression struct {
	Elements []*Expression `json:"elements"`
}

// Identifier defines the AST model for espirma.
type Identifier struct {
	Name string `json:"name"`
}

// Literal defines the AST model for espirma.
type Literal struct {
	Value interface{} `json:"value"`
	Raw   string      `json:"raw"`
	Regex *Regex      `json:"regex"`
}

// Regex defines the AST model for espirma.
type Regex struct {
	Pattern string `json:"pattern"`
	Flags   string `json:"flags"`
}

// UnmarshalJSON overrides Expression's json unmarshal.
func (e *Expression) UnmarshalJSON(bytes []byte) (err error) {
	var n *node
	if err = json.Unmarshal(bytes, &n); err != nil {
		return
	}

	switch n.Type {
	case "BinaryExpression":
		err = json.Unmarshal(bytes, &e.BinaryExpression)
	case "CallExpression":
		err = json.Unmarshal(bytes, &e.CallExpression)
	case "MemberExpression":
		err = json.Unmarshal(bytes, &e.MemberExpression)
	case "LogicalExpression":
		err = json.Unmarshal(bytes, &e.LogicalExpression)
	case "ArrayExpression":
		err = json.Unmarshal(bytes, &e.ArrayExpression)
	case "Identifier":
		err = json.Unmarshal(bytes, &e.Identifier)
	case "Literal":
		err = json.Unmarshal(bytes, &e.Literal)
	default:
		err = fmt.Errorf("invalid expression type: %s", e.Type)
	}
	if err != nil {
		return fmt.Errorf("failed to unmarshal expression %s: %v", e.Type, err)
	}
	e.Type = n.Type

	return
}
