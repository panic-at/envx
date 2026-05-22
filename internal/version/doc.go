// Package version exposes the build metadata of the envx binary: the semantic
// version, the Git commit and the build date.
//
// The variables default to development-friendly values and are replaced at
// release time by GoReleaser via linker flags, for example:
//
//	go build -ldflags "\
//	  -X github.com/panic-at/envx/internal/version.Version=v0.1.0 \
//	  -X github.com/panic-at/envx/internal/version.Commit=a1b2c3d \
//	  -X github.com/panic-at/envx/internal/version.Date=2026-05-22T10:00:00Z"
//
// For local builds without those flags, Info falls back to the VCS metadata
// the Go toolchain embeds in the binary.
package version
