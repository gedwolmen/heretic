# AGENTS.md

Guide for agents working in the **Heretic** codebase — a terminal-first AI coding
assistant written in Go. Module path: `github.com/gedwolmen/heretic` (note: it is
**not** `charmbracelet/*`; all internal imports use the `gedwolmen/heretic` path).

A separate, UI-specific guide lives at [`internal/ui/AGENTS.md`](internal/ui/AGENTS.md)
and should be consulted for any TUI rendering work.

## Essential commands

Build/test/lint are driven by [Taskfile.yaml](Taskfile.yaml) (`task`/`go-task`).
CI ([.github/workflows/build.yml](.github/workflows/build.yml)) runs the bare `go`
equivalents, so both must stay green.

| Task | Command | Notes |
| --- | --- | --- |
| Build | `task build` | `CGO_ENABLED=0 GOEXPERIMENT=greenteagc`. Auto-injects version via git describe. |
| Run (dev) | `task dev` | `go run .` with `HERETIC_PROFILE=true` (pprof at `:6060`). |
| Test | `task test` | `go test -race -failfast ./...`. **Race detector is always on.** |
| Re-record LLM cassettes | `task record` | Deletes `internal/agent/testdata` and re-records VCR cassettes; needs `HERETIC_HYPER_API_KEY`. |
| Lint | `task lint` | `golangci-lint run` (v2, config `.golangci.yml`) + `scripts/check_log_capitalization.sh`. |
| Lint+fix | `task lint:fix` | |
| Format | `task fmt` | `gofumpt -w .` (gofumpt, **not** gofmt). |
| Modernize | `task modernize` | Runs `golang.org/x/tools/.../modernize -fix -test ./...`. |
| Config schema | `task schema` | `go run main.go schema > schema.json` — regenerate after changing `internal/config/config.go`. |
| OpenAPI spec | `task swag` | swag annotations → `internal/swagger/`. Re-run after changing `internal/server`/`internal/proto`/`main.go`. |
| SQL codegen | `task sqlc` | `sqlc generate` from `sqlc.yaml` → `internal/db/*.sql.go`. |
| Hyper provider | `task hyper` | `go generate ./internal/agent/hyper/...` → `provider.json`. |
| Update deps | `task deps` | Bumps `charm.land/fantasy` + `charm.land/catwalk` (`GOPROXY=direct`). |
| Isolated onboarding | `task run:onboarding` | Runs with `HERETIC_GLOBAL_DATA`/`HERETIC_GLOBAL_CONFIG` under `tmp/onboarding/`. |
| Release | `task release` | Semver tag via `svu`; requires `main`, clean tree, **and** passing `build.yml` + `snapshot.yml` CI runs. Signs + pushes tags. |

Raw equivalents (what CI checks): `go mod tidy && git diff --exit-code`, then
`go build -race ./...`, then `go test -race -failfast ./...`.

### Lint rules that bite

- `.golangci.yml` is **v2** format. `errcheck`, `ineffassign`, `unused` are
  **disabled**; do not assume they will catch mistakes — write defensively.
  Enabled: `bodyclose`, `staticcheck`, `noctx`, `misspell`, `whitespace`,
  `sqlclosecheck`, `rowserrcheck`, `tparallel`, `nolintlint`, `goprintffuncname`.
  Formatters: `gofumpt` + `goimports`.
- **Log messages must start with a capital letter.**
  `scripts/check_log_capitalization.sh` greps for
  `slog.(Error|Info|Warn|Debug|Fatal|Print|Println|Printf)([""][a-z]` and fails
  the build. `noctx` is suppressed for `slog.*`/`log.*` calls.
- `go mod tidy` is enforced: CI runs it then `git diff --exit-code`, so
  `go.mod`/`go.sum` must be tidy in every commit.

## Architecture & control flow

```
main.go → cmd.Execute() (fang/cobra)
   │
   ▼
internal/cmd          CLI commands (root, run, server, login, projects, schema, stats…)
   │
   ▼
internal/workspace    Resolves cwd → workspace, spawns or attaches to a server
   │
   ▼
internal/server       HTTP over Unix socket (unix://) or Windows named pipe (npipe://)
   │  - SSE for events, JSON over HTTP for RPC. OpenAPI at /v1 (internal/swagger).
   ▼
internal/backend      Transport-agnostic business logic: Workspace lifecycle,
   │                    client attach/hold/grace timers, path dedup index.
   ▼
internal/app          Wires services (sessions, messages, history, permissions,
   │                    filetracker, LSP, skills) + agent.Coordinator. Owns pubsub
   │                    brokers for tea.Msg events, agent notifications, runCompletions.
   ▼
internal/agent        Core orchestration: Coordinator + sessionAgent. Runs the
                        fantasy LLM loop with tools, queuing, auto-summarization,
                        title generation. 11 builtin agents in internal/agent/builtin.
```

