package config

// BackgroundAgentConfig configures background agent concurrency and queueing.
// Mirrors packages/omo-opencode/src/config/schema/background-agent.ts.
//
// Background agents are concurrent LLM calls running alongside the main
// session. They share a per-provider/model concurrency slot pool so a
// single provider's rate limit is not exceeded.
type BackgroundAgentConfig struct {
	// ModelConcurrency is the default per-provider/model concurrency cap.
	// Default: 5.
	ModelConcurrency int `json:"model_concurrency,omitempty" jsonschema:"description=Default concurrent background agents per provider/model pair,default=5,minimum=1"`

	// ProviderConcurrency is an optional override map keyed by provider ID.
	// A provider not in the map falls back to ModelConcurrency.
	ProviderConcurrency map[string]int `json:"provider_concurrency,omitempty" jsonschema:"description=Per-provider concurrency overrides"`

	// QueueOnFull controls what happens when a slot is full. true (default)
	// queues the request FIFO; false rejects with an error.
	QueueOnFull bool `json:"queue_on_full,omitempty" jsonschema:"description=Queue requests when concurrency limit is reached (vs reject),default=true"`

	// MaxQueueDepth is the max number of pending requests per slot pool.
	// 0 means unlimited.
	MaxQueueDepth int `json:"max_queue_depth,omitempty" jsonschema:"description=Max pending requests per slot pool (0=unlimited),default=0,minimum=0"`

	// ParentWakeTimeoutMs is how long the parent session waits for a wake
	// notification before re-polling. Default: 30000 (30s).
	ParentWakeTimeoutMs int `json:"parent_wake_timeout_ms,omitempty" jsonschema:"description=Parent wake poll interval in ms,default=30000,minimum=1000"`
}

// EffectiveModelConcurrency returns ModelConcurrency with default.
func (b BackgroundAgentConfig) EffectiveModelConcurrency() int {
	if b.ModelConcurrency < 1 {
		return 5
	}
	return b.ModelConcurrency
}

// EffectiveConcurrencyFor returns the configured concurrency for a provider
// ID, falling back to the model-level default.
func (b BackgroundAgentConfig) EffectiveConcurrencyFor(providerID string) int {
	if c, ok := b.ProviderConcurrency[providerID]; ok && c > 0 {
		return c
	}
	return b.EffectiveModelConcurrency()
}

// EffectiveParentWakeTimeoutMs returns the parent wake timeout with default.
func (b BackgroundAgentConfig) EffectiveParentWakeTimeoutMs() int {
	if b.ParentWakeTimeoutMs < 1000 {
		return 30000
	}
	return b.ParentWakeTimeoutMs
}
