package resolver_test

import (
	"context"
	"testing"

	"github.com/panic-at/envx/internal/resolver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnePasswordResolver_Scheme(t *testing.T) {
	assert.Equal(t, "op", resolver.OnePasswordResolver{}.Scheme())
	assert.Equal(t, "op", resolver.SchemeOnePassword)
}

func TestParseOPURI(t *testing.T) {
	tests := []struct {
		name               string
		uri                string
		vault, item, field string
		wantErr            bool
	}{
		{name: "valid", uri: "op://Private/GitHub/token", vault: "Private", item: "GitHub", field: "token"},
		{
			name:  "escaped characters decoded",
			uri:   "op://My%20Vault/Data%2FBase/pass%20word",
			vault: "My Vault", item: "Data/Base", field: "pass word",
		},
		{name: "missing field", uri: "op://Private/GitHub", wantErr: true},
		{name: "missing item and field", uri: "op://Private", wantErr: true},
		{name: "too many segments", uri: "op://Private/GitHub/token/extra", wantErr: true},
		{name: "empty vault", uri: "op:///GitHub/token", wantErr: true},
		{name: "empty item", uri: "op://Private//token", wantErr: true},
		{name: "empty field", uri: "op://Private/GitHub/", wantErr: true},
		{name: "wrong scheme", uri: "aws-sm://Private/GitHub/token", wantErr: true},
		{name: "no scheme", uri: "Private/GitHub/token", wantErr: true},
		{name: "invalid escape", uri: "op://Private/GitHub/tok%ZZ", wantErr: true},
		{name: "empty string", uri: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vault, item, field, err := resolver.ParseOPURI(tt.uri)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, resolver.ErrInvalidURI)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.vault, vault)
			assert.Equal(t, tt.item, item)
			assert.Equal(t, tt.field, field)
		})
	}
}

func TestOnePasswordResolver_Resolve(t *testing.T) {
	r := resolver.OnePasswordResolver{}

	t.Run("valid URI reports not implemented", func(t *testing.T) {
		_, err := r.Resolve(context.Background(), "op://Private/GitHub/token")
		require.Error(t, err)
		assert.ErrorIs(t, err, resolver.ErrNotImplemented)
		assert.Contains(t, err.Error(), "op://Private/GitHub/token")
	})

	t.Run("invalid URI reports invalid", func(t *testing.T) {
		_, err := r.Resolve(context.Background(), "op://Private")
		require.Error(t, err)
		assert.ErrorIs(t, err, resolver.ErrInvalidURI)
	})
}
