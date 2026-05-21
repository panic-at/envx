package resolver

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// SchemeEnv is the scheme under which the host environment resolver is
// registered.
const SchemeEnv = "env"

// EnvResolver resolves values from the host process environment. Its URIs have
// the form env://VAR_NAME.
type EnvResolver struct{}

// Scheme returns SchemeEnv.
func (EnvResolver) Scheme() string { return SchemeEnv }

// Resolve reads the host environment variable named in uri. It returns
// ErrEnvVarNotSet if the variable is not present; a variable that is present
// but empty resolves successfully to the empty string.
func (EnvResolver) Resolve(_ context.Context, uri string) (string, error) {
	name, err := ParseEnvURI(uri)
	if err != nil {
		return "", err
	}
	value, ok := os.LookupEnv(name)
	if !ok {
		return "", fmt.Errorf("%s: %w", uri, ErrEnvVarNotSet)
	}
	return value, nil
}

// ParseEnvURI extracts the variable name from an env://VAR_NAME URI.
func ParseEnvURI(uri string) (string, error) {
	const prefix = SchemeEnv + "://"
	name, ok := strings.CutPrefix(uri, prefix)
	if !ok || name == "" {
		return "", fmt.Errorf("%s: want %sVAR_NAME: %w", uri, prefix, ErrInvalidURI)
	}
	return name, nil
}
