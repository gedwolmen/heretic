package boulder

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestManager_LoadEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	m := NewManager(filepath.Join(dir, "boulder.json"))
	st, err := m.Load()
	require.NoError(t, err)
	require.Nil(t, st.Current)
	require.Empty(t, st.History)
}

func TestManager_SetAndComplete(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	m := NewManager(filepath.Join(dir, "boulder.json"))
	cur, err := m.SetCurrent("first task")
	require.NoError(t, err)
	require.Equal(t, "first task", cur.Title)
	require.Equal(t, StatusActive, cur.Status)

	require.NoError(t, m.Complete("done"))
	st, err := m.Load()
	require.NoError(t, err)
	require.Nil(t, st.Current)
	require.Len(t, st.History, 1)
	require.Equal(t, "first task", st.History[0].Title)
	require.Equal(t, StatusCompleted, st.History[0].Status)
}

func TestManager_SetCurrentMovesPreviousToHistory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	m := NewManager(filepath.Join(dir, "boulder.json"))
	_, err := m.SetCurrent("first")
	require.NoError(t, err)
	_, err = m.SetCurrent("second")
	require.NoError(t, err)
	st, err := m.Load()
	require.NoError(t, err)
	require.Equal(t, "second", st.Current.Title)
	require.Len(t, st.History, 1)
	require.Equal(t, "first", st.History[0].Title)
	require.Equal(t, StatusRegressed, st.History[0].Status)
}

func TestManager_Block(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	m := NewManager(filepath.Join(dir, "boulder.json"))
	_, err := m.SetCurrent("x")
	require.NoError(t, err)
	require.NoError(t, m.Block("waiting on review"))
	st, _ := m.Load()
	require.Equal(t, StatusBlocked, st.Current.Status)
}

func TestManager_CompleteNoCurrent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	m := NewManager(filepath.Join(dir, "boulder.json"))
	require.Error(t, m.Complete(""))
}

func TestManager_TrimHistory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	m := NewManager(filepath.Join(dir, "boulder.json"))
	m.maxHist = 3
	for i := 0; i < 5; i++ {
		_, err := m.SetCurrent("task " + string(rune('A'+i)))
		require.NoError(t, err)
		require.NoError(t, m.Complete(""))
	}
	require.NoError(t, m.TrimHistory())
	st, _ := m.Load()
	require.LessOrEqual(t, len(st.History), 3)
}
