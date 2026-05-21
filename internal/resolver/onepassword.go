package resolver

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// SchemeOnePassword is the scheme under which the 1Password resolver is
// registered.
const SchemeOnePassword = "op"

// OnePasswordResolver resolves secrets from 1Password. Its URIs have the form
// op://VAULT/ITEM/FIELD. URI parsing is fully implemented; the live lookup is
// not yet wired up, so Resolve reports ErrNotImplemented.
type OnePasswordResolver struct{}

// Scheme returns SchemeOnePassword.
func (OnePasswordResolver) Scheme() string { return SchemeOnePassword }

// Resolve parses uri and reports ErrNotImplemented. A parse error takes
// precedence, so callers still learn about malformed URIs; the error message
// echoes the parsed components to confirm parsing succeeded.
func (OnePasswordResolver) Resolve(_ context.Context, uri string) (string, error) {
	vault, item, field, err := ParseOPURI(uri)
	if err != nil {
		return "", err
	}
	return "", fmt.Errorf("op://%s/%s/%s: %w", vault, item, field, ErrNotImplemented)
}

// ParseOPURI splits a 1Password reference of the form op://VAULT/ITEM/FIELD
// into its three components. Percent-escaped characters in any segment are
// decoded. All three segments are mandatory and must be non-empty.
func ParseOPURI(uri string) (vault, item, field string, err error) {
	const prefix = SchemeOnePassword + "://"
	rest, ok := strings.CutPrefix(uri, prefix)
	if !ok {
		return "", "", "", fmt.Errorf("%s: want %sVAULT/ITEM/FIELD: %w", uri, prefix, ErrInvalidURI)
	}
	parts := strings.Split(rest, "/")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("%s: want VAULT/ITEM/FIELD (3 segments): %w", uri, ErrInvalidURI)
	}
	vault, verr := url.PathUnescape(parts[0])
	item, ierr := url.PathUnescape(parts[1])
	field, ferr := url.PathUnescape(parts[2])
	if verr != nil || ierr != nil || ferr != nil {
		return "", "", "", fmt.Errorf("%s: invalid escape sequence: %w", uri, ErrInvalidURI)
	}
	if vault == "" || item == "" || field == "" {
		return "", "", "", fmt.Errorf("%s: empty segment: %w", uri, ErrInvalidURI)
	}
	return vault, item, field, nil
}
