package resolver

import "context"

// SchemeLiteral is the scheme under which the literal resolver is registered.
const SchemeLiteral = "literal"

// LiteralResolver resolves inline literal values. The "URI" it receives is the
// literal value itself, which it returns unchanged.
type LiteralResolver struct{}

// Scheme returns SchemeLiteral.
func (LiteralResolver) Scheme() string { return SchemeLiteral }

// Resolve returns uri unchanged: a literal value is already final.
func (LiteralResolver) Resolve(_ context.Context, uri string) (string, error) {
	return uri, nil
}
