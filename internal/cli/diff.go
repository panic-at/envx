package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/panic-at/envx/internal/config"
	"github.com/panic-at/envx/internal/mask"
	"github.com/panic-at/envx/internal/profile"
)

// newDiffCmd builds the "envx diff" command.
func newDiffCmd(opts *rootOptions) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "diff <profile1> <profile2>",
		Short: "Compare two profiles",
		Long: "diff compares the effective (extends-flattened) variables of two\n" +
			"profiles. Refs are compared as their URIs and never resolved, since\n" +
			"the resolved value may be secret. Sensitive literal values are masked;\n" +
			"ref URIs are always shown verbatim.",
		Example: "  # Human-readable diff\n" +
			"  envx diff dev prod\n\n" +
			"  # Machine-readable diff\n" +
			"  envx diff dev prod --format json",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name1, name2 := args[0], args[1]

			cfg, err := config.Load(opts.configPath)
			if err != nil {
				return err
			}
			if _, ok := cfg.Profiles[name1]; !ok {
				return fmt.Errorf("profile %q does not exist", name1)
			}
			if _, ok := cfg.Profiles[name2]; !ok {
				return fmt.Errorf("profile %q does not exist", name2)
			}
			eff1, err := cfg.Effective(name1)
			if err != nil {
				return err
			}
			eff2, err := cfg.Effective(name2)
			if err != nil {
				return err
			}
			result := profile.Diff(eff1, eff2)

			switch format {
			case "text":
				return renderDiffText(cmd.OutOrStdout(), result, opts.noColor)
			case "json":
				return renderDiffJSON(cmd.OutOrStdout(), result)
			default:
				return &ExitError{Code: 2, Err: fmt.Errorf("unknown diff format %q (valid: text, json)", format)}
			}
		},
	}
	cmd.Flags().StringVar(&format, "format", "text", "output format: text or json")
	return cmd
}

// diffValue renders the comparable value of v for the diff. A ref's URI is
// shown verbatim — the URI is structural, not the secret, and the diff never
// resolves it. A literal's value is masked when the variable is sensitive.
func diffValue(v profile.Var) string {
	if v.Type == profile.VarRef {
		return v.URI
	}
	if v.Sensitive {
		return mask.Mask(v.Value)
	}
	return v.Value
}

// renderDiffText writes a git-style colored diff. Changes are grouped into
// modified, added and removed sections, each ordered by key.
func renderDiffText(w io.Writer, d profile.DiffResult, noColor bool) error {
	add := color.New(color.FgGreen)
	del := color.New(color.FgRed)
	mod := color.New(color.FgYellow)
	if noColor {
		add.DisableColor()
		del.DisableColor()
		mod.DisableColor()
	}

	var added, removed, changed []profile.VarChange
	for _, c := range d.Changes {
		switch c.Kind {
		case profile.Added:
			added = append(added, c)
		case profile.Removed:
			removed = append(removed, c)
		case profile.Modified:
			changed = append(changed, c)
		}
	}

	var b strings.Builder
	for _, c := range changed {
		b.WriteString(mod.Sprintf("~ %s\n", c.Key))
		b.WriteString(del.Sprintf("    - %s\n", diffValue(c.From)))
		b.WriteString(add.Sprintf("    + %s\n", diffValue(c.To)))
	}
	for _, c := range added {
		b.WriteString(add.Sprintf("+ %s = %s\n", c.Key, diffValue(c.To)))
	}
	for _, c := range removed {
		b.WriteString(del.Sprintf("- %s = %s\n", c.Key, diffValue(c.From)))
	}
	if b.Len() > 0 {
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, "%d added, %d removed, %d changed\n", len(added), len(removed), len(changed))

	_, err := io.WriteString(w, b.String())
	return err
}

// diffJSONVar is the JSON representation of one side of a variable change.
type diffJSONVar struct {
	Type      string `json:"type"`
	Value     string `json:"value,omitempty"`
	URI       string `json:"uri,omitempty"`
	Sensitive bool   `json:"sensitive,omitempty"`
}

// diffJSONChange is the JSON representation of a single variable change.
type diffJSONChange struct {
	Key  string       `json:"key"`
	Kind string       `json:"kind"`
	From *diffJSONVar `json:"from,omitempty"`
	To   *diffJSONVar `json:"to,omitempty"`
}

// diffJSONOutput is the top-level JSON document produced by "diff --format json".
type diffJSONOutput struct {
	Added   int              `json:"added"`
	Removed int              `json:"removed"`
	Changed int              `json:"changed"`
	Changes []diffJSONChange `json:"changes"`
}

// toJSONVar converts a profile.Var to its JSON form, masking a sensitive
// literal's value just as the text output does.
func toJSONVar(v profile.Var) *diffJSONVar {
	jv := &diffJSONVar{Type: string(v.Type), Sensitive: v.Sensitive}
	if v.Type == profile.VarRef {
		jv.URI = v.URI
	} else if v.Sensitive {
		jv.Value = mask.Mask(v.Value)
	} else {
		jv.Value = v.Value
	}
	return jv
}

// renderDiffJSON writes the diff as a JSON document for automation.
func renderDiffJSON(w io.Writer, d profile.DiffResult) error {
	out := diffJSONOutput{Changes: []diffJSONChange{}}
	for _, c := range d.Changes {
		jc := diffJSONChange{Key: c.Key, Kind: c.Kind.String()}
		switch c.Kind {
		case profile.Added:
			jc.To = toJSONVar(c.To)
			out.Added++
		case profile.Removed:
			jc.From = toJSONVar(c.From)
			out.Removed++
		case profile.Modified:
			jc.From = toJSONVar(c.From)
			jc.To = toJSONVar(c.To)
			out.Changed++
		}
		out.Changes = append(out.Changes, jc)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
