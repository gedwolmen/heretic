# Heretic

A terminal-first AI coding assistant written in Go. Runs an in-process server
(or a separate client/server pair over a Unix socket / Windows named pipe) with
a rich TUI client built on [Bubble Tea][bubbletea].

> Go reimplementation, started as a fork of [charmbracelet/crush][crush] with an
> extended feature set ported to Go.

## Features

- **TUI client** — chat, diff views, sessions, dialogs, inline commands,
  completions ([Bubble Tea][bubbletea] + [Lipgloss][lipgloss]).
- **Client/server** — HTTP-over-socket (or named pipe) server; JSON + SSE API at
  `/v1`. In-process by default, or `HERETIC_CLIENT_SERVER=1` for a split server.
- **Many LLM providers** — Anthropic, OpenAI, Google, Bedrock, Azure,
  OpenRouter, Vercel, OpenAI-compatible, Copilot (OAuth), Hyper; local
  providers (Ollama, LM Studio, OllamaX, LiteLLM) auto-discovered.
- **Tools** — content-addressed `hashline` edits, file/shell/grep/glob, LSP
  tools, sub-agents, and `mcp_`-prefixed MCP tools.
- **Embedded POSIX shell** — [`mvdan.cc/sh/v3`][mvdan-sh] on every platform
  (no bash/cygwin/WSL needed); in-process `jq`.
- **Hooks** — user shell scripts on agent events (`PreToolUse`), cross-platform.
- **Skills** — embedded compile-time builtins (`heretic://skills/...`),
  overridable by users.
- **LSP** — manager + client (gopls by default); diagnostics feed back in.
- **Team mode** — multi-agent orchestration (worktrees, mailbox, eligibility).
- **SQLite** — [WASM][go-sqlite3], `CGO_ENABLED=0`, [sqlc][sqlc] + goose.
- **Cross-platform** — linux, darwin, windows, freebsd, openbsd, netbsd,
  android (amd64/arm64/386/arm).

## Quick start

Requires [Go](https://go.dev) 1.26+ and [Task](https://taskfile.dev).

```bash
task build      # CGO_ENABLED=0, version from git
task dev        # run with pprof at :6060
go run .        # or go directly
```

| Flag | Description |
| --- | --- |
| `-c, --cwd` | Working directory |
| `-d, --debug` | Debug mode |
| `-y, --yolo` | Auto-accept all permissions (dangerous) |
| `-s, --session` | Continue a session by ID |
| `-C, --continue` | Continue the most recent session |
| `-H, --host` | Connect to a specific server host |

Config is layered JSON: global `~/.config/heretic/heretic.json`, data
`~/.local/share/heretic/heretic.json`, workspace `<data_directory>/heretic.json`.
Regenerate the schema with `task schema`.

## Tasks

`task build` · `task test` (race on) · `task lint` (golangci-lint v2 + log-cap
check) · `task fmt` ([gofumpt][gofumpt]) · `task swag` (OpenAPI) · `task sqlc`
(SQL codegen) · `task schema` · `task record` (VCR cassettes) · `task release`.

## Architecture

```
main.go → internal/cmd (cobra) → internal/workspace → internal/server
   → internal/backend → internal/app (+ agent.Coordinator) → internal/agent
```

The TUI is a client; `heretic run` (non-interactive) is always a client. Full
control flow, gotchas, and conventions are documented in
[**AGENTS.md**](AGENTS.md) (and [internal/ui/AGENTS.md](internal/ui/AGENTS.md)
for TUI work; [docs/hooks/](docs/hooks/) for hooks).

## Tech stack

Go 1.26 (`CGO_ENABLED=0`) · [Bubble Tea][bubbletea]/[Lipgloss][lipgloss]/
[Chroma][chroma] · [`charm.land/fantasy`][fantasy] · [`mvdan.cc/sh/v3`][mvdan-sh]
· [SQLite (WASM)][go-sqlite3] + [sqlc][sqlc] + goose · [Cobra][cobra]/[fang][fang]
· MCP · LSP · [gojq][gojq].

## Contributing

Read [AGENTS.md](AGENTS.md) first — build commands, lint rules (gofumpt
formatting, capitalized log messages, tidy `go.mod`), architecture, conventions.
Run `task fmt` and `task lint` before submitting.

## License

[MIT](LICENSE)

[crush]: https://github.com/charmbracelet/crush
[bubbletea]: https://github.com/charmbracelet/bubbletea
[lipgloss]: https://github.com/charmbracelet/lipgloss
[chroma]: https://github.com/alecthomas/chroma
[fantasy]: https://charm.land/
[mvdan-sh]: https://github.com/mvdan/sh
[go-sqlite3]: https://github.com/ncruces/go-sqlite3
[sqlc]: https://sqlc.dev/
[cobra]: https://github.com/spf13/cobra
[fang]: https://github.com/charmbracelet/fang
[gojq]: https://github.com/itchyny/gojq
[gofumpt]: https://github.com/mvdan/gofumpt
