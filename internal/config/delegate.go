package config

// DelegateConfig configures subagent delegation via the `task` tool.
type DelegateConfig struct {
	// Concurrency is the per-provider/model concurrency limit. Default: 5.
	Concurrency int `json:"concurrency,omitempty" jsonschema:"description=Per-provider/model concurrency limit,default=5,minimum=1"`

	// PerProvider overrides Concurrency on a per-provider basis.
	PerProvider map[string]int `json:"per_provider,omitempty" jsonschema:"description=Per-provider overrides"`
}

// EffectiveConcurrency returns the concurrency for a given provider ID,
// falling back to the model-level default.
func (d DelegateConfig) EffectiveConcurrency(providerID string) int {
	if c, ok := d.PerProvider[providerID]; ok && c > 0 {
		return c
	}
	if d.Concurrency < 1 {
		return 5
	}
	return d.Concurrency
}
