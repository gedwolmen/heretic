package team

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStorage_CreateAndLoad(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStorage(dir)
	cfg := Config{
		Name:  "alpha",
		Owner: "sess-1",
		Members: []Member{
			{Name: "lead", Kind: KindSubagent, Agent: "sisyphus"},
			{Name: "coder", Kind: KindCategory},
		},
		Created: time.Now(),
	}
	require.NoError(t, s.Create(cfg))
	got, err := s.LoadConfig("alpha")
	require.NoError(t, err)
	require.Equal(t, cfg.Name, got.Name)
	require.Equal(t, cfg.Owner, got.Owner)
	require.Len(t, got.Members, 2)
}

func TestStorage_AlreadyExists(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStorage(dir)
	cfg := Config{Name: "x", Owner: "y"}
	require.NoError(t, s.Create(cfg))
	err := s.Create(cfg)
	require.ErrorIs(t, err, ErrAlreadyExists)
}

func TestStorage_LoadConfigNotFound(t *testing.T) {
	t.Parallel()
	s := NewStorage(t.TempDir())
	_, err := s.LoadConfig("nope")
	require.ErrorIs(t, err, ErrNotFound)
}

func TestStorage_ListTeams(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStorage(dir)
	require.NoError(t, s.Create(Config{Name: "a", Owner: "x"}))
	require.NoError(t, s.Create(Config{Name: "b", Owner: "y"}))
	got, err := s.ListTeams()
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"a", "b"}, got)
}

func TestStorage_DeleteTeam(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStorage(dir)
	require.NoError(t, s.Create(Config{Name: "a", Owner: "x"}))
	require.NoError(t, s.DeleteTeam("a"))
	_, err := s.LoadConfig("a")
	require.ErrorIs(t, err, ErrNotFound)
}

func TestStorage_StateRoundtrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStorage(dir)
	require.NoError(t, s.Create(Config{Name: "x", Owner: "y"}))
	require.NoError(t, s.SaveState("x", State{Status: StatusActive, CurrentTask: "t1"}))
	got, err := s.LoadState("x")
	require.NoError(t, err)
	require.Equal(t, StatusActive, got.Status)
	require.Equal(t, "t1", got.CurrentTask)
}

func TestMailbox_SendAndList(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "mailbox", "bob"), 0o755))
	mb := NewMailbox(dir)
	msg := Message{From: "alice", To: "bob", Body: "hello"}
	require.NoError(t, mb.Send(msg))
	list, err := mb.List("bob")
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "hello", list[0].Body)
}

func TestMailbox_UnreadFilter(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "mailbox", "b"), 0o755))
	mb := NewMailbox(dir)
	require.NoError(t, mb.Send(Message{From: "a", To: "b", Body: "1"}))
	require.NoError(t, mb.Send(Message{From: "a", To: "b", Body: "2"}))
	list, _ := mb.List("b")
	require.NoError(t, mb.MarkRead("b", list[0].ID))
	unread, err := mb.Unread("b")
	require.NoError(t, err)
	require.Len(t, unread, 1)
	require.Equal(t, "2", unread[0].Body)
}

func TestMailbox_Delete(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "mailbox", "b"), 0o755))
	mb := NewMailbox(dir)
	require.NoError(t, mb.Send(Message{From: "a", To: "b", Body: "x"}))
	list, _ := mb.List("b")
	require.NoError(t, mb.Delete("b", list[0].ID))
	list, _ = mb.List("b")
	require.Empty(t, list)
}

func TestEligibility_Defaults(t *testing.T) {
	t.Parallel()
	require.True(t, IsEligible(AgentSisyphus))
	require.True(t, IsEligible(AgentAtlas))
	require.True(t, IsEligible(AgentSisyphusJunior))
	require.False(t, IsEligible(AgentOracle))
	require.False(t, IsEligible(AgentLibrarian))
	require.False(t, IsEligible(AgentExplore))
}

func TestEligibility_AllowedIncludesConditional(t *testing.T) {
	t.Parallel()
	require.True(t, IsAllowed(AgentHephaestus))
	require.False(t, IsAllowed(AgentOracle))
}

func TestEligibility_GetEntry(t *testing.T) {
	t.Parallel()
	entry, ok := GetEntry(AgentHephaestus)
	require.True(t, ok)
	require.Equal(t, Conditional, entry.Eligibility)
}

func TestWorktree_BranchName(t *testing.T) {
	t.Parallel()
	w := NewWorktreeManager("/tmp/teams/alpha")
	got := w.BranchName("alpha", "lead")
	require.Contains(t, got, "team/alpha/lead-")
}

func TestWorktree_WorktreeDir(t *testing.T) {
	t.Parallel()
	w := NewWorktreeManager("/tmp/teams/alpha")
	got := w.WorktreeDir("lead")
	require.Equal(t, "/tmp/teams/alpha/worktrees/lead", got)
}
