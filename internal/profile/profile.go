// Package profile defines the core domain model for envx: profiles and the
// environment variables they contain, plus operations to merge and diff them.
//
// The package performs no I/O — loading, saving and validation live in the
// config package, which imports this one.
package profile

// VarType identifies how a variable's value is obtained at resolution time.
type VarType string

const (
	// VarLiteral is a variable whose value is stored inline in the config.
	VarLiteral VarType = "literal"
	// VarRef is a variable whose value is a URI resolved from an external
	// source (host environment, 1Password, AWS Secrets Manager, ...).
	VarRef VarType = "ref"
)

// Var is a single environment variable definition.
//
// A literal variable carries its value in Value; a ref variable carries a
// resolver URI in URI. The Sensitive flag controls whether the resolved value
// is masked in human-facing output.
type Var struct {
	Type      VarType `yaml:"type"`
	Value     string  `yaml:"value,omitempty"`
	URI       string  `yaml:"uri,omitempty"`
	Sensitive bool    `yaml:"sensitive,omitempty"`
}

// Profile is a named set of environment variable definitions. A profile may
// inherit from another profile by name via Extends; the inheritance is
// flattened by config.Config.Effective.
type Profile struct {
	Extends string         `yaml:"extends,omitempty"`
	Vars    map[string]Var `yaml:"vars,omitempty"`
}

// Merge returns a new Profile whose variables are those of base overridden,
// key by key, by those of override. A key present in override fully replaces
// the base definition for that key.
//
// The result has an empty Extends, since inheritance is considered flattened.
// Neither input is modified, and the returned Vars map is always non-nil.
func Merge(base, override Profile) Profile {
	out := Profile{Vars: make(map[string]Var, len(base.Vars)+len(override.Vars))}
	for k, v := range base.Vars {
		out.Vars[k] = v
	}
	for k, v := range override.Vars {
		out.Vars[k] = v
	}
	return out
}
