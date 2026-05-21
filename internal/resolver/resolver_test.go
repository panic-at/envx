package resolver_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/panic-at/envx/internal/profile"
	"github.com/panic-at/envx/internal/resolver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeResolver is a configurable Resolver used to exercise Registry and
// ResolveAll without touching real value sources. A non-zero delay makes
// Resolve block, observing ctx cancellation while it waits.
type fakeResolver struct {
	scheme string
	value  string
	err    error
	delay  time.Duration
}

func (f fakeResolver) Scheme() string { return f.scheme }

func (f fakeResolver) Resolve(ctx context.Context, _ string) (string, error) {
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	if f.err != nil {
		return "", f.err
	}
	return f.value, nil
}

func ref(uri string) profile.Var { return profile.Var{Type: profile.VarRef, URI: uri} }
func lit(v string) profile.Var   { return profile.Var{Type: profile.VarLiteral, Value: v} }

func TestRegistry(t *testing.T) {
	r := resolver.NewRegistry()
	fake := fakeResolver{scheme: "fake", value: "v"}

	require.NoError(t, r.Register(fake))

	got, ok := r.Get("fake")
	require.True(t, ok)
	assert.Equal(t, fake, got)

	_, ok = r.Get("missing")
	assert.False(t, ok)

	err := r.Register(fakeResolver{scheme: "fake"})
	require.Error(t, err, "registering a duplicate scheme must fail")
	assert.Contains(t, err.Error(), "fake")
}

func TestDefaultRegistry(t *testing.T) {
	r := resolver.DefaultRegistry()
	for _, scheme := range []string{
		resolver.SchemeLiteral,
		resolver.SchemeEnv,
		resolver.SchemeOnePassword,
		resolver.SchemeAWSSM,
	} {
		res, ok := r.Get(scheme)
		require.Truef(t, ok, "default registry must provide scheme %q", scheme)
		assert.Equal(t, scheme, res.Scheme())
	}
	_, ok := r.Get("does-not-exist")
	assert.False(t, ok)
}

// TestRegistryConcurrent registers and reads from a single registry from many
// goroutines. Run under -race it proves the RWMutex guards every access.
func TestRegistryConcurrent(t *testing.T) {
	r := resolver.NewRegistry()
	const n = 64

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			scheme := fmt.Sprintf("scheme-%d", i)
			require.NoError(t, r.Register(fakeResolver{scheme: scheme}))
			// Concurrent reads, including of schemes other goroutines own.
			r.Get(fmt.Sprintf("scheme-%d", (i+1)%n))
		}(i)
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		_, ok := r.Get(fmt.Sprintf("scheme-%d", i))
		assert.True(t, ok)
	}
}

func TestResolveAll_Literals(t *testing.T) {
	reg := resolver.DefaultRegistry()
	p := profile.Profile{Vars: map[string]profile.Var{
		"A": lit("alpha"),
		"B": lit("beta"),
		"C": lit(""),
	}}

	res := resolver.ResolveAll(context.Background(), reg, p)

	assert.Empty(t, res.Errors)
	assert.Equal(t, map[string]string{"A": "alpha", "B": "beta", "C": ""}, res.Values)
}

func TestResolveAll_RefValue(t *testing.T) {
	reg := resolver.NewRegistry()
	require.NoError(t, reg.Register(fakeResolver{scheme: "fake", value: "resolved"}))

	p := profile.Profile{Vars: map[string]profile.Var{"K": ref("fake://anything")}}
	res := resolver.ResolveAll(context.Background(), reg, p)

	assert.Empty(t, res.Errors)
	assert.Equal(t, "resolved", res.Values["K"])
}

// TestResolveAll_PartialError checks fail-soft behaviour: one failing variable
// is recorded as an error while the others still resolve.
func TestResolveAll_PartialError(t *testing.T) {
	reg := resolver.DefaultRegistry()
	p := profile.Profile{Vars: map[string]profile.Var{
		"OK1": lit("one"),
		"OK2": lit("two"),
		"BAD": ref("bogus://x"), // no resolver registered for "bogus"
	}}

	res := resolver.ResolveAll(context.Background(), reg, p)

	assert.Equal(t, map[string]string{"OK1": "one", "OK2": "two"}, res.Values)
	require.Len(t, res.Errors, 1)
	assert.ErrorIs(t, res.Errors["BAD"], resolver.ErrUnknownScheme)
}

