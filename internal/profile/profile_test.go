package profile_test

import (
	"testing"

	"github.com/panic-at/envx/internal/profile"
	"github.com/stretchr/testify/assert"
)

func lit(v string) profile.Var { return profile.Var{Type: profile.VarLiteral, Value: v} }

func TestMerge(t *testing.T) {
	tests := []struct {
		name     string
		base     profile.Profile
		override profile.Profile
		want     map[string]profile.Var
	}{
		{
			name:     "both empty",
			base:     profile.Profile{},
			override: profile.Profile{},
			want:     map[string]profile.Var{},
		},
		{
			name:     "empty base inherits override",
			base:     profile.Profile{},
			override: profile.Profile{Vars: map[string]profile.Var{"A": lit("a")}},
			want:     map[string]profile.Var{"A": lit("a")},
		},
		{
			name:     "empty override keeps base",
			base:     profile.Profile{Vars: map[string]profile.Var{"A": lit("a")}},
			override: profile.Profile{},
			want:     map[string]profile.Var{"A": lit("a")},
		},
		{
			name:     "override wins on conflict, union otherwise",
			base:     profile.Profile{Vars: map[string]profile.Var{"A": lit("base-a"), "B": lit("base-b")}},
			override: profile.Profile{Vars: map[string]profile.Var{"B": lit("over-b"), "C": lit("over-c")}},
			want:     map[string]profile.Var{"A": lit("base-a"), "B": lit("over-b"), "C": lit("over-c")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profile.Merge(tt.base, tt.override)
			assert.Equal(t, tt.want, got.Vars)
			assert.NotNil(t, got.Vars, "merged Vars must never be nil")
			assert.Empty(t, got.Extends, "merge flattens inheritance")
		})
	}
}

func TestMergeDoesNotMutateInputs(t *testing.T) {
	base := profile.Profile{Vars: map[string]profile.Var{"A": lit("base-a")}}
	override := profile.Profile{Vars: map[string]profile.Var{"A": lit("over-a")}}

	_ = profile.Merge(base, override)

	assert.Equal(t, lit("base-a"), base.Vars["A"], "base must be untouched")
	assert.Equal(t, lit("over-a"), override.Vars["A"], "override must be untouched")
}

func TestDiff(t *testing.T) {
	tests := []struct {
		name string
		a, b profile.Profile
		want []profile.VarChange
	}{
		{
			name: "identical profiles have no changes",
			a:    profile.Profile{Vars: map[string]profile.Var{"A": lit("a")}},
			b:    profile.Profile{Vars: map[string]profile.Var{"A": lit("a")}},
			want: nil,
		},
		{
			name: "added variable",
			a:    profile.Profile{},
			b:    profile.Profile{Vars: map[string]profile.Var{"A": lit("a")}},
			want: []profile.VarChange{{Key: "A", Kind: profile.Added, To: lit("a")}},
		},
		{
			name: "removed variable",
			a:    profile.Profile{Vars: map[string]profile.Var{"A": lit("a")}},
			b:    profile.Profile{},
			want: []profile.VarChange{{Key: "A", Kind: profile.Removed, From: lit("a")}},
		},
		{
			name: "modified variable",
			a:    profile.Profile{Vars: map[string]profile.Var{"A": lit("old")}},
			b:    profile.Profile{Vars: map[string]profile.Var{"A": lit("new")}},
			want: []profile.VarChange{{Key: "A", Kind: profile.Modified, From: lit("old"), To: lit("new")}},
		},
		{
			name: "mixed changes are ordered by key",
			a: profile.Profile{Vars: map[string]profile.Var{
				"KEEP": lit("same"), "DROP": lit("gone"), "EDIT": lit("v1"),
			}},
			b: profile.Profile{Vars: map[string]profile.Var{
				"KEEP": lit("same"), "ADD": lit("fresh"), "EDIT": lit("v2"),
			}},
			want: []profile.VarChange{
				{Key: "ADD", Kind: profile.Added, To: lit("fresh")},
				{Key: "DROP", Kind: profile.Removed, From: lit("gone")},
				{Key: "EDIT", Kind: profile.Modified, From: lit("v1"), To: lit("v2")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profile.Diff(tt.a, tt.b)
			assert.Equal(t, tt.want, got.Changes)
			assert.Equal(t, len(tt.want) == 0, got.Empty())
		})
	}
}

func TestChangeKindString(t *testing.T) {
	tests := []struct {
		kind profile.ChangeKind
		want string
	}{
		{profile.Added, "added"},
		{profile.Removed, "removed"},
		{profile.Modified, "modified"},
		{profile.ChangeKind(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.kind.String())
		})
	}
}
