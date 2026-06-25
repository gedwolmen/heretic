package hookengine

import (
	"context"
	"fmt"
	"sync"
)

// Hook is the contract every hook implementation must satisfy.
//
// A hook is a small, named function. It receives a context and a payload
// (the shape depends on the Event), and returns either a Decision (for
// guard-style hooks) or a Transformation (for transform-style hooks).
//
// The Result type covers both: if Decision is set, the engine aggregates
// it; if Transformed is non-nil, the engine applies it to the payload
// (or system prompt, etc.).
type Hook struct {
	// Name is the registry name. Must be unique.
	Name string
	// Tier is the hook's classification.
	Tier Tier
	// Events is the list of events this hook fires on.
	Events []Event
	// Fn is the hook function. It returns a Result.
	Fn func(ctx context.Context, p Payload) (Result, error)
}

// Payload is the common interface every event's payload implements.
// Concrete event types embed EventData and add their own fields.
type Payload interface {
	Event() Event
	SessionID() string
}

// Result is the outcome of a single hook invocation.
type Result struct {
	// Decision for guard hooks: allow/deny/none.
	Decision string
	// Reason for deny decisions.
	Reason string
	// Halt stops the whole turn.
	Halt bool
	// Transformed replaces the payload (for transform hooks). May be
	// the same type as the input payload, or a different one (e.g.
	// the system prompt after transformation).
	Transformed any
	// Context is appended to tool results (om-my-openagent semantics).
	Context string
	// Async marks the hook as fire-and-forget; the engine does not
	// wait for it to complete.
	Async bool
}

// Decision constants.
const (
	DecisionAllow = "allow"
	DecisionDeny  = "deny"
	DecisionNone  = "none"
)

// Registry is the hook catalog. Concurrent-safe.
type Registry struct {
	mu    sync.RWMutex
	hooks map[string]*Hook
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{hooks: make(map[string]*Hook)}
}

// Register adds a hook to the registry. Returns an error on duplicate
// name.
func (r *Registry) Register(h *Hook) error {
	if h.Name == "" {
		return fmt.Errorf("hookengine: hook name required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.hooks[h.Name]; ok {
		return fmt.Errorf("hookengine: hook %q already registered", h.Name)
	}
	r.hooks[h.Name] = h
	return nil
}

// MustRegister is like Register but panics on error. Use at package init.
func (r *Registry) MustRegister(h *Hook) {
	if err := r.Register(h); err != nil {
		panic(err)
	}
}

// Get returns the hook with the given name.
func (r *Registry) Get(name string) (*Hook, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.hooks[name]
	return h, ok
}

// ForEvent returns all hooks registered for the given event, ordered by
// tier (Session first, then ToolGuard, etc.).
func (r *Registry) ForEvent(e Event) []*Hook {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*Hook
	for _, h := range r.hooks {
		for _, he := range h.Events {
			if he == e {
				out = append(out, h)
				break
			}
		}
	}
	return out
}

// Names returns all registered hook names, sorted alphabetically.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.hooks))
	for n := range r.hooks {
		out = append(out, n)
	}
	return out
}
