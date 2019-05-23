package esprima

// Program defines the AST model for espirma.
type Program struct {
	Type       string                 `json:"type"`
	Body       []*ExpressionStatement `json:"body"`
	SourceType string                 `json:"sourceType"`
}
