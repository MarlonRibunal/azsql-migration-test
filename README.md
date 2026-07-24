# azsql-migration-test

[![CI](https://github.com/MarlonRibunal/azsql-migration-test/actions/workflows/ci.yml/badge.svg)](https://github.com/MarlonRibunal/azsql-migration-test/actions/workflows/ci.yml)

Validate Azure SQL Database migrations locally by running them against the **Azure SQL
Database Developer** container — the high-fidelity local Azure SQL Database
engine. Catch breaking schema changes before you deploy. No cloud account, no
cloud cost.

> _Independent community project — **not** affiliated with, endorsed by, or
> sponsored by Microsoft. "Azure", "Azure SQL Database", and "az" are trademarks
> of Microsoft Corporation. See [NOTICE](NOTICE)._

> **Configure the image.** You must point the tool at an Azure SQL Database
> Developer container image you have access to. Set the image and, if its
> registry needs authentication, the registry and credentials — via flags or
> environment variables:
>
> ```bash
> export AZDBDEV_IMAGE='<registry>/<repo>:<tag>'
> export AZDBDEV_REGISTRY='<registry host>'            # if the registry needs a login
> export AZDBDEV_REGISTRY_USER='<username>'
> export AZDBDEV_REGISTRY_PASSWORD='<password>'
> azsql-migration-test validate --source "..." --queries ./queries.sql
> ```
>
> When a registry and password are set, the CLI runs `docker login` for you; when
> they are empty it assumes you are already authenticated (or the image is public)
> and skips login. The Developer image is `linux/amd64` only; the tool runs it
> under emulation on arm64.

## Commands

| Command    | What it does                                                    |
|------------|----------------------------------------------------------------|
| `validate` | Full pass: schema comparison + query replay.                   |
| `compare`  | Schema diff (source vs. container) via `sqlpackage`.           |
| `replay`   | Run a query set against the container; report execution + timing. |

## Install

```bash
go install github.com/MarlonRibunal/azsql-migration-test@latest
```

Or download a pre-built binary for your platform from the
[Releases](https://github.com/MarlonRibunal/azsql-migration-test/releases) page
(each release ships linux/darwin/windows amd64+arm64 binaries and a
`checksums.txt`).

Or build from source:

```bash
make build   # produces ./bin/azsql-migration-test
```

Maintainers cut a release by pushing a version tag:

```bash
git tag v0.1.0 && git push origin v0.1.0   # triggers the release workflow
```

## Usage

Reports are written to `./migration-report/` (override with `--report-dir`).
While Developer is in private preview, export the registry password once so the
CLI can pull the container:

```bash
export AZDBDEV_REGISTRY_PASSWORD='<preview password from the Azure SQL DB Developer team>'
```

### `validate` — full pass (schema deploy + query replay)

```bash
azsql-migration-test validate \
  --source "Server=prod-sql.database.windows.net;Database=Sales;User Id=svc;Password=***" \
  --queries ./testdata/sample-queries.sql
```

```
✓ Schema deployed to Azure SQL Database Developer (compatible)
✓ Query replay complete: 2 passed, 0 failed
Migration validation completed successfully
```

Exits non-zero if the schema fails to deploy (incompatible) or any query fails.

### `compare` — schema diff only (non-destructive)

Extract the source schema and report what would change against the engine,
without publishing anything.

```bash
azsql-migration-test compare \
  --source "Server=prod-sql.database.windows.net;Database=Sales;User Id=svc;Password=***" \
  --report-dir ./migration-report
```

```
✓ Schema comparison complete
```

The DeployReport lands at `./migration-report/schema-diff.xml`, and
`report.md` summarizes it, e.g.:

```
## Schema
Changes: 4 (breaking: false)
DeployReport: `schema-diff.xml`
```

Exits non-zero when the diff contains a breaking operation (a drop or a
data-loss alert).

### `replay` — run a query set only

Runs each `GO`-separated batch against `--database` and reports per-batch
pass/fail and timing. Useful against a schema you already deployed (pair with
`validate --keep`, or point `--database` at an existing database).

```bash
azsql-migration-test replay \
  --queries ./testdata/sample-queries.sql \
  --database MigrationValidation
```

```
✓ Query replay complete: 2 passed, 0 failed
```

A failing batch exits non-zero and the SQL error is captured in `report.md`,
for example:

```
- [FAIL] SELECT * FROM dbo.ThisTableDoesNotExist; (168ms)
    Msg 208, Level 16, State 1 — Invalid object name 'dbo.ThisTableDoesNotExist'.
```

### Keep the container up between runs

```bash
azsql-migration-test validate --source "..." --queries ./q.sql --keep
azsql-migration-test replay --queries ./more.sql --database MigrationValidation
```

### Common flags

| Flag            | Default                 | Meaning                                   |
|-----------------|-------------------------|-------------------------------------------|
| `--source`      | `AZDBDEV_SOURCE`        | Source connection string (schema origin). |
| `--queries`     | —                       | Path to a `.sql` file to replay.          |
| `--image`       | `AZDBDEV_IMAGE` (required) | Developer container image reference.  |
| `--port`        | `1433`                  | Host port mapped to the container.        |
| `--sa-password` | `AZDBDEV_SA_PASSWORD`   | SA password for the local container.      |
| `--report-dir`  | `./migration-report`    | Output directory for reports.             |
| `--keep`        | `false`                 | Keep the container running after the run. |
| `--registry`         | `AZDBDEV_REGISTRY`         | Registry to log in to (empty skips login). |
| `--registry-user`    | `AZDBDEV_REGISTRY_USER`    | Registry username.                        |
| `--registry-password`| `AZDBDEV_REGISTRY_PASSWORD` | Registry password. Empty skips login (GA). |

Prefer `AZDBDEV_REGISTRY_PASSWORD` over `--registry-password` to keep the secret
out of your shell history and the process list. If you have already run
`docker login` yourself, leave the password empty and the CLI skips its own login.

## How it works

1. Log in to the preview registry (if a password is supplied) and pull the image.
2. Start the container locally on `--port`.
3. Import the source schema via `sqlpackage`.
4. Compare source vs. container schema; write a diff report.
5. Replay the query set via `sqlcmd`, capturing pass/fail and timing.
6. Tear the container down (unless `--keep`).

**On timing:** replay timing is reported for information only. A local container
is not a cloud service tier, so local timing is not a cloud performance signal.

## Known limits

- **A local pass is high-confidence, not a guarantee.** The Developer engine can
  be *more permissive* than the cloud service for rules the Azure gateway
  enforces above the engine — for example it accepts `USE [db]`, which classic
  Azure SQL Database rejects. So a migration can pass locally and still be
  rejected in the cloud. Treat a green result as "very likely compatible," and
  keep your normal staged rollout.
- **Breaking-change detection is heuristic.** The `compare` breaking flag is
  derived from the `sqlpackage` DeployReport (drops and data-loss alerts); it can
  under-report if Microsoft changes that report format. The XML report is always
  written so you can inspect it yourself.

## Security

Credentials passed to `sqlpackage`/`sqlcmd`/`docker` can be visible in the host
process list while those tools run. Prefer the environment variables
(`AZDBDEV_SOURCE`, `AZDBDEV_SA_PASSWORD`, `AZDBDEV_REGISTRY_PASSWORD`) over flags,
and see [SECURITY.md](SECURITY.md) for details and how to report a vulnerability
privately. The container port is bound to `127.0.0.1` only.

## Requirements

- **Docker** — runs the Developer container.
- **`sqlpackage`** on the host `PATH` — extracts, reports, and publishes schema.
- Azure SQL Database Developer preview access (while in private preview).

`sqlcmd` is invoked *inside* the container via `docker exec`, so you do not need
it on the host.

## How validation works

- `compare` — extract the source schema to a dacpac, then run a non-destructive
  `sqlpackage DeployReport` against the container's target database. Reports the
  change count and flags breaking operations (drops, data-loss alerts).
- `validate` — the above, then **publish** the schema into the container. The
  publish is the real compatibility test: if the schema uses a construct the
  Azure SQL Database engine rejects, the publish fails and validation fails.
  Then replays the queries against the deployed schema.
- `replay` — run a query set against `--database` via `sqlcmd`, per-batch
  pass/fail and timing.

## Status

Working, verified end-to-end against a live Azure SQL Database Developer
container image:

- `compare` — extracts a source schema, spins up the container, and reports the
  diff (verified: 4 create operations parsed, breaking flag correct).
- `validate` — extract → deploy-report → **publish** → replay (verified: schema
  published, queries against the deployed tables pass).
- `replay` — pass and fail paths both verified, including non-zero exit and the
  SQL error captured in the report.

`compare` and `validate` need a `--source` Azure SQL Database connection string
and `sqlpackage` on the host `PATH`.

## Changelog

See [CHANGELOG.md](CHANGELOG.md).

## License

MIT — see [LICENSE](LICENSE).