### Client/server vs. in-process

The interactive TUI is a **client**. By default it runs an in-process server, but
setting **`HERETIC_CLIENT_SERVER=1`** forces a separate server process and connects
the TUI to it over the socket. `heretic run` (non-interactive) is always a client.
The socket path: `unix://$XDG_RUNTIME_DIR/heretic-<uid>.sock` (falls back to
`/tmp/heretic-<uid>.sock` if the path would exceed the 104-byte macOS limit);
`npipe:////./pipe/heretic-<uid>.sock` on Windows. See `server.DefaultHost()`.

### Backend locking (gotcha)

`Backend.mu` is always acquired **before** `Workspace.clientsMu`. Detach paths drop
`clientsMu` before calling `teardown` (which re-takes `Backend.mu`) to avoid an
AB/BA deadlock with `CreateWorkspace`. Any new code touching both locks must
preserve this order — see the doc comment on `Backend` in
`internal/backend/backend.go`.

### Agent run lifecycle (non-obvious)

- `Coordinator.Run` is the entry point; it delegates to `sessionAgent.Run` with a
  `SessionAgentCall` carrying `RunID`, `ProviderOptions`, `OnComplete`, and an
  optional `Accepted` reservation.
- **`RunID` is the reliable completion contract.** `SessionID` alone is ambiguous
  when concurrent turns share a session; callers that need a terminal event (e.g.
  `heretic run` against a possibly-busy session) MUST set `RunID`. The
  authoritative terminal signal is a `notify.RunComplete` published on
  `app.RunCompletions()` — **not** message-finish parts.
- Busy sessions **queue** calls; `OnComplete` is intentionally stripped when
  queueing (the originating `coordinator.Run` has already returned).
- `BeginAccepted`/`AcceptedRun`/`cancelMark` implement an accept-sequence
  high-water mark so a single `Cancel` covers every accepted-but-not-yet-active
  prompt without poisoning later ones. This is intricate; read
  `internal/agent/agent.go` around `dispatchMu`/`acceptedRuns`/`cancelMark`
  before touching dispatch/cancel paths.

### LLM providers

Models are abstracted through `charm.land/fantasy`. Providers in
`internal/agent/coordinator.go`: anthropic, openai, google, bedrock, azure,
openrouter, vercel, openaicompat, plus Copilot (OAuth, `internal/oauth/copilot`)
and **Hyper** (Charm's hosted gateway, `internal/agent/hyper`,
`https://hyper.charm.land/v1`). Two maps select API style per model name:
`copilotResponsesModels` (Copilot models using the Responses API) and
`opencodeMessagesModels` (models using the Anthropic Messages API). When adding a
new model, check whether it belongs in one of these maps.

## Tool system

Tools live in `internal/agent/tools/`. Each tool is a Go file plus a description
template (`.md` or `.md.tpl`, rendered as a Go `text/template`).

Non-obvious rules:

- **Description templates** are rendered via `renderToolDescription` with a
  `toolDescriptionData{GhAvailable}` struct. `ghAvailable` is computed once at
  package init via `exec.LookPath("gh")` — **but is forced `false` under
  `testing.Testing()`**. Tests therefore never see `gh`-aware descriptions; do
  not assert on gh-gated branches in tests.
- **Tool gating** is declarative in `internal/toolregistry/registry.go` via
  `GatingFlag` bits: `GateTeamMode`, `GateTaskSystem`, `GateHashlineEdit`,
  `GateInteractiveBash` (requires the `tmux` binary on PATH), `GateLookAt`,
  `GateBackgroundAgent`. Add a new gate flag here and check it where the registry
  is filtered; do not scatter ad-hoc feature checks.
- **Context propagation**: tools read session/message/model info from the
  `context.Context` using the typed keys in `tools.go` (`SessionIDContextKey`,
  `MessageIDContextKey`, `SupportsImagesContextKey`, `ModelNameContextKey`). Use
  the `Get*FromContext` helpers; never read these from globals.
- **MCP tools** are namespaced with the `mcp_` prefix and dispatched through
  `internal/agent/tools/mcp/`. They are initialized lazily from config
  (`mcp.Initialize` in `app.New`).
- **Hooked tools**: `internal/agent/hooked_tool.go` wraps tool execution so
  `PreToolUse` hooks run before each top-level tool call (see Hooks below).
- **Permission denial** returns `NewPermissionDeniedResponse()` with
  `StopTurn = true` so the agent loop does not retry.

### hashline (content-addressed edits)

`internal/hashline/` implements `LINE#ID` content addressing. The `Read` tool
tags each output line with `#<4-char-hash>` (alphabet `ZPMQVRWSNKTXJBYH`, mixing
prev/next lines to reduce collisions), and the `hashline_edit` tool requires the
hash to match current file content before applying an edit — stale edits (file
changed between read and edit) are rejected. Gated behind `GateHashlineEdit`.
This is a reimplemented-from-scratch algorithm (FNV-64a); do not "simplify" it
without preserving the exact alphabet and prev/next mixing.

