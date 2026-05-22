# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

_Nothing yet._

## [0.1.0] - 2026-05-22

The first public release: the complete MVP command set.

### Added

- `envx init` — create the `.envx/` configuration in the current directory.
- `envx profile add` / `envx profile list` — create and list profiles, with
  single-parent inheritance via `extends`.
- `envx set` — define a variable as a literal value or as a reference URI
  (`--ref`), with a `--sensitive` flag to mark secrets.
- `envx show` — print a profile's effective (inheritance-flattened) variables;
  sensitive values are masked unless `--reveal` is given.
- `envx diff` — compare two profiles in human-readable or JSON form.
- `envx export` — serialize a resolved profile as dotenv, JSON or shell.
- `envx run` — execute a command as a child process with a profile's variables
  injected, without leaking them into the user's shell. Signals are forwarded
  to the child and its exit code is propagated.
- Resolver layer with `literal`, `env`, `op` (1Password) and `aws-sm` (AWS
  Secrets Manager) URI schemes. The `op` and `aws-sm` schemes are parsed and
  validated; live resolution against those vaults is not yet implemented.
- Cross-platform release tooling: GoReleaser configuration and a tag-triggered
  release workflow producing Linux, macOS and Windows binaries.

[Unreleased]: https://github.com/panic-at/envx/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/panic-at/envx/releases/tag/v0.1.0
