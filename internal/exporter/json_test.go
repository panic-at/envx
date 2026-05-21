package exporter_test

import (
	"encoding/json"
	"testing"

	"github.com/panic-at/envx/internal/exporter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON_BasicFixture(t *testing.T) {
	assert.Equal(t, readFixture(t, "json_basic.json"), runExport(t, exporter.FormatJSON, basicVars))
}

func TestJSON_Cases(t *testing.T) {
	tests := []struct {
		name string
		vars map[string]string
		want string
	}{
		{
			name: "empty is an empty object",
			vars: map[string]string{},
			want: "{}\n",
		},
		{
			name: "single value",
			vars: map[string]string{"FOO": "bar"},
			want: "{\n  \"FOO\": \"bar\"\n}\n",
		},
		{
			name: "multiple values are sorted by key",
			vars: map[string]string{"ZED": "z", "ALPHA": "a"},
			want: "{\n  \"ALPHA\": \"a\",\n  \"ZED\": \"z\"\n}\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, runExport(t, exporter.FormatJSON, tt.vars))
		})
	}
}

// TestJSON_SpecialCharsRoundTrip checks that values needing JSON escaping
// (quotes, newlines, unicode) survive a parse cycle unchanged.
func TestJSON_SpecialCharsRoundTrip(t *testing.T) {
	vars := map[string]string{
		"QUOTED":  `say "hi"`,
		"NEWLINE": "line1\nline2",
		"UNICODE": "café 🚀",
		"TABBED":  "a\tb",
	}
	out := runExport(t, exporter.FormatJSON, vars)

	var got map[string]string
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	assert.Equal(t, vars, got)
}