## Shell execution

`internal/shell/` uses **`mvdan.cc/sh/v3`** (an embedded POSIX shell) even on
Windows — there is no dependency on bash/cygwin/WSL. Consequences:

- Use **forward slashes** in all paths passed to the shell.
- `HereticEnvMarkers()` sets `HERETIC=1`, `AGENT=heretic`, `AI_AGENT=heretic` on
  every spawned shell (both the `bash` tool and the hook runner). Scripts can
  detect "am I run by an AI agent?" via any of these.
- `BlockFunc`s gate commands; the bash tool's banned-command list is enforced
  here (see `internal/shell/dispatch.go`).
- Background jobs (`background.go`) return shell IDs consumed by `job_output`/
  `job_kill`. Long-running processes must use the background path, not `&`.
- jq is available in-process via `itchyny/gojq` (`internal/shell/jq.go`).

## Configuration

JSON, layered. `internal/config/ConfigStore` is the **single entry point** for
all config access — never read config files directly.

| Scope | Default path | Env override |
| --- | --- | --- |
| Global config | `~/.config/heretic/heretic.json` | `HERETIC_GLOBAL_CONFIG` (dir) |
| Global **data** config (where overrides-on-load are written) | `~/.local/share/heretic/heretic.json` (`%LOCALAPPDATA%/heretic/` on Windows) | `HERETIC_GLOBAL_DATA` (dir) or `XDG_DATA_HOME` |
| Workspace config | `<data_directory>/heretic.json` | — |
| Cache dir | `~/.cache/heretic` | `HERETIC_CACHE_DIR` or `XDG_CACHE_HOME` |
| Data directory (per-project state) | `.heretic` (walked up to project boundary via `fsext.LookupClosestBounded`) | `data_directory` in config |

Key behaviors:

- `app.Name` is the constant `"heretic"`; config files are named `heretic.json`.
- Writes are serialized in-process (`mu`) **and** cross-process via `lock.File`
  with a 5s deadline (`configLockDeadline`). `ScopeGlobal`/`ScopeWorkspace`
  select the target file.
- `RuntimeOverrides` (e.g. `SkipPermissionRequests` from `--yolo`) are
  per-process only and **never persisted**.
- Config auto-reloads on file change (`reloadMu` serializes; `autoReload` uses
  `TryLock` to skip redundant reloads).
- A `VariableResolver` expands `${...}` references in config values.
- The repo-root `heretic.json` is this project's **own** config (gopls LSP
  options) and doubles as the `$schema` example; it is not a template for users.
- `schema.json` is generated (`task schema`) from the `Config` struct's
  `jsonschema` tags — regenerate it whenever you change `internal/config/config.go`.
- Early slog calls in `config.Load` are deliberately discarded (see the FIXME in
  `cmd.Execute`); the file logger starts only after the data dir is known.

## Skills

`internal/skills/`. Builtin skills are **embedded** at compile time via
`//go:embed builtin/*` in `embed.go` and addressed with the virtual
`heretic://skills/<name>/SKILL.md` prefix — the `View` tool resolves these from
the embedded FS, **not** disk. To add a builtin skill: drop a
`internal/skills/builtin/<name>/SKILL.md` (YAML frontmatter `name`+`description`,
directory name must match `name`), then add an assertion in
`TestDiscoverBuiltin` (`internal/skills/skills_test.go`). User skills with the
same name override builtins. Each workspace has one `Manager`; `WithGlobalMirror`
bridges it to package globals but is **only safe for single-workspace processes**
(local mode / client) — the backend server hosts many workspaces and must not
mirror. Repo-local agent skills live in `.agents/skills/`.

## Hooks

User-defined shell scripts that fire on agent events (`internal/hooks/` +
`internal/hookengine/`). Currently only **`PreToolUse`** is supported. Full
contract in [docs/hooks/README.md](docs/hooks/README.md). Critical points:

- Run through the same embedded `mvdan.cc/sh`; inline commands and
  shebang-less scripts run in-process, `#!` scripts dispatch via `os/exec` with
  permissive PATH fallback. Identical on macOS/Linux/Windows.
- Hooks run **in parallel** but compose in **config order**: last hook wins for
  input rewrites, **first deny wins** for blocking. Exit code `2` = block.
- `PreToolUse` fires **only on the top-level agent's** tool calls. Sub-agents
  (`agent` task tool, `agentic_fetch`, …) run **without** hook interception —
  but the outer sub-agent tool call itself *is* hooked.
- `command` is resolved relative to the **current working directory**, not the
  config file. Global hooks (`~/.config/heretic/heretic.json`) therefore need
  absolute paths or inline commands.
