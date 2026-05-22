package runner

import (
	"sort"
	"strings"
)

// BuildEnv merges a host environment with a profile's resolved variables and
// returns the result as a sorted slice of "KEY=VALUE" entries, ready to assign
// to exec.Cmd.Env.
//
// host holds the host environment in os.Environ form; entries without an "="
// are skipped. The two boolean toggles decide the merge:
//
//   - inherit false: the result contains only the profile variables; the host
//     environment is dropped entirely (an isolated environment).
//   - inherit true, override true: host plus profile variables, with profile
//     variables winning any key collision.
//   - inherit true, override false: host plus profile variables, with host
//     variables winning any key collision.
//
// The returned slice is always non-nil — including the empty case — so callers
// can assign it to Cmd.Env without the child silently inheriting the parent's
// environment.
func BuildEnv(host []string, vars map[string]string, inherit, override bool) []string {
	merged := make(map[string]string, len(host)+len(vars))
	if inherit {
		for _, e := range host {
			if k, v, ok := strings.Cut(e, "="); ok {
				merged[k] = v
			}
		}
	}
	for k, v := range vars {
		if inherit && !override {
			if _, exists := merged[k]; exists {
				continue
			}
		}
		merged[k] = v
	}
	out := make([]string, 0, len(merged))
	for k, v := range merged {
		out = append(out, k+"="+v)
	}
	sort.Strings(out)
	return out
}
