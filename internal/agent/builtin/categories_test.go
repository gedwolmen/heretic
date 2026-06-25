package builtin

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllCategories(t *testing.T) {
	t.Parallel()
	cats := AllCategories()
	require.Len(t, cats, 8)
}

func TestDefaultCategoryMeta_AllHaveMetadata(t *testing.T) {
	t.Parallel()
	metas := DefaultCategoryMeta()
	for _, m := range metas {
		require.NotEmpty(t, m.Display, "category %q has empty display name", m.Name)
		require.NotEmpty(t, m.Description, "category %q has empty description", m.Name)
		require.NotEmpty(t, m.Model, "category %q has empty model", m.Name)
	}
}

func TestCategoryRegistry_Get(t *testing.T) {
	t.Parallel()
	r := NewCategoryRegistry()
	m, ok := r.Get(CategoryQuick)
	require.True(t, ok)
	require.Equal(t, "Quick", m.Display)
}

func TestCategoryRegistry_Unknown(t *testing.T) {
	t.Parallel()
	r := NewCategoryRegistry()
	_, err := r.GetOrError("nope")
	require.Error(t, err)
	var uce *UnknownCategoryError
	require.ErrorAs(t, err, &uce)
}

func TestCategoryRegistry_All(t *testing.T) {
	t.Parallel()
	r := NewCategoryRegistry()
	require.Len(t, r.All(), 8)
	// Verify all 8 default categories are present.
	all := r.All()
	sort.Slice(all, func(i, j int) bool { return all[i].Name < all[j].Name })
	names := make([]string, 0, len(all))
	for _, m := range all {
		names = append(names, string(m.Name))
	}
	require.Equal(t, []string{
		"artistry", "deep", "quick", "ultrabrain",
		"unspecified-high", "unspecified-low", "visual-engineering", "writing",
	}, names)
}
