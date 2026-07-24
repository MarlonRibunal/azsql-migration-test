# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- CLI with three commands:
  - `validate` — extract the source schema, deploy it into the Azure SQL Database
    Developer container (the real compatibility test), and replay queries against
    the deployed schema.
  - `compare` — non-destructive schema diff via `sqlpackage DeployReport`, with
    change count and breaking-change detection (drops, data-loss alerts).
  - `replay` — run a `GO`-separated query set against the container, reporting
    per-batch pass/fail and timing.
- Container lifecycle management: pull, run (loopback-bound port,
  `--platform linux/amd64`), wait-until-ready, and teardown (`--keep`).
- Schema handling via `sqlpackage` (Extract → DeployReport → Publish); DeployReport
  XML parsed into a change count and breaking flag (with tests).
- Query replay via `sqlcmd` inside the container, with per-batch pass/fail, timing,
  and captured SQL errors in the report.
- Registry login gate: `--registry`, `--registry-user`, `--registry-password`
  (env fallbacks `AZDBDEV_REGISTRY[_USER|_PASSWORD]`). `docker login` runs via
  stdin only when a registry and password are supplied; skipped otherwise. The
  image is configured via `--image` / `AZDBDEV_IMAGE` — no image reference or
  credentials are baked into the source.
- Configuration flags with environment fallbacks: `--source` (`AZDBDEV_SOURCE`),
  `--queries`, `--image` (`AZDBDEV_IMAGE`), `--port`, `--database`, `--sa-password`
  (`AZDBDEV_SA_PASSWORD`), `--report-dir`.
- CI workflow (gofmt, vet, build, unit tests, cross-platform build) and a release
  workflow that publishes cross-platform binaries + checksums on a `v*` tag.
- `SECURITY.md`, `NOTICE` (non-affiliation / trademark / third-party terms),
  `CONTRIBUTING.md`, issue/PR templates, and Dependabot for GitHub Actions.
- MIT license.

### Security
- Container port is bound to `127.0.0.1` only.
- `--database` is validated as a safe SQL identifier before use in DDL.
- Secrets are supplied at runtime (flags/env), never stored in the source.

[Unreleased]: https://github.com/MarlonRibunal/azsql-migration-test/commits/main
