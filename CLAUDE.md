# Terraform Provider for Basis Theory

Terraform provider for managing Basis Theory resources — Go-based provider using the Terraform Plugin SDK.

## Development Workflow

```bash
go build ./...        # Build the provider
make verify           # Run full verification (acceptance tests)
```

## Testing

```bash
go test ./...                          # Run all tests
go test ./... -run "TestAccName"       # Targeted test
TF_ACC=1 go test ./... -timeout 120m  # Full acceptance tests (what make verify runs)
```

## Feedback Loops

Run `go test ./... -run "TestAccName"` for targeted test feedback.

When a failing test is discovered, always verify it passes using the appropriate feedback loop before considering the fix complete.

## Standards & Conventions

- Go, Terraform Plugin SDK
- Provider resources and data sources follow Terraform conventions
- `make update-docs` to regenerate provider documentation

## Links

- [Terraform Provider docs](https://developers.basistheory.com/docs/api/terraform/)
- [Terraform Registry](https://registry.terraform.io/providers/Basis-Theory/basistheory/latest)
