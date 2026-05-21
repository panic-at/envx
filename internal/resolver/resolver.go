// Package resolver turns a profile.Var into its final string value.
//
// Every value source — inline literals, host environment variables, 1Password,
// AWS Secrets Manager — implements the Resolver interface and is registered in
// a Registry under the URI scheme it handles. ResolveAll fans out across an
// effective profile, resolving every variable concurrently.
package resolver

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/panic-at/envx/internal/profile"
)

// Sentinel errors returned, wrapped, by resolvers and the registry. Callers
// classify failures with errors.Is.
var (
	ErrNotImplemented = errors.New("resolver: not implemented")
	ErrUnknownScheme  = errors.New("resolver: unknown scheme")
	ErrInvalidURI     = errors.New("resolver: invalid URI")
	ErrEnvVarNotSet   = errors.New("resolver: environment variable not set")
)

// maxConcurrency bounds the number of variables resolved in parallel.
const maxConcurrency = 8

// Resolver resolves a single value, either a literal or a reference URI.
type Resolver interface {
	// Scheme returns the URI scheme this resolver handles
	// (e.g. "literal", "env", "op", "aws-sm").
	Scheme() string
	// Resolve returns the final string value for the given URI. For the
	// literal resolver the URI is the literal value itself.
	Resolve(ctx context.Context, uri string) (string, error)
}

// Registry maps URI schemes to the Resolver that handles them. It is safe for
// concurrent use by multiple goroutines.
type Registry struct {
	mu        sync.RWMutex
	resolvers map[string]Resolver
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{resolvers: make(map[string]Resolver)}
}

// Register adds res to the registry under res.Scheme(). It returns an error if
// a resolver is already registered for that scheme.
func (r *Registry) Register(res Resolver) error {
	scheme := res.Scheme()
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.resolvers[scheme]; exists {
		return fmt.Errorf("resolver: scheme %q already registered", scheme)
	}
	r.resolvers[scheme] = res
	return nil
}

// Get returns the resolver registered for scheme and reports whether one was
// found.
func (r *Registry) Get(scheme string) (Resolver, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res, ok := r.resolvers[scheme]
	return res, ok
}

// DefaultRegistry returns a Registry pre-populated with the built-in
// resolvers: literal, env, op and aws-sm.
func DefaultRegistry() *Registry {
	r := NewRegistry()
	for _, res := range []Resolver{
		LiteralResolver{},
		EnvResolver{},
		OnePasswordResolver{},
		AWSSMResolver{},
	} {
		// Cannot fail: the built-in schemes are distinct constants.
		_ = r.Register(res)
	}
	return r
}

// ResolveResult holds the outcome of ResolveAll. A variable key appears in
// exactly one of the two maps: Values on success, Errors on failure.
type ResolveResult struct {
	Values map[string]string // KEY -> resolved value
	Errors map[string]error  // KEY -> error, if resolution failed
}

// ResolveAll resolves every variable of p concurrently, using reg to look up a
// resolver per variable.
//
// Resolution is fail-soft: an error on one variable is recorded in the result
// and never aborts the others. Concurrency is capped at maxConcurrency
// goroutines. Cancelling ctx stops pending and in-flight resolutions, whose
// keys are recorded in Errors. Variables are dispatched in sorted key order so
// that logs and traces are deterministic.
func ResolveAll(ctx context.Context, reg *Registry, p profile.Profile) ResolveResult {
	res := ResolveResult{
		Values: make(map[string]string, len(p.Vars)),
		Errors: make(map[string]error),
	}
	var (
		mu sync.Mutex
		g  errgroup.Group
	)
	g.SetLimit(maxConcurrency)
	for _, key := range sortedKeys(p) {
		key := key
		v := p.Vars[key]
		g.Go(func() error {
			value, err := resolveVar(ctx, reg, v)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				res.Errors[key] = err
			} else {
				res.Values[key] = value
			}
			return nil
		})
	}
	// Wait never returns an error: every goroutine returns nil so that one
	// failure cannot cancel the group.
	_ = g.Wait()
	return res
}

// sortedKeys returns the keys of p.Vars in ascending lexical order.
func sortedKeys(p profile.Profile) []string {
	keys := make([]string, 0, len(p.Vars))
	for k := range p.Vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// resolveVar resolves a single variable, honouring ctx cancellation before any
// work begins so that goroutines queued behind the concurrency limit exit
// promptly once ctx is done.
func resolveVar(ctx context.Context, reg *Registry, v profile.Var) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	scheme := SchemeLiteral
	uri := v.Value
	if v.Type == profile.VarRef {
		s, err := schemeOf(v.URI)
		if err != nil {
			return "", err
		}
		scheme, uri = s, v.URI
	}
	r, ok := reg.Get(scheme)
	if !ok {
		return "", fmt.Errorf("%q: %w", scheme, ErrUnknownScheme)
	}
	return r.Resolve(ctx, uri)
}

// schemeOf extracts the scheme component (the text before "://") of a
// reference URI.
func schemeOf(uri string) (string, error) {
	i := strings.Index(uri, "://")
	if i <= 0 {
		return "", fmt.Errorf("%q: missing scheme: %w", uri, ErrInvalidURI)
	}
	return uri[:i], nil
}
