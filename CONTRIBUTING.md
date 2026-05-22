# Contributing to envx

Thanks for your interest in improving envx. This document explains how to set
up the project, the conventions we follow, and how changes are reviewed.

By participating you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

## Prerequisites

- **Go 1.22 or newer** (the module targets 1.22).
- For local release validation only: [GoReleaser](https://goreleaser.com) and
  [golangci-lint](https://golangci-lint.run).

## Getting started

```sh
git clone https://github.com/panic-at/envx
cd envx
make build      # compiles ./bin/envx
./bin/envx --help
```

## Development workflow

The `Makefile` wraps the common tasks:

```sh
make test        # run the test suite
make test-race   # run the suite with the race detector
make lint        # gofmt check + go vet
make cover       # test coverage summary
make build       # compile ./bin/envx
```

Before opening a pull request, please make sure the following all pass:

```sh
gofmt -l .                   # must print nothing
go vet ./...
golangci-lint run ./...      # errcheck, govet, staticcheck, gosec, ...
go test ./... -race
```

Tests for signal handling and concurrency must pass under `-race`; do not skip
them.

## Coding conventions

- Every exported function, type and variable has a doc comment that starts
  with the identifier's name and is a complete sentence.
- Each package has a `doc.go` describing its role.
- Commands never write to `os.Stdout`/`os.Stderr` directly — they use the
  writers attached to their `*cobra.Command`, so tests can capture output.
- Error messages are actionable: they say what to do, not only what failed.
- No new third-party dependencies without discussion first.

## Commit messages

We follow [Conventional Commits](https://www.conventionalcommits.org). The
type prefix drives the generated changelog, so use it consistently:

```
<type>(<optional scope>): <short summary>
```

Common types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`, `ci`,
`style`. Examples:

```
feat(run): forward SIGTERM to the child process
fix(resolver): handle an empty 1Password field reference
docs: document the config file format
```

## Proposing a feature

Open a [feature request issue](https://github.com/panic-at/envx/issues/new/choose)
before writing code for a non-trivial change. Describe the problem you are
solving and the use case — that lets us agree on the approach before you
invest time. Small, self-contained fixes can go straight to a pull request.

## Pull request policy

- One logical change per pull request; keep the diff focused.
- Fill in the pull request template checklist.
- Add or update tests for any behaviour change, and update `CHANGELOG.md`
  under `[Unreleased]`.
- CI must be green: build, tests (including `-race`), `gofmt`, `go vet` and
  `golangci-lint` all run on every pull request.
- A maintainer review is required before merge. We squash-merge, so the pull
  request title should itself be a valid Conventional Commit.

## License

By contributing, you agree that your contributions are licensed under the
[MIT License](LICENSE).
