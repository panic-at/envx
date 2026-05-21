# envx

> Local-first secrets & environment profile manager.

**Status: WIP** — this project is under active development. The CLI is being
built incrementally; commands and config formats may change until the first
tagged release.

`envx` manages environment-variable *profiles* (dev / staging / prod) in a
versionable, secure way. Unlike a flat `.env` file it supports:

- multiple named profiles with inheritance (`extends`)
- literal values **and** references to external vaults (1Password, AWS Secrets
  Manager, host environment, …)
- injecting variables into child processes without polluting your shell

## Demo

![envx demo: init, profile add, set, show, diff, export](docs/demo.gif)

> Recorded with [asciinema](https://asciinema.org). The script lives in
> [`docs/demo.sh`](docs/demo.sh) — run it standalone with `bash docs/demo.sh`,
> or replay the cast with `asciinema play docs/demo.cast`. Regeneration steps
> are documented at the top of `docs/demo.sh`.

## Installation

Requires **Go 1.22+**.

```sh
go install github.com/panic-at/envx/cmd/envx@latest
```

Or build from source:

```sh
git clone https://github.com/panic-at/envx
cd envx
make build      # produces ./bin/envx
```

## Commands

| Command | Purpose | Status |
|---------|---------|--------|
| `envx init` | create `.envx/` config in the current directory | ✅ done |
| `envx profile add <name>` | create a new profile | ✅ done |
| `envx profile list` | list profiles | ✅ done |
| `envx set <KEY> <value> --profile <p>` | set a literal value | ✅ done |
| `envx set <KEY> --ref <uri> --profile <p>` | set a vault reference | ✅ done |
| `envx show <profile>` | show vars (sensitive values masked; `--reveal` to show) | ✅ done |
| `envx diff <p1> <p2>` | diff two profiles | ✅ done |
| `envx export --profile <p> --format dotenv\|json\|shell` | export resolved vars | ✅ done |
| `envx run --profile <p> -- <cmd>` | run a command with vars injected | 🚧 planned |

Vault references (`op://…`, `aws-sm://…`) are parsed and validated today; live
resolution against 1Password and AWS Secrets Manager is still in progress.

## Development

```sh
make build       # compile to ./bin/envx
make test        # run tests
make test-race   # run tests with the race detector
make lint        # gofmt check + go vet
make cover       # test coverage summary
```

## License

To be added (MIT planned).
