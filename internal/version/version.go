package version

import "runtime/debug"

// Build metadata of the envx binary.
//
// These variables default to development-friendly values and are overridden at
// release time by GoReleaser via -ldflags. See doc.go for the exact flags.
var (
	// Version is the semantic version of the build, e.g. "v0.1.0". It is
	// "dev" for local builds.
	Version = "dev"
	// Commit is the Git commit the build was produced from. It is "none"
	// when unset; for local builds Info fills it from the embedded VCS data.
	Commit = "none"
	// Date is the build timestamp in RFC 3339 form. It is "unknown" when
	// unset.
	Date = "unknown"
)

// Info returns a single-line, human-readable build description, for example
// "v0.1.0 (a1b2c3d, built 2026-05-22T10:00:00Z)".
//
// For local builds, where Commit was not injected, Info recovers the commit
// from the binary's embedded VCS metadata when available, so that `envx
// --version` is still informative without a release build.
func Info() string {
	commit := Commit
	if commit == "none" {
		if rev, ok := vcsRevision(); ok {
			commit = rev
		}
	}
	return Version + " (" + commit + ", built " + Date + ")"
}

// vcsRevision returns the Git revision embedded by the Go toolchain in the
// build, abbreviated to 7 characters, and reports whether one was found.
func vcsRevision() (string, bool) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", false
	}
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" && s.Value != "" {
			rev := s.Value
			if len(rev) > 7 {
				rev = rev[:7]
			}
			return rev, true
		}
	}
	return "", false
}
