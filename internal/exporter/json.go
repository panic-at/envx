package exporter

import (
	"encoding/json"
	"io"
)

// jsonExporter writes variables as a pretty-printed JSON object.
type jsonExporter struct{}

// Format returns FormatJSON.
func (jsonExporter) Format() Format { return FormatJSON }

// Export writes vars as a JSON object with two-space indentation and a
// trailing newline. The object is emitted directly, with no wrapper, so it can
// be consumed by jq and similar tools. encoding/json marshals map keys in
// sorted order, satisfying the deterministic-output requirement.
func (jsonExporter) Export(w io.Writer, vars map[string]string) error {
	if vars == nil {
		vars = map[string]string{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(vars)
}