- `hookengine` organizes hooks into tiers (`continuation`, `session`, `skill`,
  `toolguard`, `transform`) under `internal/hookengine/hooks/`.

## Database

SQLite via `github.com/ncruces/go-sqlite3` (WASM-based, hence `CGO_ENABLED=0`).
`internal/db/` holds sqlc-generated query code (`*.sql.go`), goose migrations in
`internal/db/migrations/`, and models. `db.New(conn)` returns the queries;
connections are pooled per data dir and released on shutdown. After changing SQL
or queries: edit `internal/db/sql/*.sql` + `sqlc.yaml`, run `task sqlc`, and
**add a goose migration** if the schema changes. `sqlclosecheck`/`rowserrcheck`
linters enforce `rows.Close()` and `rows.Err()`.

## Testing

- **Always run with `-race`** (`task test` and CI both enable it). Race-specific
  suites live in `*_race_on_test.go` / `*_race_off_test.go`.
- **VCR cassettes** (`charm.land/x/vcr`) record/replay LLM HTTP traffic in
  `internal/agent/testdata/TestCoderAgent/<model>/*.yaml`. Replay is the default;
  `task record` (alias `test:record`) wipes and re-records — requires
  `HERETIC_HYPER_API_KEY` and network access. Do not commit cassette changes
  incidentally.
- **Golden files** are used heavily in the UI layer, especially
  `internal/ui/diffview/testdata/**/*.golden` (regenerate with `-update` /
  `testify`'s golden helpers).
- `.envrc` (direnv) + `.env.sample` + `godotenv/autoload` (imported in test
  files) load env for tests.
- `internal/cmd/clientserverrace/` is a **standalone package** regression test
  for the `HERETIC_CLIENT_SERVER=1` socket-init race; it builds the binary
  directly so it still compiles if the `cmd` package is temporarily broken.
- `internal/agent/agenttest/` provides a coordinator test harness.

## Platform & build constraints

- `CGO_ENABLED=0` always (SQLite is WASM). `GOEXPERIMENT=greenteagc` for builds
  (note: **unset during lint** via `GOEXPERIMENT: null` in the `lint` task).
- Platform splits via filename suffixes: `*_windows.go` / `*_other.go`
  (dial, exec, net, drive, lock, root). When adding platform-specific code,
  follow the existing `_windows.go` + `_other.go` (or `_unix.go`) pair pattern.
- `goreleaser` builds for linux/darwin/windows/freebsd/openbsd/netbsd/android
  across amd64/arm64/386/arm. License is **FSL-1.1-MIT**. Completions
  (bash/zsh/fish) and manpages are generated in `before.hooks` and shipped.

## Other packages worth knowing

- `internal/agent/builtin/` — the 11 builtin agents (primary: sisyphus,
  hephaestus, atlas, prometheus; subagent: oracle, librarian, explore,
  multimodal-looker, metis, momus, sisyphus-junior). Each has a prompt
  `prompts/<name>.md`. Adding an agent = struct + registry entry + prompt.
- `internal/agent/prompt/`, `internal/agent/notify/` — prompt assembly and the
  notification/`RunComplete` event types bridged into app pubsub.
- `internal/intentgate/` — classifies user intent (prompts in `prompts/`:
  `analyze`, `search`, `team`, `ultrawork`) to route prompts (e.g. search vs.
  agent vs. team mode).
- `internal/team/` — team mode (multi-agent), gated by `GateTeamMode` /
  `team_mode.enabled`. Worktrees, mailbox, eligibility, storage.
- `internal/boulder/` — `boulder.json` work tracker persisted across sessions.
- `internal/lsp/` — LSP manager + client; gopls is configured in the repo-root
  `heretic.json`. Diagnostics feed back into the app via a callback set in
  `app.New`. `lsp_references`/`lsp_diagnostics`/`lsp_restart` are tools.
- `internal/csync/` — concurrency primitives (`Map`, `Slice`, `Value`,
  `VersionedMap`) used throughout the agent layer; prefer these over raw
  `sync.*` for consistency with the dispatch/queue machinery.
- `internal/version/` — version string injected via `-ldflags` at build time
  (see `task build` / `task install`); empty unless built with git tags.
- `internal/update/` — self-update checker launched in `app.New`.

## Conventions

- Formatting: **gofumpt** (stricter than gofmt). Run `task fmt` before committing.
- Package comments: every package has a doc comment on its `package` declaration
  explaining its role; preserve and extend this convention.
- Errors: package-level sentinel `var Err... = errors.New(...)` are the norm in
  `backend`, `config`, etc. Wrap with `%w` when re-raising.
- Doc comments on exported identifiers are expected (staticcheck + review).
- When referencing code locations to humans or in PRs, use `file_path:line`.
