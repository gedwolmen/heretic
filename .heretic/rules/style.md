# Heretic Style

- Prefer stdlib over external dependencies.
- Use `gofmt`/`gofumpt`. Do not bikeshed whitespace in PRs.
- Tests go next to the code they test (`foo.go` → `foo_test.go`).
- Use `errors.Is` / `errors.As` for error inspection. Avoid `==` on errors.
