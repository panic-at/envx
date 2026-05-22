package exporter

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// Format identifies a supported serialization format.
type Format string

const (
	// FormatDotenv is the KEY=value dotenv format.
	FormatDotenv Format = "dotenv"
	// FormatJSON is a pretty-printed JSON object.
	FormatJSON Format = "json"
	// FormatShell is a POSIX shell script of export statements.
	FormatShell Format = "shell"
)

// Exporter serializes a set of environment variables.
type Exporter interface {
	// Format returns the format this exporter produces.
	Format() Format
	// Export writes the serialized representation of vars to w. Keys are
	// iterated in deterministic (alphabetical) order.
	Export(w io.Writer, vars map[string]string) error
}

// registry holds the exporter for every supported format.
var registry = map[Format]Exporter{
	FormatDotenv: dotenvExporter{},
	FormatJSON:   jsonExporter{},
	FormatShell:  shellExporter{},
}

// Get returns the Exporter for format f, or an error if f is not supported.
func Get(f Format) (Exporter, error) {
	e, ok := registry[f]
	if !ok {
		return nil, fmt.Errorf("unknown export format %q (valid: %s)", f, strings.Join(formatStrings(), ", "))
	}
	return e, nil
}

// All returns every supported format in alphabetical order.
func All() []Format {
	formats := make([]Format, 0, len(registry))
	for f := range registry {
		formats = append(formats, f)
	}
	sort.Slice(formats, func(i, j int) bool { return formats[i] < formats[j] })
	return formats
}

// formatStrings returns the supported formats as a sorted string slice.
func formatStrings() []string {
	all := All()
	out := make([]string, len(all))
	for i, f := range all {
		out[i] = string(f)
	}
	return out
}

// sortedKeys returns the keys of vars in ascending lexical order.
func sortedKeys(vars map[string]string) []string {
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
