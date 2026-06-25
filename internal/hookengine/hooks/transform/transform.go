// Package transformhooks contains Transform-tier hooks.
package transformhooks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gedwolmen/heretic/internal/hookengine"
)

// ChatTransformData is the payload for chat.transform events.
type ChatTransformData struct {
	EventName    hookengine.Event `json:"-"`
	SessID       string           `json:"session_id"`
	SystemPrompt string           `json:"system_prompt"`
	Messages     []ChatMessage    `json:"messages"`
}

func (e *ChatTransformData) Event() hookengine.Event { return e.EventName }
func (e *ChatTransformData) SessionID() string       { return e.SessID }

// ChatMessage is a single message in the chat history.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DirectoryAgentsInjector scans the cwd for AGENTS.md, CLAUDE.md, etc.
// and prepends their content to the system prompt.
func DirectoryAgentsInjector(cwd string, filenames []string) *hookengine.Hook {
	if len(filenames) == 0 {
		filenames = []string{"AGENTS.md", "CLAUDE.md", "CRUSH.md", "HERETIC.md"}
	}
	return &hookengine.Hook{
		Name:   "directory-agents-injector",
		Tier:   hookengine.TierTransform,
		Events: []hookengine.Event{hookengine.EventChatTransform, hookengine.EventSessionStart},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*ChatTransformData)
			if !ok {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			var sb strings.Builder
			sb.WriteString(d.SystemPrompt)
			sb.WriteString("\n\n# Project context\n")
			for _, fn := range filenames {
				path := filepath.Join(cwd, fn)
				b, err := os.ReadFile(path)
				if err != nil {
					continue
				}
				sb.WriteString(fmt.Sprintf("\n## %s\n\n%s\n", fn, string(b)))
			}
			d.SystemPrompt = sb.String()
			return hookengine.Result{Decision: hookengine.DecisionNone, Transformed: d}, nil
		},
	}
}

// DirectoryReadmeInjector scans the cwd for README.md and appends its
// content to the system prompt.
func DirectoryReadmeInjector(cwd string) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "directory-readme-injector",
		Tier:   hookengine.TierTransform,
		Events: []hookengine.Event{hookengine.EventChatTransform, hookengine.EventSessionStart},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*ChatTransformData)
			if !ok {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			b, err := os.ReadFile(filepath.Join(cwd, "README.md"))
			if err != nil {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			d.SystemPrompt = d.SystemPrompt + "\n\n# README\n\n" + string(b)
			return hookengine.Result{Decision: hookengine.DecisionNone, Transformed: d}, nil
		},
	}
}

// CategorySkillReminder reminds the agent which skills are available
// for the current category.
func CategorySkillReminder(category string, skills []string) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "category-skill-reminder",
		Tier:   hookengine.TierTransform,
		Events: []hookengine.Event{hookengine.EventChatTransform},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*ChatTransformData)
			if !ok {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			if len(skills) == 0 {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			d.SystemPrompt = d.SystemPrompt + "\n\n# Skills for category " + category + "\n\n" + strings.Join(skills, "\n")
			return hookengine.Result{Decision: hookengine.DecisionNone, Transformed: d}, nil
		},
	}
}

// MonitorStatusInjector injects system status (CPU, memory, etc.) into
// the system prompt.
func MonitorStatusInjector(status string) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "monitor-status-injector",
		Tier:   hookengine.TierTransform,
		Events: []hookengine.Event{hookengine.EventChatTransform},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*ChatTransformData)
			if !ok {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			d.SystemPrompt = d.SystemPrompt + "\n\n# System status\n\n" + status
			return hookengine.Result{Decision: hookengine.DecisionNone, Transformed: d}, nil
		},
	}
}

// HephaestusAgentsMdInjector forces the hephaestus agent to always
// include the project's AGENTS.md content.
func HephaestusAgentsMdInjector(agentName, cwd string) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "hephaestus-agents-md-injector",
		Tier:   hookengine.TierTransform,
		Events: []hookengine.Event{hookengine.EventChatTransform},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			if agentName != "hephaestus" {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			d, ok := p.(*ChatTransformData)
			if !ok {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			b, err := os.ReadFile(filepath.Join(cwd, "AGENTS.md"))
			if err != nil {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			d.SystemPrompt = d.SystemPrompt + "\n\n# Project AGENTS.md (mandatory for hephaestus)\n\n" + string(b)
			return hookengine.Result{Decision: hookengine.DecisionNone, Transformed: d}, nil
		},
	}
}
