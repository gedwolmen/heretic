# AGENTS.md — Heretic

> Heretic is a terminal-first AI assistant, forked from
> [charmbracelet/crush](https://github.com/charmbracelet/crush) (FSL-1.1-MIT).
> See [NOTICE.md](./NOTICE.md) for upstream attribution.

## Project Overview

Heretic is a Go CLI that drives an LLM in your terminal. It reuses the entire
heretic runtime (agent loop, tool registry, session storage, hook system, LSP, MCP,
permissions, skills, etc.) and adds four Go-idiomatic concept ports from the
[oh-my-opencode](https://github.com/code-yeongyu/oh-my-openagent) project:

| Port | Package | Description |
| --- | --- | --- |
| Subagent delegation | `internal/delegate/` | `task` tool that spawns child sessions, with per-provider/model concurrency limits. |
| Hashline edit+read | `internal/hashline/` | `LINE#ID` short hashes appended to read output; edit tool rejects stale hashes. |
| IntentGate | `internal/intentgate/` | First-message keyword scan that injects mode-specific prompts. |
| Rules injector | `internal/rules/` | Loads `.heretic/rules/*.md` into the system prompt on session start. |

## Build/Test/Lint Commands

- **Build**: `go build ./...` (produces `./heretic` binary)
- **Run**: `go run .` (or `./heretic --help`)
- **Test**: `go test ./... -race -failfast` (skips network/integration tests in CI)
- **Vet**: `go vet ./...`
- **Format**: `gofmt -l .` (empty output = all formatted)
- **Schema regeneration**: `go run main.go schema > schema.json`
- **Tidy**: `go mod tidy`

## Architecture

```
main.go                          # CLI entry
internal/
  agent/                         # SessionAgent (LLM loop) + tool registry
  cmd/                           # Cobra commands
  config/                        # Config schema + load logic
  session/                       # Session CRUD (SQLite via sqlc)
  message/                       # Message model + content types
  permission/                    # Tool permission checks
  skills/                        # Skill discovery (builtin + .skills/)
  hooks/                         # Hook engine (PreToolUse, PostToolUse, etc.)
  lsp/                           # LSP client manager (on-demand)
  db/                            # SQLite + sqlc + migrations
  event/                         # Telemetry (PostHog)
  pubsub/                        # Internal event bus
  shell/                         # Bash command execution
  delegate/                      # OmO port: subagent delegation
  hashline/                      # OmO port: LINE#ID content hashing
  intentgate/                    # OmO port: keyword detector + mode injection
  rules/                         # OmO port: .heretic/rules/*.md loader
  filetracker/                   # Tracks files touched per session
  history/                       # Prompt history
```

## Conventions

- **Runtime**: Go 1.26+; CGO disabled (`CGO_ENABLED=0`); `GOEXPERIMENT=greenteagc`.
- **Formatting**: `gofmt` strict; CI runs `gofmt -l .` and fails on any output.
- **Naming**: Go stdlib conventions; exported = PascalCase, unexported = camelCase.
- **Errors**: Return explicitly; wrap with `fmt.Errorf("...: %w", err)`.
- **Context**: First parameter on all long-running operations.
- **Tests**: Stdlib `testing` + `stretchr/testify`; co-located `*_test.go`.
- **Config**: JSONC with comments + trailing commas; Zod-style validation.
- **No emojis** in code, commit messages, or rendered output.

## Anti-Patterns

- Never `panic()` in non-error paths.
- Never use `as any`-style escape hatches.
- Never commit binary artifacts.
- Never use `background_cancel(all=true)` — cancel by `taskId`.
- Never delete a failing test to make a build green.
- Never add new external dependencies unless strictly necessary for a feature port.

## Build and Distribution

- **No distribution channels** beyond `go install` and `go build` from source.
- The `.goreleaser.yml` is preserved (with brew/AUR/winget/npm sections removed) for
  future use, but no release pipeline is wired up.
- No CI workflows are required to run; the existing `.github/workflows/build.yml`
  remains for any future CI.

## How the OmO Ports Fit In

### delegate/

The `delegate` package implements a subagent spawner. It exposes a `DelegateTask`
function and a `Registry` that maps category strings to available agent types. The
`task` tool is registered in `internal/agent/tools/` so the LLM can call it. Each
spawned child session is a fresh `Session` with its own provider/model selection and
its own concurrency slot.

The default concurrency is `5` per provider/model, overridable via
`delegate.concurrency` in `heretic.json`.

### hashline/

The `hashline` package implements a content-aware line ID. `HashLine(line, prev, next)`
returns a 4-character ID from the alphabet `ZPMQVRWSNKTXJBYH` (a subset of base-16
plus a few letters). The `TagReadOutput` helper appends `#ID` to each line of a read
result. The `ValidateEdit` helper rejects an edit if the hash doesn't match the
current file content.

The read and edit tools in `internal/agent/tools/` are wrapped to call these helpers.

### intentgate/

The `intentgate` package detects keywords on the first user message of each session:
`ultrawork` / `ulw`, `search`, `analyze`, `team`. On a hit, it loads the matching
prompt from `internal/intentgate/prompts/<mode>.md` and concatenates it into the
system prompt before the LLM is invoked.

### rules/

The `rules` package implements `LoadRules(dir string) ([]Rule, error)`. It scans
`.heretic/rules/*.md` (per the heretic convention; no legacy `.omo/rules/` support),
returns one `Rule` per file, and is invoked from a session-startup hook so its output
is appended to the system prompt.

## License

FSL-1.1-MIT — see [LICENSE.md](./LICENSE.md). Upstream attribution in
[NOTICE.md](./NOTICE.md).
