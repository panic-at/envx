# envx

> Manage environment variables across named profiles, with references to
> 1Password and AWS Secrets Manager. Local-first. No server. No leaked secrets.

<!-- TODO: demo.gif — record with vhs or asciinema and drop it here.
     The script lives in docs/demo.sh. -->
<!-- ![envx demo](docs/demo.gif) -->

[![CI](https://github.com/panic-at/envx/actions/workflows/ci.yml/badge.svg)](https://github.com/panic-at/envx/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/panic-at/envx/branch/main/graph/badge.svg)](https://codecov.io/gh/panic-at/envx)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Latest release](https://img.shields.io/github/v/release/panic-at/envx?sort=semver)](https://github.com/panic-at/envx/releases)

> **Status: WIP.** The MVP command set is complete and tested. Vault references
> (`op://`, `aws-sm://`) are parsed and validated today; live resolution against
> 1Password and AWS Secrets Manager is the next milestone. See the
> [roadmap](#roadmap).

## The problem

Environment configuration tends to rot. A project accumulates a `.env`, a
`.env.local`, a `.env.staging`, each a near-copy of the others. Secrets get
pasted by hand from a password manager and go stale the moment they rotate.
Sooner or later someone commits a file with a live key in it — or you `export`
a variable into your shell and it silently bleeds into every other process you
run that day.

envx replaces that with **named profiles** kept in one versionable file,
**references** to the vault that already owns each secret, and a `run` command
that injects variables into a single child process and nowhere else.

## Quickstart

Requires **Go 1.22+**.

```sh
go install github.com/panic-at/envx/cmd/envx@latest
```

```sh
# 1. Create .envx/config.yaml in the current directory
envx init

# 2. Add a profile
envx profile add dev

# 3. Set some variables
envx set DATABASE_URL postgres://localhost/myapp --profile dev
envx set PORT 8080 --profile dev

# 4. Inspect the profile (sensitive values are masked)
envx show dev

# 5. Run a command with the profile injected — and nowhere else
envx run --profile dev -- printenv DATABASE_URL
```

After step 5, `echo $DATABASE_URL` in your shell still prints nothing: the
variables lived only inside the child process.

## How it works

- **Profiles.** A profile is a named set of variables — `dev`, `staging`,
  `prod`. They all live in a single `.envx/config.yaml` you commit to the repo.
- **Inheritance.** A profile may `extend` another. The child inherits every
  variable of its parent and overrides only what differs, so `prod` is a small
  diff over `staging` rather than a full copy.
- **Literals vs references.** A variable is either a **literal** (the value is
  stored inline) or a **reference** — a URI like `op://vault/item/field` or
  `aws-sm://region/secret` that points at the system already holding the
  secret. References mean secrets are never duplicated into the config.
- **Lazy resolution.** References are resolved only when a value is actually
  needed — by `run`, `export`, or `show --reveal`. `show` and `diff` work on
  the URIs themselves, so you can inspect and compare profiles without ever
  touching the underlying secret.
- **No leakage.** `envx run` builds the environment for one child process.
  Nothing is written to your shell, and nothing is written back to disk.

## Commands

| Command | Description | Example |
|---------|-------------|---------|
| `envx init` | Create `.envx/config.yaml` in the current directory. | `envx init` |
| `envx profile add <name>` | Create a profile, optionally inheriting another. | `envx profile add prod --extends dev` |
| `envx profile list` | List profiles with variable counts. | `envx profile list` |
| `envx set <KEY> <value>` | Set a literal variable. | `envx set PORT 8080 --profile dev` |
| `envx set <KEY> --ref <uri>` | Set a vault reference. | `envx set DB_PASS --ref op://v/db/pass --profile dev` |
| `envx show <profile>` | Show effective variables (masked; `--reveal` to resolve). | `envx show dev` |
| `envx diff <p1> <p2>` | Diff two profiles (text or `--format json`). | `envx diff dev prod` |
| `envx export --profile <p>` | Export resolved vars as `dotenv`, `json` or `shell`. | `envx export --profile dev --format json` |
| `envx run --profile <p> -- <cmd>` | Run a command with the profile injected. | `envx run --profile dev -- ./server` |

Run `envx <command> --help` for the full flag list and more examples.

## Comparison

| | envx | plain `.env` | direnv | dotenv-vault | Doppler |
|---|:--:|:--:|:--:|:--:|:--:|
| Named profiles | ✅ | ❌ | ❌ | ✅ | ✅ |
| Profile inheritance | ✅ | ❌ | ⚠️ dir nesting | ❌ | ⚠️ |
| References to a vault | ✅ | ❌ | ⚠️ via hooks | ❌ | ✅ |
| Local-first (no sync) | ✅ | ✅ | ✅ | ❌ | ❌ |
| Works with no server/account | ✅ | ✅ | ✅ | ❌ | ❌ |
| Open source | ✅ | ✅ | ✅ | ⚠️ partial | ❌ |

envx aims to sit where direnv's locality meets Doppler's profiles and vault
integration — without a hosted service in the path.

## Config format

`envx init` writes `.envx/config.yaml`. The file is plain YAML you can read,
diff and commit. Directories are created `0700` and the file `0600`, since it
may hold literal values.

```yaml
# Schema version — managed by envx, do not edit by hand.
version: 1

profiles:
  # A base profile with two literal variables.
  dev:
    vars:
      DATABASE_URL:
        type: literal
        value: postgres://localhost/myapp
      LOG_LEVEL:
        type: literal
        value: debug

  # prod inherits everything from dev and overrides one variable.
  prod:
    extends: dev
    vars:
      DATABASE_URL:
        # A reference: the value is resolved from 1Password at run time.
        type: ref
        uri: op://prod-vault/database/url
        sensitive: true   # masked in `show` / `diff` output
```

Prefer the `envx` commands over editing the file directly — they validate the
schema on every write.

## Roadmap

Post-MVP, roughly in priority order:

- [ ] Live **1Password** resolution (`op://`) via the 1Password CLI.
- [ ] Live **AWS Secrets Manager** resolution (`aws-sm://`).
- [ ] **HashiCorp Vault** resolver.
- [ ] Shell hook to auto-load a profile on `cd` (direnv-style).
- [ ] `envx init --force`, `envx profile rm`, `envx profile rename`.
- [ ] `envx unset <KEY>` and bulk import from an existing `.env`.
- [ ] Shell completions (bash, zsh, fish) and man pages.
- [ ] `--unsafe` flag to reveal sensitive values in full when explicitly asked.

See [CHANGELOG.md](CHANGELOG.md) for what has shipped.

## Development

```sh
make build       # compile to ./bin/envx
make test        # run tests
make test-race   # run tests with the race detector
make lint        # gofmt check + go vet
make cover       # test coverage summary
```

`golangci-lint run ./...` runs the full linter set used in CI.

## Contributing

Contributions are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md) for the
development workflow, commit conventions and PR policy. All participants are
expected to follow the [Code of Conduct](CODE_OF_CONDUCT.md).

## License

[MIT](LICENSE) © 2026 the envx authors
