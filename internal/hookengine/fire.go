package hookengine

import (
	"context"
	"sync"
)

// Aggregated is the combined result of all hooks firing on one event.
// For guard events it contains the combined decision; for transform
// events it contains the chain of transformations.
type Aggregated struct {
	// Decision is the most restrictive decision across all hooks.
	// Deny > Allow > None.
	Decision string
	// Reason is the concatenation of all deny reasons (newline-joined).
	Reason string
	// Halt is true if any hook asked to halt.
	Halt bool
	// Context is the concatenation of all context additions.
	Context string
	// Transformed is the last Transformed result from any hook.
	Transformed any
	// HookCount is how many hooks ran.
	HookCount int
	// Errors are non-fatal hook errors (logged but not returned).
	Errors []error
}

// Engine runs hooks. It uses a Registry to find hooks for each event
// and a Runner to execute them with timeout and aggregation.
type Engine struct {
	registry *Registry
	runner   *Runner
}

// NewEngine returns an Engine backed by the given registry and runner.
func NewEngine(r *Registry, runner *Runner) *Engine {
	return &Engine{registry: r, runner: runner}
}

// Fire invokes all hooks registered for the event in tier order, then
// aggregates their results. Returns the Aggregated.
func (e *Engine) Fire(ctx context.Context, event Event, payload Payload) (*Aggregated, error) {
	hooks := e.registry.ForEvent(event)
	if len(hooks) == 0 {
		return &Aggregated{Decision: DecisionNone}, nil
	}
	results := make([]Result, len(hooks))
	agg := &Aggregated{Decision: DecisionNone, HookCount: len(hooks)}
	var wg sync.WaitGroup
	for i, h := range hooks {
		wg.Add(1)
		go func(i int, h *Hook) {
			defer wg.Done()
			r, err := e.runner.Run(ctx, h, payload)
			results[i] = r
			if err != nil {
				// Best-effort: append to Errors. We don't want one
				// hook failure to break the whole event chain.
				mu.Lock()
				agg.Errors = append(agg.Errors, err)
				mu.Unlock()
			}
		}(i, h)
	}
	wg.Wait()
	// Merge in-aggregation state from results.
	merged := aggregate(results)
	merged.Errors = append(merged.Errors, agg.Errors...)
	return merged, nil
}

var mu sync.Mutex

func aggregate(results []Result) *Aggregated {
	agg := &Aggregated{Decision: DecisionNone, HookCount: len(results)}
	for _, r := range results {
		if r.Halt {
			agg.Halt = true
		}
		if r.Transformed != nil {
			agg.Transformed = r.Transformed
		}
		if r.Context != "" {
			if agg.Context != "" {
				agg.Context += "\n"
			}
			agg.Context += r.Context
		}
		// Deny wins over allow wins over none.
		switch r.Decision {
		case DecisionDeny:
			agg.Decision = DecisionDeny
			if r.Reason != "" {
				if agg.Reason != "" {
					agg.Reason += "\n"
				}
				agg.Reason += r.Reason
			}
		case DecisionAllow:
			if agg.Decision != DecisionDeny {
				agg.Decision = DecisionAllow
			}
		}
	}
	return agg
}
