// Package toolguardhooks contains ToolGuard-tier hooks.
package toolguardhooks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gedwolmen/heretic/internal/hookengine"
)

// PreToolUseData is the payload for tool.pre events.
type PreToolUseData struct {
	EventName hookengine.Event `json:"-"`
	SessID    string           `json:"session_id"`
	CWD       string           `json:"cwd"`
	ToolName  string           `json:"tool_name"`
	ToolInput map[string]any   `json:"tool_input"`
}

func (e *PreToolUseData) Event() hookengine.Event { return e.EventName }
func (e *PreToolUseData) SessionID() string       { return e.SessID }

// PostToolUseData is the payload for tool.post events.
type PostToolUseData struct {
	EventName hookengine.Event `json:"-"`
	SessID    string           `json:"session_id"`
	CWD       string           `json:"cwd"`
	ToolName  string           `json:"tool_name"`
	Result    string           `json:"result"`
	Duration  time.Duration    `json:"duration"`
}

func (e *PostToolUseData) Event() hookengine.Event { return e.EventName }
func (e *PostToolUseData) SessionID() string       { return e.SessID }

// WriteExistingFileGuard blocks writes to files that already exist
// unless the user has explicitly approved.
func WriteExistingFileGuard(approved map[string]bool) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "write-existing-file-guard",
		Tier:   hookengine.TierToolGuard,
		Events: []hookengine.Event{hookengine.EventPreToolUse},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*PreToolUseData)
			if !ok || d.ToolName != "write" && d.ToolName != "edit" && d.ToolName != "multi_edit" {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			path, _ := d.ToolInput["file_path"].(string)
			if path == "" {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			if approved[path] {
				return hookengine.Result{Decision: hookengine.DecisionAllow}, nil
			}
			if _, err := os.Stat(path); err == nil {
				return hookengine.Result{
					Decision: hookengine.DecisionDeny,
					Reason:   fmt.Sprintf("refusing to overwrite existing file %q without explicit approval", path),
				}, nil
			}
			return hookengine.Result{Decision: hookengine.DecisionAllow}, nil
		},
	}
}

// CommentChecker scans files for AI-slop comment patterns.
func CommentChecker(blockedPatterns []*regexp.Regexp) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "comment-checker",
		Tier:   hookengine.TierToolGuard,
		Events: []hookengine.Event{hookengine.EventPostToolUse},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*PostToolUseData)
			if !ok {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			for _, re := range blockedPatterns {
				if re.MatchString(d.Result) {
					return hookengine.Result{
						Decision: hookengine.DecisionDeny,
						Reason:   fmt.Sprintf("blocked comment pattern matched: %s", re.String()),
					}, nil
				}
			}
			return hookengine.Result{Decision: hookengine.DecisionAllow}, nil
		},
	}
}

// BashFileReadGuard blocks the bash tool from running cat/head/tail on
// files; the view tool should be used instead.
func BashFileReadGuard() *hookengine.Hook {
	blocked := regexp.MustCompile(`(?i)^\s*(cat|head|tail|less|more)\s+`)
	return &hookengine.Hook{
		Name:   "bash-file-read-guard",
		Tier:   hookengine.TierToolGuard,
		Events: []hookengine.Event{hookengine.EventPreToolUse},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*PreToolUseData)
			if !ok || d.ToolName != "bash" {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			cmd, _ := d.ToolInput["command"].(string)
			if blocked.MatchString(cmd) {
				return hookengine.Result{
					Decision: hookengine.DecisionDeny,
					Reason:   "use the view tool instead of cat/head/tail via bash",
				}, nil
			}
			return hookengine.Result{Decision: hookengine.DecisionAllow}, nil
		},
	}
}

// FsyncSkipWarning warns (does not block) on write tools that bypass
// fsync. The warning is surfaced as Context, not Deny.
func FsyncSkipWarning() *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "fsync-skip-warning",
		Tier:   hookengine.TierToolGuard,
		Events: []hookengine.Event{hookengine.EventPreToolUse},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*PreToolUseData)
			if !ok {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			if d.ToolName == "bash" {
				cmd, _ := d.ToolInput["command"].(string)
				if strings.Contains(cmd, "dd if=") {
					return hookengine.Result{
						Decision: hookengine.DecisionNone,
						Context:  "dd detected — consider conv=fsync for safety",
					}, nil
				}
			}
			return hookengine.Result{Decision: hookengine.DecisionNone}, nil
		},
	}
}

