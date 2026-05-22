// Package config loads, saves and validates the envx project configuration.
//
// The configuration is a YAML file, conventionally stored at .envx/config.yaml
// within a project. It declares the schema version and the named profiles; the
// profile package owns the profile and variable types themselves.
//
// Load rejects unknown fields and schema violations, so a successfully loaded
// Config is always well-formed. Save validates before writing and uses 0600
// permissions because the file may hold literal secret values. Effective
// flattens a profile's extends chain into a single variable set.
package config
