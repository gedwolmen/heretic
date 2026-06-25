// Package teamtools contains the 12 LLM-facing team tools.
//
//	team_create
//	team_delete
//	team_shutdown_request
//	team_approve_shutdown
//	team_reject_shutdown
//	team_send_message
//	team_task_create
//	team_task_list
//	team_task_update
//	team_task_get
//	team_status
//	team_list
//
// Each tool is a thin wrapper over the team primitives (storage,
// mailbox, tasklist, worktree).
package teamtools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"charm.land/fantasy"

	"github.com/gedwolmen/heretic/internal/team"
)

// teamBackend is the contract the tools need. The caller (coordinator)
// supplies a concrete implementation backed by team.Storage and a
// in-memory tasklist; tests can supply a fake.
type teamBackend interface {
	teamCreate(ctx context.Context, cfg team.Config) (teamResult, error)
	teamDelete(ctx context.Context, name string) (teamResult, error)
	teamShutdownRequest(ctx context.Context, name, fromMember string) (teamResult, error)
	teamApproveShutdown(ctx context.Context, name, fromMember string) (teamResult, error)
	teamRejectShutdown(ctx context.Context, name, fromMember, reason string) (teamResult, error)
	teamSendMessage(ctx context.Context, name string, msg team.Message) (teamResult, error)
	teamTaskCreate(ctx context.Context, name, title, owner string) (string, error)
	teamTaskList(ctx context.Context, name string) ([]teamTask, error)
	teamTaskUpdate(ctx context.Context, name, taskID, status string) (teamResult, error)
	teamTaskGet(ctx context.Context, name, taskID string) (teamTask, error)
	teamStatus(ctx context.Context, name string) (team.Status, error)
	teamList(ctx context.Context) ([]string, error)
}

type teamResult struct {
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
}

