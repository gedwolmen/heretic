package hookengine

import (
	"context"
	"time"
)

// Runner executes a single hook with a timeout.
type Runner struct {
	// DefaultTimeout is the per-hook timeout. Default: 30s.
	DefaultTimeout time.Duration
}

// NewRunner returns a Runner with default timeout.
func NewRunner() *Runner {
	return &Runner{DefaultTimeout: 30 * time.Second}
}

// Run invokes the hook's function with the payload, applying the
// per-hook timeout. If the hook panics or returns an error, the result
// is zero-valued with the error.
func (r *Runner) Run(ctx context.Context, h *Hook, p Payload) (Result, error) {
	timeout := r.DefaultTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	// We don't recover panics here; let them propagate so the engine
	// can log them and continue with other hooks.
	return h.Fn(ctx, p)
}
