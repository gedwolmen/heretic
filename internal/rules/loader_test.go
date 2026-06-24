package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestLoadRules_HappyPath verifies that rules are loaded from a directory.
func TestLoadRules_HappyPath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "style.md"), []byte("Always use Go."), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "deps.md"), []byte("Prefer stdlib."), 0o644))

	got, err := LoadRules(dir)
	require.NoError(t, err)
	require.Len(t, got, 2)
	// Sorted alphabetically: deps, style
	require.Equal(t, "deps", got[0].Name)
	require.Equal(t, "Prefer stdlib.", got[0].Content)
	require.Equal(t, "style", got[1].Name)
	require.Equal(t, "Always use Go.", got[1].Content)
}

// TestLoadRules_MissingDir verifies that a missing directory is not an error.
func TestLoadRules_MissingDir(t *testing.T) {
	t.Parallel()
	got, err := LoadRules(filepath.Join(t.TempDir(), "does-not-exist"))
	require.NoError(t, err)
	require.Empty(t, got)
}

// TestLoadRules_SkipsNonMarkdown verifies that non-.md files are skipped.
func TestLoadRules_SkipsNonMarkdown(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "rule.md"), []byte("valid"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ignored.txt"), []byte("ignored"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README"), []byte("ignored"), 0o644))

	got, err := LoadRules(dir)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "rule", got[0].Name)
}

// TestLoadRules_EmptyDir verifies that an empty directory returns an empty
// slice, not an error.
func TestLoadRules_EmptyDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	got, err := LoadRules(dir)
	require.NoError(t, err)
	require.Empty(t, got)
}

// TestLoadRules_PathIsFile verifies that passing a file (not a directory)
// returns an error.
func TestLoadRules_PathIsFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "not-a-dir.md")
	require.NoError(t, os.WriteFile(f, []byte("x"), 0o644))
	_, err := LoadRules(f)
	require.Error(t, err)
}

// TestLoadRules_SubdirsIgnored verifies that subdirectories are not
// recursively scanned (single-level directory only).
func TestLoadRules_SubdirsIgnored(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	require.NoError(t, os.Mkdir(sub, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "nested.md"), []byte("nested"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "top.md"), []byte("top"), 0o644))

	got, err := LoadRules(dir)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "top", got[0].Name)
}

// TestInject_NoRules verifies the no-op case.
func TestInject_NoRules(t *testing.T) {
	t.Parallel()
	original := "you are a helpful assistant"
	got := Inject(original, nil)
	require.Equal(t, original, got)
}

// TestInject_OneRule verifies single-rule injection includes a header.
func TestInject_OneRule(t *testing.T) {
	t.Parallel()
	original := "you are a helpful assistant"
	got := Inject(original, []Rule{
		{Name: "style", Content: "Always use Go."},
	})
	require.Contains(t, got, original)
	require.Contains(t, got, "# Rule: style")
	require.Contains(t, got, "Always use Go.")
}

// TestInject_MultipleRules verifies that multiple rules are concatenated in
// order, each with its own header.
func TestInject_MultipleRules(t *testing.T) {
	t.Parallel()
	original := "you are a helpful assistant"
	got := Inject(original, []Rule{
		{Name: "alpha", Content: "first"},
		{Name: "beta", Content: "second"},
	})
	idxAlpha := indexOf(got, "# Rule: alpha")
	idxBeta := indexOf(got, "# Rule: beta")
	require.NotEqual(t, -1, idxAlpha)
	require.NotEqual(t, -1, idxBeta)
	require.Less(t, idxAlpha, idxBeta)
}

// indexOf is a tiny helper to avoid pulling strings into the test file.
func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