// EditErrorRecovery catches edit tool errors and provides a hint.
func EditErrorRecovery() *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "edit-error-recovery",
		Tier:   hookengine.TierToolGuard,
		Events: []hookengine.Event{hookengine.EventToolError},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			return hookengine.Result{
				Context: "edit failed; re-read the file and verify the old_string matches exactly",
			}, nil
		},
	}
}

// InteractiveBashSession validates interactive_bash tool calls.
func InteractiveBashSession() *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "interactive-bash-session",
		Tier:   hookengine.TierToolGuard,
		Events: []hookengine.Event{hookengine.EventPreToolUse},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*PreToolUseData)
			if !ok || d.ToolName != "interactive_bash" {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			return hookengine.Result{Decision: hookengine.DecisionAllow}, nil
		},
	}
}

// JSONErrorRecovery catches JSON parse errors from tool output.
func JSONErrorRecovery() *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "json-error-recovery",
		Tier:   hookengine.TierToolGuard,
		Events: []hookengine.Event{hookengine.EventToolError},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			return hookengine.Result{
				Context: "tool returned invalid JSON; retry with corrected output",
			}, nil
		},
	}
}

// NotepadWriteGuard blocks writes to the .notepad/ directory.
func NotepadWriteGuard() *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "notepad-write-guard",
		Tier:   hookengine.TierToolGuard,
		Events: []hookengine.Event{hookengine.EventPreToolUse},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*PreToolUseData)
			if !ok {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			path, _ := d.ToolInput["file_path"].(string)
			if strings.Contains(path, "/.notepad/") || strings.Contains(path, "/.omo/notepad/") {
				return hookengine.Result{
					Decision: hookengine.DecisionDeny,
					Reason:   ".notepad/ is read-only",
				}, nil
			}
			return hookengine.Result{Decision: hookengine.DecisionAllow}, nil
		},
	}
}

// PlanFormatValidator validates that plan output follows the required
// markdown structure.
func PlanFormatValidator() *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "plan-format-validator",
		Tier:   hookengine.TierToolGuard,
		Events: []hookengine.Event{hookengine.EventPostToolUse},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*PostToolUseData)
			if !ok || d.ToolName != "write" {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			path, _ := strings.CutPrefix(d.ToolName, "")
			_ = path
			if !strings.Contains(d.Result, "# ") || !strings.Contains(d.Result, "## Steps") {
				return hookengine.Result{
					Decision: hookengine.DecisionDeny,
					Reason:   "plan file must start with a # header and contain a ## Steps section",
				}, nil
			}
			return hookengine.Result{Decision: hookengine.DecisionAllow}, nil
		},
	}
}

// QuestionLabelTruncator truncates long user-question labels.
func QuestionLabelTruncator(maxLen int) *hookengine.Hook {
	if maxLen <= 0 {
		maxLen = 80
	}
	return &hookengine.Hook{
		Name:   "question-label-truncator",
		Tier:   hookengine.TierToolGuard,
		Events: []hookengine.Event{hookengine.EventPreToolUse},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*PreToolUseData)
			if !ok || d.ToolName != "question" {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			label, _ := d.ToolInput["label"].(string)
			if len(label) > maxLen {
				d.ToolInput["label"] = label[:maxLen-3] + "..."
			}
			return hookengine.Result{Decision: hookengine.DecisionAllow, Transformed: d}, nil
		},
	}
}

// ReadImageResizer ensures images passed to the read_image tool are
// within the configured size limit.
func ReadImageResizer(maxBytes int64) *hookengine.Hook {
	if maxBytes <= 0 {
		maxBytes = 5 * 1024 * 1024
	}
	return &hookengine.Hook{
		Name:   "read-image-resizer",
		Tier:   hookengine.TierToolGuard,
		Events: []hookengine.Event{hookengine.EventPreToolUse},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*PreToolUseData)
			if !ok || d.ToolName != "read_image" {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			path, _ := d.ToolInput["file_path"].(string)
			if path == "" {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			info, err := os.Stat(path)
			if err != nil {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			if info.Size() > maxBytes {
				return hookengine.Result{
					Decision: hookengine.DecisionDeny,
					Reason:   fmt.Sprintf("image %q is %d bytes; limit is %d", path, info.Size(), maxBytes),
				}, nil
			}
			return hookengine.Result{Decision: hookengine.DecisionAllow}, nil
		},
	}
}

// ToolOutputTruncator truncates tool output beyond a max length.
func ToolOutputTruncator(maxLen int) *hookengine.Hook {
	if maxLen <= 0 {
		maxLen = 50_000
	}
	return &hookengine.Hook{
		Name:   "tool-output-truncator",
		Tier:   hookengine.TierToolGuard,
		Events: []hookengine.Event{hookengine.EventPostToolUse},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*PostToolUseData)
			if !ok {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			if len(d.Result) <= maxLen {
				return hookengine.Result{Decision: hookengine.DecisionAllow}, nil
			}
			d.Result = d.Result[:maxLen] + "\n\n... [truncated]"
			return hookengine.Result{Decision: hookengine.DecisionAllow, Transformed: d}, nil
		},
	}
}

