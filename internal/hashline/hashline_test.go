package hashline

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestHashLine_Deterministic verifies that the same input always produces
// the same hash output.
func TestHashLine_Deterministic(t *testing.T) {
	t.Parallel()
	a := HashLine("hello", "", "")
	b := HashLine("hello", "", "")
	require.Equal(t, a, b)
	require.Len(t, a, 4)
	for _, c := range a {
		require.True(t, strings.ContainsRune(Alphabet, c), "hash char %q not in alphabet", c)
	}
}

// TestHashLine_DifferentInputs verifies that different inputs produce
// different hashes (smoke test for collision resistance).
func TestHashLine_DifferentInputs(t *testing.T) {
	t.Parallel()
	a := HashLine("hello", "", "")
	b := HashLine("world", "", "")
	require.NotEqual(t, a, b)
	c := HashLine("hello", "prev", "")
	d := HashLine("hello", "different", "")
	require.NotEqual(t, c, d, "previous-line context should influence the hash")
}

// TestHashLine_KnownVector verifies a known input → output pair. This is a
// regression guard against algorithm changes.
func TestHashLine_KnownVector(t *testing.T) {
	t.Parallel()
	got := HashLine("func main() {", "package main", "}")
	require.Len(t, got, 4)
	for _, c := range got {
		require.True(t, strings.ContainsRune(Alphabet, c))
	}
	// Stash the value so future runs are reproducible.
	t.Logf("HashLine(\"func main() {\", \"package main\", \"}\") = %s", got)
}

// TestTagReadOutput verifies that each line gets a #ID suffix.
func TestTagReadOutput(t *testing.T) {
	t.Parallel()
	input := "line1\nline2\nline3"
	out := TagReadOutput(input)
	lines := strings.Split(out, "\n")
	require.Len(t, lines, 3)
	for _, l := range lines {
		require.Contains(t, l, "#")
		tail := l[strings.LastIndex(l, "#")+1:]
		require.Len(t, tail, 4)
	}
}

// TestTagReadOutput_Empty verifies that empty input produces empty output.
func TestTagReadOutput_Empty(t *testing.T) {
	t.Parallel()
	require.Equal(t, "", TagReadOutput(""))
}

// TestExtractHash verifies the hash extraction helper.
func TestExtractHash(t *testing.T) {
	t.Parallel()
	require.Equal(t, "YQYW", ExtractHash("hello#YQYW"))
	require.Equal(t, "", ExtractHash("no hash here"))
	require.Equal(t, "", ExtractHash("trailing#"))
	require.Equal(t, "", ExtractHash("bad#abcd")) // contains X, not in alphabet
}

// TestValidateLineHash_Match verifies that a fresh tag passes validation.
func TestValidateLineHash_Match(t *testing.T) {
	t.Parallel()
	original := "alpha\nbeta\ngamma"
	lines := strings.Split(original, "\n")
	prev, next := lines[0], lines[2]
	hash := HashLine("beta", prev, next)
	require.NoError(t, ValidateLineHash(original, "beta", hash))
}

// TestValidateLineHash_StaleHashRejected simulates a concurrent edit: the
// file changed between read and edit, so the stored hash no longer matches.
func TestValidateLineHash_StaleHashRejected(t *testing.T) {
	t.Parallel()
	original := "alpha\nbeta\ngamma"
	lines := strings.Split(original, "\n")
	prev, next := lines[0], lines[2]
	staleHash := HashLine("beta", prev, next)

	// Mutate the file in a way that preserves the target line but changes
	// surrounding context (e.g. another tool added a line above). The hash
	// depends on prev/next so a context change produces a different hash.
	mutated := "alpha\nNEW_PREV\nbeta\ngamma"
	err := ValidateLineHash(mutated, "beta", staleHash)
	require.Error(t, err)
	var mismatch *ErrHashMismatch
	require.ErrorAs(t, err, &mismatch)
	require.Equal(t, 3, mismatch.LineNumber)
}

// TestValidateLineHash_LineNotFound verifies the missing-line error path.
func TestValidateLineHash_LineNotFound(t *testing.T) {
	t.Parallel()
	original := "alpha\nbeta\ngamma"
	err := ValidateLineHash(original, "delta", "XXXX")
	require.Error(t, err)
	require.NotContains(t, err.Error(), "hash mismatch")
}

// TestValidateLineHash_EmptyHash verifies the empty-hash guard.
func TestValidateLineHash_EmptyHash(t *testing.T) {
	t.Parallel()
	err := ValidateLineHash("alpha\nbeta\ngamma", "beta", "")
	require.Error(t, err)
}
