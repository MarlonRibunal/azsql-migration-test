<!-- Thanks for contributing! Keep PRs focused — one logical change is easiest to review. -->

## What & why

<!-- What does this change, and what problem does it solve? -->

## How it was tested

<!-- CI covers gofmt/vet/build/unit tests. If you touched the container,
     sqlpackage, or replay path, note whether you ran it against a live
     Azure SQL Database Developer container. -->

- [ ] `go test ./...` passes
- [ ] `gofmt -l .` is clean and `go vet ./...` passes
- [ ] Ran the affected command(s) against a live container (if applicable)

## Checklist

- [ ] No secrets committed (registry password, connection strings, SA password)
- [ ] Updated README / CHANGELOG.md if user-facing behavior changed
