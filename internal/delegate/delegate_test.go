package delegate

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestRegistry_RegisterResolve verifies the basic registry behavior.
func TestRegistry_RegisterResolve(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	r.Register("explore", "explore-agent")
	got, err := r.Resolve("explore")
	require.NoError(t, err)
	require.Equal(t, "explore-agent", got)
}

// TestRegistry_UnknownCategory verifies that unknown categories error.
func TestRegistry_UnknownCategory(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	_, err := r.Resolve("nope")
	require.Error(t, err)
}

// TestRegistry_Default verifies the default registry has the standard categories.
func TestRegistry_Default(t *testing.T) {
	t.Parallel()
	r := DefaultRegistry()
	for _, cat := range []Category{"explore", "librarian", "search", "edit"} {
		_, err := r.Resolve(cat)
		require.NoError(t, err, "category %q should be registered", cat)
	}
}

// TestLimiter_DefaultLimit verifies the limiter caps concurrency at the
// configured limit and that excess tasks wait (not fail).
func TestLimiter_DefaultLimit(t *testing.T) {
	t.Parallel()
	lim := NewLimiter(2)
	release1, err := lim.Acquire(context.Background(), "k")
	require.NoError(t, err)
	release2, err := lim.Acquire(context.Background(), "k")
	require.NoError(t, err)

	// Third acquire must block; verify with a short context.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err = lim.Acquire(ctx, "k")
	require.ErrorIs(t, err, context.DeadlineExceeded)

	release1()
	release2()

	// After release, acquire succeeds.
	release3, err := lim.Acquire(context.Background(), "k")
	require.NoError(t, err)
	release3()
}

// TestLimiter_ReleaseIsIdempotent verifies that calling release twice does
// not double-free a slot.
func TestLimiter_ReleaseIsIdempotent(t *testing.T) {
	t.Parallel()
	lim := NewLimiter(1)
	release, err := lim.Acquire(context.Background(), "k")
	require.NoError(t, err)
	release()
	release() // should not panic / not double-release

	// We should still be able to acquire.
	release2, err := lim.Acquire(context.Background(), "k")
	require.NoError(t, err)
	release2()
}

// TestLimiter_PerKeyIsolation verifies that limit is per-key, not global.
func TestLimiter_PerKeyIsolation(t *testing.T) {
	t.Parallel()
	lim := NewLimiter(1)
	rel1, err := lim.Acquire(context.Background(), "k1")
	require.NoError(t, err)
	defer rel1()
	// Different key should not be blocked.
	rel2, err := lim.Acquire(context.Background(), "k2")
	require.NoError(t, err)
	defer rel2()
}

// TestTool_Execute_HappyPath verifies the basic delegation flow with a fake
// spawner.
func TestTool_Execute_HappyPath(t *testing.T) {
	t.Parallel()
	spawner := &fakeSpawner{}
	t2 := NewTool(spawner)
	res, err := t2.Execute(context.Background(), Params{
		Category: "explore",
		Prompt:   "find something",
	})
	require.NoError(t, err)
	require.Equal(t, "child-1", res.SessionID)
	require.Equal(t, int32(1), spawner.count.Load())
}

// TestTool_Execute_EmptyPrompt verifies the empty-prompt guard.
func TestTool_Execute_EmptyPrompt(t *testing.T) {
	t.Parallel()
	spawner := &fakeSpawner{}
	t2 := NewTool(spawner)
	_, err := t2.Execute(context.Background(), Params{
		Category: "explore",
		Prompt:   "",
	})
	require.ErrorIs(t, err, ErrEmptyPrompt)
	require.Equal(t, int32(0), spawner.count.Load())
}

// TestTool_Execute_UnknownCategory verifies the registry error path.
func TestTool_Execute_UnknownCategory(t *testing.T) {
	t.Parallel()
	spawner := &fakeSpawner{}
	t2 := NewTool(spawner)
	_, err := t2.Execute(context.Background(), Params{
		Category: "bogus",
		Prompt:   "x",
	})
	require.Error(t, err)
	require.Equal(t, int32(0), spawner.count.Load())
}

// TestTool_Execute_ConcurrencyLimit verifies the limiter blocks excess
// concurrent calls (default 5). Spawns 10 with limit 5; the first 5 run
// concurrently, the next 5 wait. No spawns should fail.
func TestTool_Execute_ConcurrencyLimit(t *testing.T) {
	t.Parallel()
	spawner := &slowSpawner{delay: 30 * time.Millisecond}
	t2 := NewToolWithLimits(spawner, 5)

	const N = 10
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := t2.Execute(context.Background(), Params{
				Category: "explore",
				Prompt:   "x",
			})
			require.NoError(t, err)
		}()
	}
	wg.Wait()
	require.Equal(t, int32(N), spawner.count.Load(), "all 10 spawns must complete without error")
}

// TestTool_Execute_RespectsContextCancel verifies the limiter respects ctx.
func TestTool_Execute_RespectsContextCancel(t *testing.T) {
	t.Parallel()
	spawner := &fakeSpawner{}
	t2 := NewToolWithLimits(spawner, 1)
	// Saturate the limit.
	rel, err := t2.Limiter.Acquire(context.Background(), "/")
	require.NoError(t, err)
	defer rel()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err = t2.Execute(ctx, Params{
		Category: "explore",
		Prompt:   "x",
	})
	require.Error(t, err)
}

// fakeSpawner returns a deterministic session ID and counts calls.
type fakeSpawner struct {
	count atomic.Int32
}

func (f *fakeSpawner) SpawnChild(ctx context.Context, req DelegateRequest) (DelegateResult, error) {
	f.count.Add(1)
	return DelegateResult{
		SessionID: fmtSessionID(f.count.Load()),
		Output:    "ok",
	}, nil
}

func fmtSessionID(n int32) string {
	return "child-" + itoa(n)
}

// itoa is a tiny integer-to-string helper to avoid pulling in fmt for the
// test fixture.
func itoa(n int32) string {
	if n == 0 {
		return "0"
	}
	var buf [16]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}

// slowSpawner counts calls and sleeps briefly to force the limiter to
// actually block concurrent callers.
type slowSpawner struct {
	delay time.Duration
	count atomic.Int32
}

func (s *slowSpawner) SpawnChild(ctx context.Context, req DelegateRequest) (DelegateResult, error) {
	s.count.Add(1)
	select {
	case <-time.After(s.delay):
	case <-ctx.Done():
		return DelegateResult{}, ctx.Err()
	}
	return DelegateResult{
		SessionID: fmtSessionID(s.count.Load()),
		Output:    "ok",
	}, nil
}
