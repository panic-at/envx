package resolver

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// SchemeAWSSM is the scheme under which the AWS Secrets Manager resolver is
// registered.
const SchemeAWSSM = "aws-sm"

// AWSSMResolver resolves secrets from AWS Secrets Manager. Its URIs have the
// form aws-sm://REGION/SECRET_NAME, optionally with a ?key=JSON_KEY query to
// select a field of a JSON-encoded secret. URI parsing is fully implemented;
// the live lookup is not yet wired up, so Resolve reports ErrNotImplemented.
type AWSSMResolver struct{}

// Scheme returns SchemeAWSSM.
func (AWSSMResolver) Scheme() string { return SchemeAWSSM }

// Resolve parses uri and reports ErrNotImplemented. A parse error takes
// precedence, so callers still learn about malformed URIs; the error message
// echoes the parsed components to confirm parsing succeeded.
func (AWSSMResolver) Resolve(_ context.Context, uri string) (string, error) {
	region, name, jsonKey, err := ParseAWSSMURI(uri)
	if err != nil {
		return "", err
	}
	if jsonKey != "" {
		return "", fmt.Errorf("aws-sm://%s/%s?key=%s: %w", region, name, jsonKey, ErrNotImplemented)
	}
	return "", fmt.Errorf("aws-sm://%s/%s: %w", region, name, ErrNotImplemented)
}

// ParseAWSSMURI splits an AWS Secrets Manager reference of the form
// aws-sm://REGION/SECRET_NAME or aws-sm://REGION/SECRET_NAME?key=JSON_KEY.
//
// The secret name may itself contain slashes. Percent-escaped characters in
// the region and name are decoded. jsonKey is empty when no key query
// parameter is present.
func ParseAWSSMURI(uri string) (region, name, jsonKey string, err error) {
	const prefix = SchemeAWSSM + "://"
	rest, ok := strings.CutPrefix(uri, prefix)
	if !ok {
		return "", "", "", fmt.Errorf("%s: want %sREGION/SECRET_NAME: %w", uri, prefix, ErrInvalidURI)
	}
	if i := strings.IndexByte(rest, '?'); i >= 0 {
		query, qerr := url.ParseQuery(rest[i+1:])
		if qerr != nil {
			return "", "", "", fmt.Errorf("%s: invalid query: %w", uri, ErrInvalidURI)
		}
		jsonKey = query.Get("key")
		rest = rest[:i]
	}
	rawRegion, rawName, found := strings.Cut(rest, "/")
	if !found {
		return "", "", "", fmt.Errorf("%s: want REGION/SECRET_NAME: %w", uri, ErrInvalidURI)
	}
	region, rerr := url.PathUnescape(rawRegion)
	name, nerr := url.PathUnescape(rawName)
	if rerr != nil || nerr != nil {
		return "", "", "", fmt.Errorf("%s: invalid escape sequence: %w", uri, ErrInvalidURI)
	}
	if region == "" || name == "" {
		return "", "", "", fmt.Errorf("%s: empty segment: %w", uri, ErrInvalidURI)
	}
	return region, name, jsonKey, nil
}
