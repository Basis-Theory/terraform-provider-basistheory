# Terraform Provider for Basis Theory

Go-based Terraform provider using Terraform Plugin SDK v2 and the Basis Theory Go SDK v5.

## Build & Test

```bash
go build ./...                                    # Build
go test ./... -run "TestAccName"                  # Run specific acceptance test
TF_ACC=1 go test ./... -timeout 120m             # All acceptance tests (make verify)
make update-docs                                  # Regenerate provider docs (go generate)
```

Always verify fixes with targeted tests before considering done.

## Project Structure

- `internal/provider/` — All provider code (resources, data sources, helpers, tests)
- `templates/` — Doc templates for `terraform-plugin-docs`
- `docs/` — Generated docs (do NOT edit manually — run `make update-docs`)
- `main.go` — Entry point

## Gotchas

- **ALL tests are acceptance tests**: Every test requires `TF_ACC=1` and real API credentials. There are no unit tests. Running `go test ./...` without `TF_ACC=1` skips all tests silently.
- **Environment variables required**: Tests load `.env.local` from repo root via `godotenv`. Required vars: `BASISTHEORY_API_KEY`, `BASISTHEORY_API_URL`. Copy `.env.example` to `.env.local`.
- **Tests hit real API**: Acceptance tests create/read/update/delete real resources against the configured API. Use a dev/test environment, never production.
- **Single package**: All provider code is in `internal/provider/` — resources, tests, and helpers are all in package `provider`.
- **Resource naming**: Files follow `resource_basistheory_<name>.go` with matching `resource_basistheory_<name>_test.go`.
- **Resources available**: `basistheory_application`, `basistheory_application_key`, `basistheory_reactor`, `basistheory_proxy`, `basistheory_webhook`, `basistheory_applepay_domain`.
- **Go SDK v5**: Uses `github.com/Basis-Theory/go-sdk/v5` with `client` and `option` sub-packages.
- **Docs are generated**: `make update-docs` runs `go generate` which uses `terraform-plugin-docs`. Templates are in `templates/`. Never edit files in `docs/` directly.
- **Go 1.22**: Required Go version per go.mod.
- **Test timeout**: Acceptance tests can be slow — CI uses 120m timeout.

## Release

Automated on push to `master`. CI tags, updates CHANGELOG, then runs GoReleaser with GPG signing to publish to Terraform Registry.

## Docs

- [Terraform Provider docs](https://developers.basistheory.com/docs/api/terraform/)
- [Terraform Registry](https://registry.terraform.io/providers/Basis-Theory/basistheory/latest)
