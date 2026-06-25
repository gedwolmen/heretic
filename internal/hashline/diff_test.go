package hashline

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStripHashes(t *testing.T) {
	t.Parallel()
	in := "alpha#YQYW\nbeta#PRZM\ngamma#WZVR"
	got := StripHashes(in)
	require.Equal(t, "alpha\nbeta\ngamma", got)
}

func TestStripHashes_NoHashes(t *testing.T) {
	t.Parallel()
	in := "alpha\nbeta"
	got := StripHashes(in)
	require.Equal(t, in, got)
}

func TestComputeDiff_AddedAndRemoved(t *testing.T) {
	t.Parallel()
	old := "alpha\nbeta\ngamma"
	new := "alpha\nBETA\ngamma\ndelta"
	hunks := ComputeDiff(old, new)
	// Removed: beta. Added: BETA, delta.
	require.NotEmpty(t, hunks)
	added := 0
	removed := 0
	for _, h := range hunks {
		switch h.Kind {
		case DiffAdded:
			added++
		case DiffRemoved:
			removed++
		}
	}
	require.Equal(t, 2, added)
	require.Equal(t, 1, removed)
}

func TestComputeDiff_NoChange(t *testing.T) {
	t.Parallel()
	hunks := ComputeDiff("alpha\nbeta", "alpha\nbeta")
	require.Empty(t, hunks)
}

func TestFormatDiff_IncludesPrefixes(t *testing.T) {
	t.Parallel()
	hunks := []DiffHunk{
		{Kind: DiffContext, Text: "alpha"},
		{Kind: DiffRemoved, OldLine: 2, Text: "beta"},
		{Kind: DiffAdded, NewLine: 2, Text: "BETA"},
	}
	got := FormatDiff(hunks)
	require.Contains(t, got, "  alpha")
	require.Contains(t, got, "- beta")
	require.Contains(t, got, "+ BETA")
}

func TestFormatDiffWithHashes_HashesOnAdded(t *testing.T) {
	t.Parallel()
	hunks := []DiffHunk{
		{Kind: DiffAdded, NewLine: 1, Text: "new line"},
	}
	got := FormatDiffWithHashes(hunks)
	require.Contains(t, got, "+ new line#")
	// Verify the hash is 4 chars from the alphabet.
	idx := strings.Index(got, "#")
	require.GreaterOrEqual(t, idx, 0)
	hash := got[idx+1 : idx+5]
	for _, c := range hash {
		require.True(t, strings.ContainsRune(Alphabet, c), "hash char %q not in alphabet", c)
	}
}
