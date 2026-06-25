# Heretic Testing

- Use `github.com/stretchr/testify/require` for assertions.
- Co-locate `*_test.go` next to the file under test.
- Always use `t.Parallel()` unless the test mutates global state.
- Use `t.TempDir()` for temporary directories; do not clean up manually.
