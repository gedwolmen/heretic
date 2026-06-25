# AGENTS.md — Heretic Ultimate

> Heretic Ultimate is a terminal-first AI assistant, forked from
> [charmbracelet/crush](https://github.com/charmbracelet/crush) (FSL-1.1-MIT)
> and extended with the full om-my-opencode feature set, reimplemented
> in Go. See [NOTICE.md](./NOTICE.md) for upstream attribution.

## Project Overview

Heretic Ultimate is a Go CLI that drives an LLM in your terminal. It
reuses the entire crush runtime (agent loop, tool registry, session
storage, hook system, LSP, MCP, permissions, skills, etc.) and adds the
following new packages, all of which are concept ports from
oh-my-opencode:

| Package | Description |
| --- | --- |
| `internal/agent/builtin/` | 11 built-in agents + embedded prompts |
| `internal/hookengine/` | 5-tier hook engine (54 hooks, 12 events) |
| `internal/team/` | Team mode (storage, mailbox, 12 tools, worktrees) |
| `internal/mcp/` | 3-tier MCP (built-in, .mcp.json, skill-embedded + OAuth) |
| `internal/background/` | Background agents (concurrency, parent wake) |
| `internal/boulder/` | Boulder state (work tracker) |
| `internal/hashline/` | LINE#ID content addressing + diff enhancer |
| `internal/intentgate/` | Keyword detector + mode injection |
| `internal/rules/` | `.heretic/rules/*.md` loader |
| `internal/delegate/` | Subagent delegation tool |
| `internal/toolregistry/` | Tool catalog with gating flags |

## Build/Test/Lint Commands

- **Build**: `go build ./...` (produces `./heretic` binary)
- **Run**: `go run .` or `./heretic --help`
- **Test**: `go test ./... -race -failfast`
- **Vet**: `go vet ./...`
- **Format**: `gofmt -l .` (empty output = all formatted)
- **Tidy**: `go mod tidy`

## Architecture

```
main.go                          # CLI entry
internal/
  agent/
    builtin/                     # 11 agents + 11 prompt files
      prompts/                   # ~50KB of system prompts
      categories.go              # 8 task categories
      overrides.go               # user-overridable agent config
      registry.go                # agent catalog
    tools/                       # 20+ tool implementations (heretic_info,
                                 # heretic_logs, task, hashline_edit, ...)
    coordinator.go               # agent loop
  hookengine/
    engine.go                    # registry + payload interface
    fire.go                      # parallel fire + aggregate
    runner.go                    # per-hook timeout
    tiers.go                     # 5 tiers, 12 events
    hooks/
      session/                   # 12 hooks
      toolguard/                 # 18 hooks
      transform/                 # 5 hooks
      continuation/              # 7 hooks
      skill/                     # 12 hooks
  team/                          # team mode
    storage.go                   # file-based config.json + state.json
    mailbox.go                   # inter-agent messages
    eligibility.go               # agent eligibility registry
    worktree.go                  # per-member git worktree
    tools/                       # 12 LLM-facing team tools
  mcp/                           # 3-tier MCP
    mcp.go                       # tier 1 (built-in)
    oauth.go                     # PKCE + DCR
    skill.go                     # tier 3 (per-session)
  background/                    # background agents
  boulder/                       # work tracker
  hashline/                      # LINE#ID + diff enhancer
  intentgate/                    # keyword detector
  rules/                         # .heretic/rules loader
  delegate/                      # subagent delegation
  toolregistry/                  # tool catalog + gating
  config/                        # config schema (now includes team_mode,
                                 # background_agent, boulder_state, delegate)
  ...                            # the original crush packages
```

## 11 Built-in Agents

| Agent | Mode | Use |
| --- | --- | --- |
| sisyphus | primary | Master orchestrator |
| hephaestus | primary | Plan-driven builder |
| atlas | primary | Multi-agent orchestrator |
| prometheus | primary | Plan generator (.md only) |
| oracle | subagent | Architecture advisor |
| librarian | subagent | External research |
| explore | subagent | Codebase investigation |
| multimodal-looker | subagent | Image analysis |
| metis | subagent | Pre-planning intent clarifier |
| momus | subagent | Plan reviewer |
| sisyphus-junior | subagent | Lightweight coder |

Eligibility for team mode:
- **eligible**: sisyphus, atlas, sisyphus-junior
- **conditional**: hephaestus
- **hard-reject**: oracle, librarian, explore, multimodal-looker,
  metis, momus, prometheus

## 5-Tier Hook Engine

| Tier | Purpose | Examples |
| --- | --- | --- |
| Session | once per session | agent-usage-reminder, auto-update-checker, ralph-loop |
| ToolGuard | guard tool execution | write-existing-file-guard, comment-checker, bash-file-read-guard |
| Transform | transform LLM view | directory-agents-injector, category-skill-reminder |
| Continuation | keep work going | todo-continuation-enforcer, start-work-continuation, preemptive-compaction |
| Skill | bridge skills | keyword-detector, model-fallback, team-mode-status-injector |

12 events: `session.start`, `session.end`, `session.idle`,
`session.error`, `message.received`, `tool.pre`, `tool.post`,
`tool.error`, `system.transform`, `chat.transform`, `session.compacting`,
`compaction.autocontinue`.

## 3-Tier MCP

- **Tier 1 (built-in)**: context7, codegraph, grep-app, lsp, websearch
- **Tier 2 (Claude Code compat)**: `.mcp.json` parser with
  `${VAR}` env expansion
- **Tier 3 (skill-embedded)**: per-session MCPs with OAuth 2.0 PKCE
  and Dynamic Client Registration (DCR) stubs

## Conventions

- **Runtime**: Go 1.26+; CGO disabled; `GOEXPERIMENT=greenteagc`.
- **Formatting**: `gofmt` strict.
- **Naming**: stdlib conventions; exported = PascalCase, unexported = camelCase.
- **Errors**: Return explicitly; wrap with `fmt.Errorf("...: %w", err)`.
- **Context**: First parameter on all long-running operations.
- **Tests**: Stdlib `testing` + `stretchr/testify`; co-located `*_test.go`.
- **Config**: JSONC; heretic.json overrides defaults.

## Anti-Patterns

- Never `panic()` in non-error paths.
- Never use `as any`-style escape hatches.
- Never commit binary artifacts.
- Never use `background_cancel(all=true)`.
- Never delete a failing test to make a build green.
- Never add new external dependencies unless strictly necessary.

## License

FSL-1.1-MIT — see [LICENSE.md](./LICENSE.md). Upstream attribution in
[NOTICE.md](./NOTICE.md).
