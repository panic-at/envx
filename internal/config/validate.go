package config

import (
	"errors"
	"fmt"
	"regexp"
	"sort"

	"github.com/panic-at/envx/internal/profile"
)

var (
	// profileNameRe restricts profile names to a filesystem- and CLI-friendly
	// character set.
	profileNameRe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	// envVarNameRe is the conventional POSIX-ish environment variable name.
	envVarNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

// Validate checks the config against the envx schema rules, returning all
// violations joined into a single error, or nil if the config is valid.
//
// It enforces a supported schema version; well-formed profile and variable
// names; that each variable is a valid literal or ref; and that profile
// extends links point to existing profiles without self-reference or cycles.
func (c *Config) Validate() error {
	var errs []error

	if c.Version != CurrentVersion {
		errs = append(errs, fmt.Errorf("unsupported config version %d (expected %d)", c.Version, CurrentVersion))
	}

	names := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		p := c.Profiles[name]
		if !profileNameRe.MatchString(name) {
			errs = append(errs, fmt.Errorf("profile %q: invalid name (allowed: letters, digits, '_', '-')", name))
		}
		errs = append(errs, validateVars(name, p)...)
		errs = append(errs, validateExtends(c, name, p)...)
	}

	errs = append(errs, detectCycles(c, names)...)

	return errors.Join(errs...)
}

// validateVars checks the variable names and literal/ref invariants of a
// single profile.
func validateVars(profileName string, p profile.Profile) []error {
	keys := make([]string, 0, len(p.Vars))
	for k := range p.Vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var errs []error
	for _, k := range keys {
		v := p.Vars[k]
		if !envVarNameRe.MatchString(k) {
			errs = append(errs, fmt.Errorf("profile %q: variable %q has an invalid name", profileName, k))
		}
		switch v.Type {
		case profile.VarLiteral:
			if v.URI != "" {
				errs = append(errs, fmt.Errorf("profile %q: variable %q is literal but sets a uri", profileName, k))
			}
		case profile.VarRef:
			if v.URI == "" {
				errs = append(errs, fmt.Errorf("profile %q: variable %q is a ref but has no uri", profileName, k))
			}
			if v.Value != "" {
				errs = append(errs, fmt.Errorf("profile %q: variable %q is a ref but sets a value", profileName, k))
			}
		default:
			errs = append(errs, fmt.Errorf("profile %q: variable %q has unknown type %q (want %q or %q)",
				profileName, k, v.Type, profile.VarLiteral, profile.VarRef))
		}
	}
	return errs
}

// validateExtends checks that a profile's direct extends link, if any, is not
// a self-reference and points to an existing profile. Multi-node cycles are
// handled separately by detectCycles.
func validateExtends(c *Config, name string, p profile.Profile) []error {
	if p.Extends == "" {
		return nil
	}
	if p.Extends == name {
		return []error{fmt.Errorf("profile %q: cannot extend itself", name)}
	}
	if _, ok := c.Profiles[p.Extends]; !ok {
		return []error{fmt.Errorf("profile %q: extends unknown profile %q", name, p.Extends)}
	}
	return nil
}

// detectCycles reports every profile that is part of an extends cycle. A
// profile is in a cycle if and only if following its extends chain returns to
// it. Self-extends and dangling links are skipped here, since validateExtends
// already reports them.
func detectCycles(c *Config, names []string) []error {
	var errs []error
	for _, start := range names {
		seen := map[string]bool{}
		for cur := start; ; {
			p, ok := c.Profiles[cur]
			if !ok || p.Extends == "" || p.Extends == cur {
				break // dangling or self-extend: already reported elsewhere
			}
			if p.Extends == start {
				errs = append(errs, fmt.Errorf("profile %q: extends forms a cycle", start))
				break
			}
			if seen[p.Extends] {
				break // a cycle that does not include start; its members report it
			}
			seen[p.Extends] = true
			cur = p.Extends
		}
	}
	return errs
}
