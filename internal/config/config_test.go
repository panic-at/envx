package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/panic-at/envx/internal/config"
	"github.com/panic-at/envx/internal/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixture returns the path to a YAML fixture under testdata/config.
func fixture(name string) string {
	return filepath.Join("..", "..", "testdata", "config", name)
}

func TestLoadFixtures(t *testing.T) {
	tests := []struct {
		file    string
		wantErr string // substring; empty means the fixture must load cleanly
	}{
		{"valid_basic.yaml", ""},
		{"valid_extends.yaml", ""},
		{"valid_empty.yaml", ""},
		{"valid_minimal.yaml", ""},
		{"invalid_version.yaml", "unsupported config version"},
		{"invalid_self_extends.yaml", "cannot extend itself"},
		{"invalid_cycle.yaml", "forms a cycle"},
		{"invalid_missing_parent.yaml", "extends unknown profile"},
		{"invalid_unknown_field.yaml", "parse config"},
		{"invalid_var_type.yaml", "unknown type"},
		{"invalid_ref_no_uri.yaml", "has no uri"},
		{"invalid_literal_with_uri.yaml", "sets a uri"},
		{"invalid_var_name.yaml", "invalid name"},
		{"invalid_profile_name.yaml", "invalid name"},
		{"malformed.yaml", "parse config"},
	}
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			c, err := config.Load(fixture(tt.file))
			if tt.wantErr == "" {
				require.NoError(t, err)
				require.NotNil(t, c)
				assert.Equal(t, config.CurrentVersion, c.Version)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestLoadContents(t *testing.T) {
	c, err := config.Load(fixture("valid_basic.yaml"))
	require.NoError(t, err)

	dev := c.Profiles["dev"]
	assert.Equal(t, profile.Var{Type: profile.VarLiteral, Value: "postgres://localhost/myapp"}, dev.Vars["DATABASE_URL"])
	assert.Equal(t, profile.Var{Type: profile.VarRef, URI: "env://LOCAL_API_KEY", Sensitive: true}, dev.Vars["API_KEY"])
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := config.Load(filepath.Join(t.TempDir(), "missing.yaml"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read config")
}

func TestSaveLoadRoundTrip(t *testing.T) {
	c := config.New()
	c.Profiles["dev"] = profile.Profile{
		Vars: map[string]profile.Var{
			"FOO": {Type: profile.VarLiteral, Value: "bar"},
			"KEY": {Type: profile.VarRef, URI: "env://KEY", Sensitive: true},
		},
	}
	c.Profiles["prod"] = profile.Profile{Extends: "dev"}

	// A nested path also exercises parent-directory creation.
	path := filepath.Join(t.TempDir(), "nested", config.DefaultFile)
	require.NoError(t, c.Save(path))

	got, err := config.Load(path)
	require.NoError(t, err)
	assert.Equal(t, c, got)
}

func TestSaveUsesRestrictivePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), config.DefaultFile)
	require.NoError(t, config.New().Save(path))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestSaveRefusesInvalidConfig(t *testing.T) {
	c := &config.Config{Version: 99}
	err := c.Save(filepath.Join(t.TempDir(), config.DefaultFile))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "refusing to save invalid config")
}

func TestSaveDirCreationError(t *testing.T) {
	// A regular file cannot act as a parent directory: MkdirAll must fail.
	blocker := filepath.Join(t.TempDir(), "blocker")
	require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o600))

	err := config.New().Save(filepath.Join(blocker, "config.yaml"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create config dir")
}

func TestNewAndDefaultPath(t *testing.T) {
	c := config.New()
	assert.Equal(t, config.CurrentVersion, c.Version)
	assert.NotNil(t, c.Profiles)
	assert.NoError(t, c.Validate())

	assert.Equal(t, filepath.Join(".envx", "config.yaml"), config.DefaultPath())
}

func TestValidate(t *testing.T) {
	literalFOO := map[string]profile.Var{"FOO": {Type: profile.VarLiteral, Value: "bar"}}

	tests := []struct {
		name    string
		cfg     config.Config
		wantErr string // empty means valid
	}{
		{
			name: "valid",
			cfg: config.Config{Version: 1, Profiles: map[string]profile.Profile{
				"dev": {Vars: literalFOO},
			}},
		},
		{
			name:    "zero version",
			cfg:     config.Config{},
			wantErr: "unsupported config version 0",
		},
		{
			name: "invalid profile name",
			cfg: config.Config{Version: 1, Profiles: map[string]profile.Profile{
				"bad name": {},
			}},
			wantErr: "invalid name",
		},
		{
			name: "invalid variable name",
			cfg: config.Config{Version: 1, Profiles: map[string]profile.Profile{
				"dev": {Vars: map[string]profile.Var{"1BAD": {Type: profile.VarLiteral}}},
			}},
			wantErr: `variable "1BAD" has an invalid name`,
		},
		{
			name: "unknown variable type",
			cfg: config.Config{Version: 1, Profiles: map[string]profile.Profile{
				"dev": {Vars: map[string]profile.Var{"FOO": {Type: "weird"}}},
			}},
			wantErr: "unknown type",
		},
		{
			name: "ref without uri",
			cfg: config.Config{Version: 1, Profiles: map[string]profile.Profile{
				"dev": {Vars: map[string]profile.Var{"FOO": {Type: profile.VarRef}}},
			}},
			wantErr: "has no uri",
		},
		{
			name: "ref with value",
			cfg: config.Config{Version: 1, Profiles: map[string]profile.Profile{
				"dev": {Vars: map[string]profile.Var{"FOO": {Type: profile.VarRef, URI: "env://X", Value: "x"}}},
			}},
			wantErr: "is a ref but sets a value",
		},
		{
			name: "literal with uri",
			cfg: config.Config{Version: 1, Profiles: map[string]profile.Profile{
				"dev": {Vars: map[string]profile.Var{"FOO": {Type: profile.VarLiteral, URI: "env://X"}}},
			}},
			wantErr: "is literal but sets a uri",
		},
		{
			name: "self extends",
			cfg: config.Config{Version: 1, Profiles: map[string]profile.Profile{
				"dev": {Extends: "dev"},
			}},
			wantErr: "cannot extend itself",
		},
		{
			name: "extends unknown profile",
			cfg: config.Config{Version: 1, Profiles: map[string]profile.Profile{
				"dev": {Extends: "ghost"},
			}},
			wantErr: "extends unknown profile",
		},
		{
			name: "two node cycle",
			cfg: config.Config{Version: 1, Profiles: map[string]profile.Profile{
				"a": {Extends: "b"},
				"b": {Extends: "a"},
			}},
			wantErr: "forms a cycle",
		},
		{
			name: "three node cycle",
			cfg: config.Config{Version: 1, Profiles: map[string]profile.Profile{
				"a": {Extends: "b"},
				"b": {Extends: "c"},
				"c": {Extends: "a"},
			}},
			wantErr: "forms a cycle",
		},
		{
			name: "tail into cycle does not flag the tail",
			cfg: config.Config{Version: 1, Profiles: map[string]profile.Profile{
				"tail": {Extends: "b"},
				"b":    {Extends: "c"},
				"c":    {Extends: "b"},
			}},
			wantErr: "forms a cycle",
		},
		{
			name: "valid extends chain",
			cfg: config.Config{Version: 1, Profiles: map[string]profile.Profile{
				"base":    {Vars: literalFOO},
				"dev":     {Extends: "base"},
				"staging": {Extends: "dev"},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestValidateTailNotFlaggedAsCycleMember(t *testing.T) {
	// "tail" feeds into the b<->c cycle but is not itself part of it.
	cfg := config.Config{Version: 1, Profiles: map[string]profile.Profile{
		"tail": {Extends: "b"},
		"b":    {Extends: "c"},
		"c":    {Extends: "b"},
	}}
	err := cfg.Validate()
	require.Error(t, err)
	assert.NotContains(t, err.Error(), `profile "tail": extends forms a cycle`)
}

func TestEffective(t *testing.T) {
	cfg := config.Config{Version: 1, Profiles: map[string]profile.Profile{
		"base": {Vars: map[string]profile.Var{
			"A": {Type: profile.VarLiteral, Value: "base-a"},
			"B": {Type: profile.VarLiteral, Value: "base-b"},
		}},
		"dev": {Extends: "base", Vars: map[string]profile.Var{
			"B": {Type: profile.VarLiteral, Value: "dev-b"},
			"C": {Type: profile.VarLiteral, Value: "dev-c"},
		}},
		"staging": {Extends: "dev", Vars: map[string]profile.Var{
			"C": {Type: profile.VarLiteral, Value: "staging-c"},
		}},
	}}

	eff, err := cfg.Effective("staging")
	require.NoError(t, err)
	assert.Empty(t, eff.Extends, "effective profile is flattened")
	assert.Equal(t, "base-a", eff.Vars["A"].Value, "inherited from base")
	assert.Equal(t, "dev-b", eff.Vars["B"].Value, "dev overrides base")
	assert.Equal(t, "staging-c", eff.Vars["C"].Value, "staging overrides dev")
}

func TestEffectiveErrors(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Config
		profile string
		wantErr string
	}{
		{
			name:    "unknown profile",
			cfg:     config.Config{Version: 1, Profiles: map[string]profile.Profile{}},
			profile: "ghost",
			wantErr: "not found",
		},
		{
			name: "cyclic chain terminates",
			cfg: config.Config{Version: 1, Profiles: map[string]profile.Profile{
				"a": {Extends: "b"},
				"b": {Extends: "a"},
			}},
			profile: "a",
			wantErr: "cyclic extends chain",
		},
		{
			name: "dangling parent",
			cfg: config.Config{Version: 1, Profiles: map[string]profile.Profile{
				"dev": {Extends: "ghost"},
			}},
			profile: "dev",
			wantErr: "extends unknown profile",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.cfg.Effective(tt.profile)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}
