# Security Policy

## Reporting a vulnerability

**Do not open a public issue for security reports**, and do not paste any
secrets (registry passwords, database connection strings, SA passwords, tokens)
into an issue, PR, or discussion.

Report privately using GitHub's
[Private Vulnerability Reporting](https://github.com/MarlonRibunal/azsql-migration-test/security/advisories/new)
(the **Security → Report a vulnerability** button on the repository). If that is
unavailable, open a minimal issue asking for a private contact channel — without
details — and the maintainer will follow up.

Please include: affected version (`azsql-migration-test version`), a description,
and reproduction steps with all secrets redacted.

## Handling credentials (important for users)

This tool shells out to external programs, and those programs accept credentials
as command-line arguments. As a result, **while `sqlpackage`, `sqlcmd`, or
`docker` are running, credentials passed to them can be visible in the host
process list** (`ps` / `/proc/<pid>/cmdline`) to other local users. This applies
to:

- the source connection-string password (`--source`), and
- the container SA password (`--sa-password`).

Mitigations in place and recommended practice:

- Supply secrets via environment variables (`AZDBDEV_SOURCE`,
  `AZDBDEV_SA_PASSWORD`, `AZDBDEV_REGISTRY_PASSWORD`) rather than flags, to keep
  them out of shell history.
- The `docker login` password is piped via stdin and never appears as an
  argument.
- The container's port is bound to `127.0.0.1` only, so it is not reachable from
  the network.
- Run the tool on a trusted host where you control who can read the process list.

## Supported versions

This is an early-stage project; only the latest release is supported.
