// Package version exposes the build version of the envx binary.
//
// The Version variable is intended to be overridden at build time via
// linker flags, for example:
//
//	go build -ldflags "-X github.com/SEU_USER/envx/internal/version.Version=v1.2.3"
package version

// Version is the semantic version of the envx build. It defaults to "dev"
// for local builds and is replaced via -ldflags in release builds.
var Version = "dev"
