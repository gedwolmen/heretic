package builtin

import (
	_ "embed"
)

//go:embed prompts/prometheus.md
var prometheusPrompt string

// NewPrometheusAgent returns the Prometheus agent: plan generator.
// Prometheus is special: it can only edit .md files (per the
// `prometheus-md-only` hook). It produces a plan, not code.
func NewPrometheusAgent() Agent {
	enabled := false
	return &baseAgent{
		name:         NamePrometheus,
		display:      "Prometheus",
		mode:         ModePrimary,
		modelPref:    "large",
		systemPrompt: prometheusPrompt,
		description:  "Plan generator. Reads a request, writes a detailed plan to plan.md. Cannot edit code.",
		allowedTools: []ToolRef{
			{Name: "view"},
			{Name: "glob"},
			{Name: "grep"},
			{Name: "write", Enabled: &enabled}, // md only, enforced by hook
			{Name: "edit", Enabled: &enabled},   // md only, enforced by hook
		},
	}
}
