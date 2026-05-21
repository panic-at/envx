package resolver_test

import (
	"context"
	"testing"

	"github.com/panic-at/envx/internal/resolver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvResolver_Scheme(t *testing.T) {
	assert.Equal(t, "env", resolver.EnvResolver{}.Scheme())
	assert.Equal(t, "env", resolver.SchemeEnv)
}

func TestParseEnvURI(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		want    string
		wantErr bool
	}{
		{name: "valid", uri: "env://PATH", want: "PATH"},
		{name: "lowercase name", uri: "env://my_var", want: "my_var"},
		{name: "missing prefix", uri: "PATH", wantErr: true},
		{name: "wrong scheme", uri: "op://PATH", wantErr: true},
		{name: "empty name", uri: "env://", wantErr: true},
		{name: "empty string", uri: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolver.ParseEnvURI(tt.uri)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, resolver.ErrInvalidURI)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEnvResolver_Resolve(t *testing.T) {
	r := resolver.EnvResolver{}

	t.Run("set variable", func(t *testing.T) {
		t.Setenv("ENVX_TEST_VALUE", "hello")
		got, err := r.Resolve(context.Background(), "env://ENVX_TEST_VALUE")
		require.NoError(t, err)
		assert.Equal(t, "hello", got)
	})

	t.Run("variable set but empty resolves to empty string", func(t *testing.T) {
		t.Setenv("ENVX_TEST_EMPTY", "")
		got, err := r.Resolve(context.Background(), "env://ENVX_TEST_EMPTY")
		require.NoError(t, err, "an empty value is distinct from an unset variable")
		assert.Equal(t, "", got)
	})

	t.Run("unset variable returns ErrEnvVarNotSet", func(t *testing.T) {
		_, err := r.Resolve(context.Background(), "env://ENVX_TEST_DEFINITELY_UNSET_42")
		require.Error(t, err)
		assert.ErrorIs(t, err, resolver.ErrEnvVarNotSet)
	})

	t.Run("malformed URI", func(t *testing.T) {
		_, err := r.Resolve(context.Background(), "not-a-uri")
		require.Error(t, err)
		assert.ErrorIs(t, err, resolver.ErrInvalidURI)
	})
}
