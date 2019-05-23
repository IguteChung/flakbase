package esprima

// ExpressionStatement defines the AST model for espirma.
type ExpressionStatement struct {
	Type       string      `json:"type"`
	Expression *Expression `json:"expression"`
	Directive  string      `json:"directive"`
}
