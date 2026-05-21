package resolver_test

import (
	"context"
	"testing"

	"github.com/panic-at/envx/internal/resolver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLiteralResolver_Scheme(t *testing.T) {
	assert.Equal(t, "literal", resolver.LiteralResolver{}.Scheme())
	assert.Equal(t, "literal", resolver.SchemeLiteral)
}

func TestLiteralResolver_Resolve(t *testing.T) {
	r := resolver.LiteralResolver{}
	for _, in := range []string{"", "plain", "with spaces", "op://looks/like/uri"} {
		got, err := r.Resolve(context.Background(), in)
		require.NoError(t, err)
		assert.Equal(t, in, got, "literal resolver returns its input unchanged")
	}
}
