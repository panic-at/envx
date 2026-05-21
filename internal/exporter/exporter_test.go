package exporter_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/panic-at/envx/internal/exporter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// basicVars is the variable set behind every *_basic fixture.
var basicVars = map[string]string{
	"API_URL": "https://api.example.com",
	"DEBUG":   "true",
	"PORT":    "8080",
}

// readFixture returns the contents of a fixture under testdata/expected.
func readFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "expected", name))
	require.NoErrorf(t, err, "fixture %s", name)
	return string(data)
}

// runExport serializes vars with the exporter for f and returns the output.
func runExport(t *testing.T, f exporter.Format, vars map[string]string) string {
	t.Helper()
	e, err := exporter.Get(f)
	require.NoError(t, err)
	var buf bytes.Buffer
	require.NoError(t, e.Export(&buf, vars))
	return buf.String()
}

func TestGet(t *testing.T) {
	for _, f := range []exporter.Format{exporter.FormatDotenv, exporter.FormatJSON, exporter.FormatShell} {
		e, err := exporter.Get(f)
		require.NoErrorf(t, err, "format %s", f)
		assert.Equal(t, f, e.Format())
	}

	_, err := exporter.Get(exporter.Format("toml"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown export format")
	assert.Contains(t, err.Error(), "dotenv, json, shell")
}

func TestAll(t *testing.T) {
	assert.Equal(t,
		[]exporter.Format{exporter.FormatDotenv, exporter.FormatJSON, exporter.FormatShell},
		exporter.All(),
		"All must return formats in alphabetical order")
}
