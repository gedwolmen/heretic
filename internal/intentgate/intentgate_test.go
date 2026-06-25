package intentgate

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDetect_Ultrawork verifies the canonical keyword matches.
func TestDetect_Ultrawork(t *testing.T) {
	t.Parallel()
	got := Detect("ultrawork build the thing")
	require.Equal(t, []Mode{ModeUltrawork}, got)
}

// TestDetect_UltraworkAlias verifies the ulw alias works.
func TestDetect_UltraworkAlias(t *testing.T) {
	t.Parallel()
	got := Detect("ulw ship it")
	require.Equal(t, []Mode{ModeUltrawork}, got)
}

// TestDetect_Search verifies search keyword.
func TestDetect_Search(t *testing.T) {
	t.Parallel()
	got := Detect("search for cats")
	require.Equal(t, []Mode{ModeSearch}, got)
}

// TestDetect_Analyze verifies analyze keyword.
func TestDetect_Analyze(t *testing.T) {
	t.Parallel()
	got := Detect("analyze this function")
	require.Equal(t, []Mode{ModeAnalyze}, got)
}

// TestDetect_Team verifies team keyword.
func TestDetect_Team(t *testing.T) {
	t.Parallel()
	got := Detect("team up and refactor this")
	require.Equal(t, []Mode{ModeTeam}, got)
}

// TestDetect_NoKeyword verifies that neutral input returns no modes.
func TestDetect_NoKeyword(t *testing.T) {
	t.Parallel()
	got := Detect("hello world")
	require.Empty(t, got)
}

// TestDetect_MultipleKeywords verifies that multiple keywords stack in
// canonical order (ultrawork, search, analyze, team).
func TestDetect_MultipleKeywords(t *testing.T) {
	t.Parallel()
	got := Detect("team: ultrawork and search for X then analyze Y")
	require.Equal(t, []Mode{ModeUltrawork, ModeSearch, ModeAnalyze, ModeTeam}, got)
}

// TestDetect_WordBoundary verifies that partial matches do not trigger.
func TestDetect_WordBoundary(t *testing.T) {
	t.Parallel()
	got := Detect("ultraworks is a word") // no boundary match
	require.Empty(t, got)
}

// TestDetect_CaseInsensitive verifies case insensitivity.
func TestDetect_CaseInsensitive(t *testing.T) {
	t.Parallel()
	got := Detect("ULTRAWORK build it")
	require.Equal(t, []Mode{ModeUltrawork}, got)
}

// TestInject_NoModes verifies the no-op case.
func TestInject_NoModes(t *testing.T) {
	t.Parallel()
	original := "you are a helpful assistant"
	got := Inject(original, nil)
	require.Equal(t, original, got)
}

// TestInject_OneMode verifies single-mode injection.
func TestInject_OneMode(t *testing.T) {
	t.Parallel()
	original := "you are a helpful assistant"
	got := Inject(original, []Mode{ModeUltrawork})
	require.Contains(t, got, original)
	require.Contains(t, got, "ultrawork mode")
}

// TestInject_MultipleModes verifies multi-mode injection preserves order.
func TestInject_MultipleModes(t *testing.T) {
	t.Parallel()
	original := "you are a helpful assistant"
	got := Inject(original, []Mode{ModeUltrawork, ModeSearch})
	idxUltra := strings.Index(got, "ultrawork mode")
	idxSearch := strings.Index(got, "search mode")
	require.NotEqual(t, -1, idxUltra)
	require.NotEqual(t, -1, idxSearch)
	require.Less(t, idxUltra, idxSearch, "ultrawork prompt must come before search prompt")
}