func TestResolveAll_InvalidRefURI(t *testing.T) {
	reg := resolver.DefaultRegistry()
	p := profile.Profile{Vars: map[string]profile.Var{"BAD": ref("no-scheme")}}

	res := resolver.ResolveAll(context.Background(), reg, p)

	assert.Empty(t, res.Values)
	require.Len(t, res.Errors, 1)
	assert.ErrorIs(t, res.Errors["BAD"], resolver.ErrInvalidURI)
}

func TestResolveAll_Empty(t *testing.T) {
	res := resolver.ResolveAll(context.Background(), resolver.DefaultRegistry(), profile.Profile{})
	assert.Empty(t, res.Values)
	assert.Empty(t, res.Errors)
}

// TestResolveAll_Cancel cancels the context while resolutions are in flight
// and checks that every variable ends up in Errors with context.Canceled.
func TestResolveAll_Cancel(t *testing.T) {
	reg := resolver.NewRegistry()
	require.NoError(t, reg.Register(fakeResolver{scheme: "slow", delay: 100 * time.Millisecond}))

	vars := make(map[string]profile.Var, 10)
	for i := 0; i < 10; i++ {
		vars[fmt.Sprintf("V%02d", i)] = ref("slow://x")
	}
	p := profile.Profile{Vars: vars}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	res := resolver.ResolveAll(ctx, reg, p)

	assert.Empty(t, res.Values)
	require.Len(t, res.Errors, 10)
	for k, err := range res.Errors {
		assert.ErrorIsf(t, err, context.Canceled, "var %s", k)
	}
}

// TestResolveAll_Timeout uses a short deadline against a slow resolver and
// expects every variable to fail with context.DeadlineExceeded.
func TestResolveAll_Timeout(t *testing.T) {
	reg := resolver.NewRegistry()
	require.NoError(t, reg.Register(fakeResolver{scheme: "slow", delay: 100 * time.Millisecond}))

	p := profile.Profile{Vars: map[string]profile.Var{
		"A": ref("slow://x"),
		"B": ref("slow://x"),
		"C": ref("slow://x"),
	}}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	res := resolver.ResolveAll(ctx, reg, p)

	assert.Empty(t, res.Values)
	require.Len(t, res.Errors, 3)
	for k, err := range res.Errors {
		assert.ErrorIsf(t, err, context.DeadlineExceeded, "var %s", k)
	}
}

// TestResolveAll_Deterministic runs the same resolution repeatedly and checks
// the result is identical every time: variables are dispatched in sorted key
// order and resolution is otherwise side-effect free.
func TestResolveAll_Deterministic(t *testing.T) {
	reg := resolver.DefaultRegistry()
	p := profile.Profile{Vars: map[string]profile.Var{
		"GAMMA": lit("g"),
		"ALPHA": lit("a"),
		"DELTA": lit("d"),
		"BETA":  lit("b"),
		"EPS":   lit("e"),
	}}

	first := resolver.ResolveAll(context.Background(), reg, p)
	for i := 0; i < 10; i++ {
		got := resolver.ResolveAll(context.Background(), reg, p)
		assert.Equalf(t, first.Values, got.Values, "run %d values diverged", i)
		assert.Equalf(t, first.Errors, got.Errors, "run %d errors diverged", i)
	}
}

// TestResolveAll_ResolverError checks that an error returned by a resolver is
// surfaced verbatim in the result.
func TestResolveAll_ResolverError(t *testing.T) {
	sentinel := errors.New("boom")
	reg := resolver.NewRegistry()
	require.NoError(t, reg.Register(fakeResolver{scheme: "fake", err: sentinel}))

	p := profile.Profile{Vars: map[string]profile.Var{"K": ref("fake://x")}}
	res := resolver.ResolveAll(context.Background(), reg, p)

	assert.Empty(t, res.Values)
	assert.ErrorIs(t, res.Errors["K"], sentinel)
}
