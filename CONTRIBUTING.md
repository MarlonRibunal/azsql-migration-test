# Contributing

Thanks for your interest in improving `azsql-migration-test`. This is a small,
focused CLI — contributions of any size are welcome.

## Prerequisites

- **Go** (version pinned in [`go.mod`](go.mod))
- **Docker** — to run the Azure SQL Database Developer container
- **`sqlpackage`** on your `PATH` — for `compare`/`validate` (schema
  extract/deploy-report/publish)
- **Azure SQL Database Developer preview access** — to run anything against a
  live container. While Developer is in private preview you need the registry
  password; see the [README](README.md#usage). You do **not** need preview
  access to build, vet, or run the unit tests.

## Build and test

```bash
make build   # -> ./bin/azsql-migration-test
make test    # go test ./...
make vet     # go vet ./...
make fmt     # go fmt ./...
```

Before opening a PR, make sure the same checks CI runs pass locally:

```bash
gofmt -l .      # must print nothing
go vet ./...
go build ./...
go test ./...
make release    # cross-compiles all five targets
```

## What CI does and does not cover

CI runs `gofmt`, `go vet`, `go build`, the unit tests, and the cross-platform
build. It does **not** exercise the container or `sqlpackage` integration —
that needs Docker plus preview access, which CI intentionally does not have. If
your change touches the container lifecycle, schema handling, or query replay,
please run the affected command against a live Developer container yourself and
say so in the PR.

## Code style

- Keep it idiomatic Go; match the surrounding code.
- `gofmt` is required (CI enforces it).
- Prefer the standard library — this project has no third-party Go dependencies,
  and keeping it that way is a feature.

## Commits and pull requests

- Write a clear commit subject in the imperative mood ("Add X", "Fix Y").
- Keep PRs focused; one logical change per PR is easiest to review.
- Update the [README](README.md) and [CHANGELOG.md](CHANGELOG.md) when behavior
  changes. New user-facing changes go under `## [Unreleased]`.

## Security and secrets

**Never commit secrets or paste them into issues/PRs.** That includes the
preview registry password, database connection-string passwords, and SA
passwords. Supply those at runtime via flags or environment variables (see the
README). If you think you have committed a secret, rotate it and open an issue.

## Reporting bugs and requesting features

Use the issue templates. Please redact any passwords before pasting command
output.
