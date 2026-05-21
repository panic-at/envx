// Package config loads, saves and validates the envx project configuration,
// stored as YAML under a project's .envx directory.
package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/panic-at/envx/internal/profile"
	"gopkg.in/yaml.v3"
)

const (
	// CurrentVersion is the config schema version this build reads and writes.
	CurrentVersion = 1
	// DefaultDir is the per-project directory holding envx configuration.
	DefaultDir = ".envx"
	// DefaultFile is the configuration file name inside DefaultDir.
	DefaultFile = "config.yaml"
)

// Config is the top-level envx project configuration: a schema version and a
// set of named profiles.
type Config struct {
	Version  int                        `yaml:"version"`
	Profiles map[string]profile.Profile `yaml:"profiles"`
}

// DefaultPath returns the conventional config file path, ".envx/config.yaml".
func DefaultPath() string {
	return filepath.Join(DefaultDir, DefaultFile)
}

// New returns an empty, valid Config at the current schema version.
func New() *Config {
	return &Config{
		Version:  CurrentVersion,
		Profiles: map[string]profile.Profile{},
	}
}

// Load reads, parses and validates the envx config file at path. Unknown YAML
// fields and schema violations are reported as errors.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var c Config
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config %s: %w", path, err)
	}
	return &c, nil
}

// Save validates the config and writes it as YAML to path, creating parent
// directories as needed. The file is written with 0600 permissions because it
// may hold literal secret values.
func (c *Config) Save(path string) error {
	if err := c.Validate(); err != nil {
		return fmt.Errorf("refusing to save invalid config: %w", err)
	}
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create config dir %s: %w", dir, err)
		}
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(c); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}
	return nil
}

// Effective resolves the named profile by walking its extends chain to the
// root and merging variables downward, so child definitions override their
// parents'. The returned profile has a flattened variable set and no Extends.
//
// Effective terminates even on a cyclic or dangling extends chain, returning a
// descriptive error in that case; a valid config (see Validate) never has one.
func (c *Config) Effective(name string) (profile.Profile, error) {
	if _, ok := c.Profiles[name]; !ok {
		return profile.Profile{}, fmt.Errorf("profile %q not found", name)
	}
	var chain []string
	seen := map[string]bool{}
	for cur := name; cur != ""; {
		if seen[cur] {
			return profile.Profile{}, fmt.Errorf("cyclic extends chain at profile %q", cur)
		}
		seen[cur] = true
		p, ok := c.Profiles[cur]
		if !ok {
			return profile.Profile{}, fmt.Errorf("profile %q extends unknown profile %q", chain[len(chain)-1], cur)
		}
		chain = append(chain, cur)
		cur = p.Extends
	}
	result := profile.Profile{Vars: map[string]profile.Var{}}
	for i := len(chain) - 1; i >= 0; i-- {
		result = profile.Merge(result, c.Profiles[chain[i]])
	}
	return result, nil
}