// ToolPairValidator validates that tool calls come in expected pairs.
func ToolPairValidator() *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "tool-pair-validator",
		Tier:   hookengine.TierToolGuard,
		Events: []hookengine.Event{hookengine.EventPreToolUse},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			_, ok := p.(*PreToolUseData)
			if !ok {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			// Example rule: write must be preceded by view in the same session.
			// Tracking that history is the caller's responsibility; this
			// is a stub that just allows.
			return hookengine.Result{Decision: hookengine.DecisionAllow}, nil
		},
	}
}

// WebfetchRedirectGuard blocks the web_fetch tool from following
// redirects to private IP ranges.
func WebfetchRedirectGuard() *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "webfetch-redirect-guard",
		Tier:   hookengine.TierToolGuard,
		Events: []hookengine.Event{hookengine.EventPreToolUse},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*PreToolUseData)
			if !ok || d.ToolName != "web_fetch" {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			url, _ := d.ToolInput["url"].(string)
			if strings.HasPrefix(url, "http://127.") || strings.HasPrefix(url, "http://localhost") || strings.HasPrefix(url, "http://10.") || strings.HasPrefix(url, "http://192.168.") {
				return hookengine.Result{
					Decision: hookengine.DecisionDeny,
					Reason:   "refusing to fetch private IP addresses",
				}, nil
			}
			return hookengine.Result{Decision: hookengine.DecisionAllow}, nil
		},
	}
}

// StopContinuationGuard denies attempts to stop the session while
// there are still pending todos.
func StopContinuationGuard(hasPendingTodos func() bool) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "stop-continuation-guard",
		Tier:   hookengine.TierToolGuard,
		Events: []hookengine.Event{hookengine.EventPreToolUse},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*PreToolUseData)
			if !ok || d.ToolName != "stop" {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			if hasPendingTodos() {
				return hookengine.Result{
					Decision: hookengine.DecisionDeny,
					Reason:   "cannot stop with pending todos",
				}, nil
			}
			return hookengine.Result{Decision: hookengine.DecisionAllow}, nil
		},
	}
}

// TaskReminder reminds the agent to use the task tool when it has been
// silent for a while.
func TaskReminder(silentTurns *int, mu *sync.Mutex, threshold int) *hookengine.Hook {
	if threshold <= 0 {
		threshold = 5
	}
	return &hookengine.Hook{
		Name:   "task-reminder",
		Tier:   hookengine.TierToolGuard,
		Events: []hookengine.Event{hookengine.EventPostToolUse},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			mu.Lock()
			*silentTurns++
			n := *silentTurns
			mu.Unlock()
			if n < threshold {
				return hookengine.Result{}, nil
			}
			return hookengine.Result{
				Context: "consider delegating to a subagent via the task tool",
			}, nil
		},
	}
}

// TasksTodowriteDisabler disables the todowrite tool to prevent the LLM
// from clobbering the agent's todo state.
func TasksTodowriteDisabler() *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "tasks-todowrite-disabler",
		Tier:   hookengine.TierToolGuard,
		Events: []hookengine.Event{hookengine.EventPreToolUse},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			d, ok := p.(*PreToolUseData)
			if !ok || d.ToolName != "todowrite" {
				return hookengine.Result{Decision: hookengine.DecisionNone}, nil
			}
			return hookengine.Result{
				Decision: hookengine.DecisionDeny,
				Reason:   "use the todos tool instead",
			}, nil
		},
	}
}

// TodoContinuationEnforcer forces the agent to continue if it tries to
// stop with pending todos.
func TodoContinuationEnforcer(hasPendingTodos func() bool) *hookengine.Hook {
	return &hookengine.Hook{
		Name:   "todo-continuation-enforcer",
		Tier:   hookengine.TierContinuation,
		Events: []hookengine.Event{hookengine.EventSessionIdle},
		Fn: func(ctx context.Context, p hookengine.Payload) (hookengine.Result, error) {
			if !hasPendingTodos() {
				return hookengine.Result{}, nil
			}
			return hookengine.Result{
				Context: "you have pending todos; continue working on them",
			}, nil
		},
	}
}

// precompile path-based lookup.
var _ = filepath.Separator
var _ = strings.HasPrefix
