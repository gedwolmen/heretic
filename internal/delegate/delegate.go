// Package delegate implements subagent delegation for heretic.
//
// The package exposes a `task` tool that lets a parent agent spawn a child
// session to perform a discrete subtask. Each spawn is a fresh session with
// its own provider/model selection. Concurrency is bounded per provider/model
// so the LLM API is not overwhelmed with concurrent requests.
//
// Design notes:
//   - This is a Go-idiomatic concept port of oh-my-opencode's `task` tool
//     (which is in TypeScript and tightly coupled to the OpenCode plugin API).
//     We do NOT copy code; we reimplement the concept in Go against the
//     heretic agent infrastructure.
//   - The `agent` tool (internal/agent/agent_tool.go) is heretic's lower-level
//     escape hatch. The `delegate` tool is the high-level category-based
//     entry point and is what the LLM should reach for first.
//   - Concurrency is configurable via `delegate.concurrency` in heretic.json.
//     The default is 5, matching OmO's default.
package delegate

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

// DefaultDelegateConcurrency is the fallback per-provider/model concurrency
// limit when `delegate.concurrency` is not set in heretic.json.
const DefaultDelegateConcurrency = 5

// Category is a routing key that maps to a registered agent type.
type Category string

// Registry maps categories to agent-type names. Categories used in heretic
// are minimal; users can extend at registration time.
type Registry struct {
	mu         sync.RWMutex
	categories map[Category]string
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{categories: make(map[Category]string)}
}

// Register binds a category to an agent-type name.
func (r *Registry) Register(category Category, agentType string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.categories[category] = agentType
}

// Resolve returns the agent type for a category. Returns an error if the
// category is unknown.
func (r *Registry) Resolve(category Category) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	agentType, ok := r.categories[category]
	if !ok {
		return "", fmt.Errorf("unknown category %q", category)
	}
	return agentType, nil
}

// DefaultRegistry returns a registry populated with the standard heretic
// categories.
func DefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register("explore", "explore")
	r.Register("librarian", "librarian")
	r.Register("search", "search")
	r.Register("edit", "edit")
	return r
}

// Limiter bounds concurrent delegate spawns per provider/model.
type Limiter struct {
	mu    sync.Mutex
	slots map[string]chan struct{}
	max   int
}

// NewLimiter returns a Limiter that allows up to `max` concurrent spawns
// per provider/model key.
func NewLimiter(max int) *Limiter {
	if max < 1 {
		max = DefaultDelegateConcurrency
	}
	return &Limiter{
		slots: make(map[string]chan struct{}),
		max:   max,
	}
}

// slot returns the semaphore channel for a key, creating it on first use.
func (l *Limiter) slot(key string) chan struct{} {
	l.mu.Lock()
	defer l.mu.Unlock()
	s, ok := l.slots[key]
	if !ok {
		s = make(chan struct{}, l.max)
		l.slots[key] = s
	}
	return s
}

// Acquire blocks until a slot is available for the given key, or until ctx
// is canceled. Returns a release function that must be called to free the
// slot. If ctx is canceled before a slot becomes available, the returned
// release function is a no-op and the error is non-nil.
func (l *Limiter) Acquire(ctx context.Context, key string) (release func(), err error) {
	s := l.slot(key)
	select {
	case s <- struct{}{}:
		var released atomic.Bool
		return func() {
			if released.Swap(true) {
				return
			}
			<-s
		}, nil
	case <-ctx.Done():
		return func() {}, ctx.Err()
	}
}

// DelegateRequest describes a single subagent spawn.
type DelegateRequest struct {
	Category   Category
	Prompt     string
	ProviderID string
	ModelID    string
}

// DelegateResult is what the parent session receives from a spawn.
type DelegateResult struct {
	SessionID string
	Output    string
}

// Spawner is the contract the delegate tool depends on for actually creating
// a child session. The heretic coordinator implements this; tests use a fake.
type Spawner interface {
	SpawnChild(ctx context.Context, req DelegateRequest) (DelegateResult, error)
}

// Tool is the entry point exposed to the LLM as the `task` tool. It validates
// the request, resolves the category to an agent type, and acquires a
// concurrency slot before delegating to the Spawner.
type Tool struct {
	Registry *Registry
	Limiter  *Limiter
	Spawner  Spawner
}

// NewTool returns a Tool wired with the default registry and limiter.
func NewTool(spawner Spawner) *Tool {
	return &Tool{
		Registry: DefaultRegistry(),
		Limiter:  NewLimiter(DefaultDelegateConcurrency),
		Spawner:  spawner,
	}
}

// NewToolWithLimits returns a Tool with a custom concurrency limit.
func NewToolWithLimits(spawner Spawner, concurrency int) *Tool {
	return &Tool{
		Registry: DefaultRegistry(),
		Limiter:  NewLimiter(concurrency),
		Spawner:  spawner,
	}
}

// Params is the input shape the LLM provides.
type Params struct {
	Category Category `json:"category" description:"Routing key: explore | librarian | search | edit"`
	Prompt   string   `json:"prompt" description:"The task for the subagent to perform"`
	Provider string   `json:"provider,omitempty" description:"Provider ID override (defaults to parent's)"`
	Model    string   `json:"model,omitempty" description:"Model ID override (defaults to parent's)"`
}

// ErrEmptyPrompt is returned when the LLM calls the tool with no prompt.
var ErrEmptyPrompt = errors.New("prompt is required")

// Execute performs the delegation. It returns the child's session ID and
// output, or an error.
func (t *Tool) Execute(ctx context.Context, params Params) (DelegateResult, error) {
	if params.Prompt == "" {
		return DelegateResult{}, ErrEmptyPrompt
	}
	agentType, err := t.Registry.Resolve(params.Category)
	if err != nil {
		return DelegateResult{}, err
	}
	_ = agentType // currently informational; Spawner uses the request as-is

	provider := params.Provider
	model := params.Model
	key := provider + "/" + model
	release, err := t.Limiter.Acquire(ctx, key)
	if err != nil {
		return DelegateResult{}, fmt.Errorf("concurrency limit: %w", err)
	}
	defer release()

	return t.Spawner.SpawnChild(ctx, DelegateRequest{
		Category:   params.Category,
		Prompt:     params.Prompt,
		ProviderID: provider,
		ModelID:    model,
	})
}
