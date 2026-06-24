// Package tools exposes the `task` tool to the LLM. It is a thin wrapper
// over the internal/delegate package that uses the coordinator's session
// infrastructure to spawn a real child session.
package tools

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"charm.land/fantasy"

	"github.com/gedwolmen/heretic/internal/delegate"
)

//go:embed task.md
var taskDescription string

// TaskToolName is the LLM-facing name of the task tool.
const TaskToolName = "task"

// TaskParams mirrors delegate.Params but uses the JSON shapes the LLM
// supplies.
type TaskParams struct {
	Category string `json:"category" description:"Routing key: explore | librarian | search | edit"`
	Prompt   string `json:"prompt" description:"The task for the subagent to perform"`
	Provider string `json:"provider,omitempty" description:"Optional provider override"`
	Model    string `json:"model,omitempty" description:"Optional model override"`
}

// TaskResponse is what we return to the parent agent.
type TaskResponse struct {
	SessionID string `json:"session_id"`
	Output    string `json:"output"`
}

// TaskSpawner is the contract the task tool needs from the coordinator.
type TaskSpawner interface {
	SpawnChild(ctx context.Context, req delegate.DelegateRequest) (delegate.DelegateResult, error)
	ProviderModel() (provider, model string)
}

// NewTaskTool returns the `task` tool wired with the provided spawner.
// concurrency controls the per-provider/model limit; pass 0 for the default.
func NewTaskTool(spawner TaskSpawner, concurrency int) fantasy.AgentTool {
	dt := delegate.NewToolWithLimits(spawner, concurrency)
	if concurrency <= 0 {
		dt.Limiter = delegate.NewLimiter(delegate.DefaultDelegateConcurrency)
	}
	return fantasy.NewAgentTool(
		TaskToolName,
		taskDescription,
		func(ctx context.Context, params TaskParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Prompt == "" {
				return fantasy.NewTextErrorResponse("prompt is required"), nil
			}
			provider, model := spawner.ProviderModel()
			if params.Provider != "" {
				provider = params.Provider
			}
			if params.Model != "" {
				model = params.Model
			}
			res, err := dt.Execute(ctx, delegate.Params{
				Category: delegate.Category(params.Category),
				Prompt:   params.Prompt,
				Provider: provider,
				Model:    model,
			})
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("delegate failed: %v", err)), nil
			}
			return fantasy.NewTextResponse(fmt.Sprintf("session_id=%s\n\n%s", res.SessionID, res.Output)), nil
		},
	)
}

// ErrNoSpawner is returned when no spawner is provided.
var ErrNoSpawner = errors.New("no task spawner configured")
