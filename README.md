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

## Installation

Requires **Go 1.22+**.

```sh
go install github.com/SEU_USER/envx/cmd/envx@latest
```

Or build from source:

```sh
git clone https://github.com/SEU_USER/envx
cd envx
make build      # produces ./bin/envx
```

## Planned commands

| Command | Purpose |
|---------|---------|
| `envx init` | create `.envx/` config in the current directory |
| `envx profile add <name>` | create a new profile |
| `envx profile list` | list profiles |
| `envx set <KEY> <value> --profile <p>` | set a literal value |
| `envx set <KEY> --ref <uri> --profile <p>` | set a vault reference |
| `envx show <profile>` | show vars (sensitive values masked; `--reveal` to show) |
| `envx diff <p1> <p2>` | diff two profiles |
| `envx run --profile <p> -- <cmd>` | run a command with vars injected |
| `envx export --profile <p> --format dotenv\|json\|shell` | export resolved vars |

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