type teamTask struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Owner     string    `json:"owner,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// DefaultBackend is a simple in-memory + on-disk backend that the
// coordinator wires up. Tests use NewFakeBackend.
type DefaultBackend struct {
	Storage *team.Storage
	mu      sync.Mutex
	tasks   map[string][]teamTask // keyed by team name
}

func NewDefaultBackend(storage *team.Storage) *DefaultBackend {
	return &DefaultBackend{Storage: storage, tasks: make(map[string][]teamTask)}
}

func (b *DefaultBackend) teamCreate(ctx context.Context, cfg team.Config) (teamResult, error) {
	if err := b.Storage.Create(cfg); err != nil {
		return teamResult{OK: false, Detail: err.Error()}, nil
	}
	return teamResult{OK: true, Detail: "team created"}, nil
}

func (b *DefaultBackend) teamDelete(ctx context.Context, name string) (teamResult, error) {
	if err := b.Storage.DeleteTeam(name); err != nil {
		return teamResult{OK: false, Detail: err.Error()}, nil
	}
	return teamResult{OK: true, Detail: "team deleted"}, nil
}

func (b *DefaultBackend) teamShutdownRequest(ctx context.Context, name, fromMember string) (teamResult, error) {
	st, err := b.Storage.LoadState(name)
	if err != nil {
		return teamResult{OK: false, Detail: err.Error()}, nil
	}
	st.Status = team.StatusShutdown
	if err := b.Storage.SaveState(name, st); err != nil {
		return teamResult{OK: false, Detail: err.Error()}, nil
	}
	return teamResult{OK: true, Detail: "shutdown requested by " + fromMember}, nil
}

func (b *DefaultBackend) teamApproveShutdown(ctx context.Context, name, fromMember string) (teamResult, error) {
	if err := b.Storage.DeleteTeam(name); err != nil {
		return teamResult{OK: false, Detail: err.Error()}, nil
	}
	return teamResult{OK: true, Detail: "shutdown approved by " + fromMember}, nil
}

func (b *DefaultBackend) teamRejectShutdown(ctx context.Context, name, fromMember, reason string) (teamResult, error) {
	st, err := b.Storage.LoadState(name)
	if err != nil {
		return teamResult{OK: false, Detail: err.Error()}, nil
	}
	st.Status = team.StatusActive
	if err := b.Storage.SaveState(name, st); err != nil {
		return teamResult{OK: false, Detail: err.Error()}, nil
	}
	return teamResult{OK: true, Detail: "shutdown rejected: " + reason}, nil
}

func (b *DefaultBackend) teamSendMessage(ctx context.Context, name string, msg team.Message) (teamResult, error) {
	cfg, err := b.Storage.LoadConfig(name)
	if err != nil {
		return teamResult{OK: false, Detail: err.Error()}, nil
	}
	mb := team.NewMailbox(b.Storage.TeamDir(name))
	if err := mb.Send(msg); err != nil {
		return teamResult{OK: false, Detail: err.Error()}, nil
	}
	_ = cfg
	return teamResult{OK: true, Detail: "message sent"}, nil
}

func (b *DefaultBackend) teamTaskCreate(ctx context.Context, name, title, owner string) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	id := fmt.Sprintf("t-%d", time.Now().UnixNano())
	b.tasks[name] = append(b.tasks[name], teamTask{
		ID: id, Title: title, Owner: owner, Status: "pending", CreatedAt: time.Now(),
	})
	return id, nil
}

func (b *DefaultBackend) teamTaskList(ctx context.Context, name string) ([]teamTask, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := append([]teamTask{}, b.tasks[name]...)
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (b *DefaultBackend) teamTaskUpdate(ctx context.Context, name, taskID, status string) (teamResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, t := range b.tasks[name] {
		if t.ID == taskID {
			b.tasks[name][i].Status = status
			return teamResult{OK: true, Detail: "task updated"}, nil
		}
	}
	return teamResult{OK: false, Detail: "task not found"}, nil
}

func (b *DefaultBackend) teamTaskGet(ctx context.Context, name, taskID string) (teamTask, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, t := range b.tasks[name] {
		if t.ID == taskID {
			return t, nil
		}
	}
	return teamTask{}, fmt.Errorf("task not found: %s", taskID)
}

func (b *DefaultBackend) teamStatus(ctx context.Context, name string) (team.Status, error) {
	st, err := b.Storage.LoadState(name)
	if err != nil {
		return "", err
	}
	return st.Status, nil
}

func (b *DefaultBackend) teamList(ctx context.Context) ([]string, error) {
	return b.Storage.ListTeams()
}

// Compile-time check that DefaultBackend implements teamBackend.
var _ teamBackend = (*DefaultBackend)(nil)

// NewTools returns the 12 team tools backed by the given backend.
func NewTools(b teamBackend) []fantasy.AgentTool {
	tools := []struct {
		name string
		desc string
		fn   func(ctx context.Context, params json.RawMessage, call fantasy.ToolCall) (fantasy.ToolResponse, error)
	}{
		{
			name: "team_create",
			desc: "Create a new team. Requires: name, members[].",
			fn: func(ctx context.Context, params json.RawMessage, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
				var p struct {
					Name    string        `json:"name"`
					Owner   string        `json:"owner"`
					Members []team.Member `json:"members"`
				}
				if err := json.Unmarshal(params, &p); err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				res, err := b.teamCreate(ctx, team.Config{Name: p.Name, Owner: p.Owner, Members: p.Members})
				if err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				return fantasy.NewTextResponse(formatResult(res)), nil
			},
		},
		{
			name: "team_delete",
			desc: "Delete a team. Requires: name.",
			fn: func(ctx context.Context, params json.RawMessage, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
				var p struct {
					Name string `json:"name"`
				}
				if err := json.Unmarshal(params, &p); err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				res, err := b.teamDelete(ctx, p.Name)
				if err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				return fantasy.NewTextResponse(formatResult(res)), nil
			},
		},
		{
			name: "team_shutdown_request",
			desc: "Request a team shutdown. Requires: name, from_member.",
			fn: func(ctx context.Context, params json.RawMessage, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
				var p struct {
					Name       string `json:"name"`
					FromMember string `json:"from_member"`
				}
				if err := json.Unmarshal(params, &p); err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				res, err := b.teamShutdownRequest(ctx, p.Name, p.FromMember)
				if err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				return fantasy.NewTextResponse(formatResult(res)), nil
			},
		},
		{
			name: "team_approve_shutdown",
			desc: "Approve a pending shutdown. Requires: name, from_member.",
			fn: func(ctx context.Context, params json.RawMessage, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
				var p struct {
					Name       string `json:"name"`
					FromMember string `json:"from_member"`
				}
				if err := json.Unmarshal(params, &p); err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				res, err := b.teamApproveShutdown(ctx, p.Name, p.FromMember)
				if err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				return fantasy.NewTextResponse(formatResult(res)), nil
			},
		},
		{
			name: "team_reject_shutdown",
			desc: "Reject a shutdown. Requires: name, from_member, reason.",
			fn: func(ctx context.Context, params json.RawMessage, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
				var p struct {
					Name       string `json:"name"`
					FromMember string `json:"from_member"`
					Reason     string `json:"reason"`
				}
				if err := json.Unmarshal(params, &p); err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				res, err := b.teamRejectShutdown(ctx, p.Name, p.FromMember, p.Reason)
				if err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				return fantasy.NewTextResponse(formatResult(res)), nil
			},
		},
		{
			name: "team_send_message",
			desc: "Send a message to a team member. Requires: name, from, to, body.",
			fn: func(ctx context.Context, params json.RawMessage, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
				var p struct {
					Name string `json:"name"`
					From string `json:"from"`
					To   string `json:"to"`
					Body string `json:"body"`
				}
				if err := json.Unmarshal(params, &p); err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				res, err := b.teamSendMessage(ctx, p.Name, team.Message{From: p.From, To: p.To, Body: p.Body})
				if err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				return fantasy.NewTextResponse(formatResult(res)), nil
			},
		},
		{
			name: "team_task_create",
			desc: "Create a task in the team's tasklist. Requires: name, title, owner.",
			fn: func(ctx context.Context, params json.RawMessage, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
				var p struct {
					Name  string `json:"name"`
					Title string `json:"title"`
					Owner string `json:"owner"`
				}
				if err := json.Unmarshal(params, &p); err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				id, err := b.teamTaskCreate(ctx, p.Name, p.Title, p.Owner)
				if err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				return fantasy.NewTextResponse("task_id=" + id), nil
			},
		},
		{
			name: "team_task_list",
			desc: "List the team's tasks. Requires: name.",
			fn: func(ctx context.Context, params json.RawMessage, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
				var p struct {
					Name string `json:"name"`
				}
				if err := json.Unmarshal(params, &p); err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				tasks, err := b.teamTaskList(ctx, p.Name)
				if err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				return fantasy.NewTextResponse(formatTasks(tasks)), nil
			},
		},
		{
			name: "team_task_update",
			desc: "Update a task's status. Requires: name, task_id, status.",
			fn: func(ctx context.Context, params json.RawMessage, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
				var p struct {
					Name   string `json:"name"`
					TaskID string `json:"task_id"`
					Status string `json:"status"`
				}
				if err := json.Unmarshal(params, &p); err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				res, err := b.teamTaskUpdate(ctx, p.Name, p.TaskID, p.Status)
				if err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				return fantasy.NewTextResponse(formatResult(res)), nil
			},
		},
		{
			name: "team_task_get",
			desc: "Fetch a single task by id. Requires: name, task_id.",
			fn: func(ctx context.Context, params json.RawMessage, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
				var p struct {
					Name   string `json:"name"`
					TaskID string `json:"task_id"`
				}
				if err := json.Unmarshal(params, &p); err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				t, err := b.teamTaskGet(ctx, p.Name, p.TaskID)
				if err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				return fantasy.NewTextResponse(formatTasks([]teamTask{t})), nil
			},
		},
		{
			name: "team_status",
			desc: "Get the team's runtime status. Requires: name.",
			fn: func(ctx context.Context, params json.RawMessage, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
				var p struct {
					Name string `json:"name"`
				}
				if err := json.Unmarshal(params, &p); err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				st, err := b.teamStatus(ctx, p.Name)
				if err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				return fantasy.NewTextResponse("status=" + string(st)), nil
			},
		},
		{
			name: "team_list",
			desc: "List all teams.",
			fn: func(ctx context.Context, params json.RawMessage, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
				teams, err := b.teamList(ctx)
				if err != nil {
					return fantasy.NewTextErrorResponse(err.Error()), nil
				}
				return fantasy.NewTextResponse(strings.Join(teams, "\n")), nil
			},
		},
	}
	out := make([]fantasy.AgentTool, 0, len(tools))
	for _, t := range tools {
		t := t // capture
		out = append(out, fantasy.NewAgentTool(
			t.name,
			t.desc,
			func(ctx context.Context, params json.RawMessage, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
				return t.fn(ctx, params, call)
			},
		))
	}
	return out
}

// formatResult is a tiny formatter to avoid pulling fmt/strings in.
func formatResult(r teamResult) string {
	if r.OK {
		return "ok: " + r.Detail
	}
	return "error: " + r.Detail
}

func formatTasks(ts []teamTask) string {
	if len(ts) == 0 {
		return "(no tasks)"
	}
	out := ""
	for _, t := range ts {
		out += t.ID + "\t" + t.Status + "\t" + t.Title + "\n"
	}
	return out
}

// strings is imported in the file via team_list but Go requires it; we
// add a minimal alias to satisfy the linter.
var _ = json.Marshal
var _ = fmt.Sprintf
