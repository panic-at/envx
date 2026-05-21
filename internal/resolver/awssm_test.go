package resolver_test

import (
	"context"
	"testing"

	"github.com/panic-at/envx/internal/resolver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAWSSMResolver_Scheme(t *testing.T) {
	assert.Equal(t, "aws-sm", resolver.AWSSMResolver{}.Scheme())
	assert.Equal(t, "aws-sm", resolver.SchemeAWSSM)
}

func TestParseAWSSMURI(t *testing.T) {
	tests := []struct {
		name                    string
		uri                     string
		region, secret, jsonKey string
		wantErr                 bool
	}{
		{name: "valid", uri: "aws-sm://us-east-1/prod-db", region: "us-east-1", secret: "prod-db"},
		{
			name: "with json key", uri: "aws-sm://eu-west-1/app/secret?key=password",
			region: "eu-west-1", secret: "app/secret", jsonKey: "password",
		},
		{
			name: "secret name with slashes", uri: "aws-sm://us-east-1/prod/db/password",
			region: "us-east-1", secret: "prod/db/password",
		},
		{
			name: "escaped characters decoded", uri: "aws-sm://us-east-1/my%20secret?key=json%2Fkey",
			region: "us-east-1", secret: "my secret", jsonKey: "json/key",
		},
		{name: "missing secret name", uri: "aws-sm://us-east-1", wantErr: true},
		{name: "empty region", uri: "aws-sm:///prod-db", wantErr: true},
		{name: "empty secret", uri: "aws-sm://us-east-1/", wantErr: true},
		{name: "wrong scheme", uri: "op://us-east-1/prod-db", wantErr: true},
		{name: "no scheme", uri: "us-east-1/prod-db", wantErr: true},
		{name: "invalid escape", uri: "aws-sm://us-east-1/sec%ZZ", wantErr: true},
		{name: "empty string", uri: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			region, secret, jsonKey, err := resolver.ParseAWSSMURI(tt.uri)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, resolver.ErrInvalidURI)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.region, region)
			assert.Equal(t, tt.secret, secret)
			assert.Equal(t, tt.jsonKey, jsonKey)
		})
	}
}

func TestAWSSMResolver_Resolve(t *testing.T) {
	r := resolver.AWSSMResolver{}

	t.Run("valid URI reports not implemented", func(t *testing.T) {
		_, err := r.Resolve(context.Background(), "aws-sm://us-east-1/prod-db")
		require.Error(t, err)
		assert.ErrorIs(t, err, resolver.ErrNotImplemented)
		assert.Contains(t, err.Error(), "aws-sm://us-east-1/prod-db")
	})

	t.Run("valid URI with json key reports not implemented", func(t *testing.T) {
		_, err := r.Resolve(context.Background(), "aws-sm://us-east-1/prod-db?key=user")
		require.Error(t, err)
		assert.ErrorIs(t, err, resolver.ErrNotImplemented)
		assert.Contains(t, err.Error(), "key=user")
	})

	t.Run("invalid URI reports invalid", func(t *testing.T) {
		_, err := r.Resolve(context.Background(), "aws-sm://us-east-1")
		require.Error(t, err)
		assert.ErrorIs(t, err, resolver.ErrInvalidURI)
	})
}
