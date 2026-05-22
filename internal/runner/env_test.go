package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildEnv(t *testing.T) {
	host := []string{"PATH=/bin", "SHARED=host"}
	vars := map[string]string{"FOO": "profile", "SHARED": "profile"}

	tests := []struct {
		name     string
		inherit  bool
		override bool
		want     []string
	}{
		{
			name:     "no inherit yields only profile vars",
			inherit:  false,
			override: true,
			want:     []string{"FOO=profile", "SHARED=profile"},
		},
		{
			name:     "no inherit ignores override",
			inherit:  false,
			override: false,
			want:     []string{"FOO=profile", "SHARED=profile"},
		},
		{
			name:     "inherit with override lets the profile win",
			inherit:  true,
			override: true,
			want:     []string{"FOO=profile", "PATH=/bin", "SHARED=profile"},
		},
		{
			name:     "inherit without override lets the host win",
			inherit:  true,
			override: false,
			want:     []string{"FOO=profile", "PATH=/bin", "SHARED=host"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildEnv(host, vars, tt.inherit, tt.override)
			assert.Equal(t, tt.want, got, "merged environment must be sorted and correct")
		})
	}
}

func TestBuildEnv_SkipsEntriesWithoutEquals(t *testing.T) {
	got := BuildEnv([]string{"VALID=1", "GARBAGE"}, nil, true, true)
	assert.Equal(t, []string{"VALID=1"}, got, "host entries without '=' are dropped")
}

func TestBuildEnv_IsAlwaysNonNil(t *testing.T) {
	got := BuildEnv(nil, nil, false, false)
	assert.NotNil(t, got, "an empty result must still be a non-nil slice so Cmd.Env is not inherited")
	assert.Empty(t, got)
}

func TestBuildEnv_EmptyHostValue(t *testing.T) {
	got := BuildEnv([]string{"EMPTY="}, nil, true, true)
	assert.Equal(t, []string{"EMPTY="}, got, "a host variable set to the empty string is preserved")
}
