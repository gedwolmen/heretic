# Heretic

A terminal-first AI assistant for software development, forked from
[charmbracelet/crush](https://github.com/charmbracelet/crush) and extended with
four concept ports from the [oh-my-opencode](https://github.com/code-yeongyu/oh-my-openagent)
project, reimplemented in Go.

## What is Heretic?

Heretic is a glamorous, terminal-first AI agent. It is a drop-in rename of
heretic, plus four new features borrowed from the oh-my-opencode project:

1. **Subagent delegation** — spawn child sessions via a `task` tool to parallelize work,
   with configurable per-provider/model concurrency.
2. **Hashline edit + read** — every read output is tagged with a short `LINE#ID` hash so
   the edit tool can reject stale edits before they corrupt your files.
3. **IntentGate keyword detector** — first-message keyword scan that injects
   mode-specific prompts (`ultrawork`, `search`, `analyze`, `team`).
4. **Rules injector hook** — automatically loads `.heretic/rules/*.md` files into the
   system prompt so per-project conventions are honored on every session.

See [NOTICE.md](./NOTICE.md) for the upstream attribution and license terms.

## Install

```bash
go install github.com/gedwolmen/heretic@latest
```

Pre-built binaries are not yet published; building from source requires Go 1.26+.

## Build

```bash
go build ./...
./heretic --help
```

## Quick Start

```bash
# Interactive mode
heretic

# Non-interactive
heretic run "Summarize the README"

# Subagent delegation
heretic run "Delegate refactoring of all .go files to a subagent"

# Hashline-protected editing
heretic run "Add a doc comment to func Foo in main.go"

# IntentGate keyword
heretic run "ultrawork build a hello world program"

# Rules injection: place rules under .heretic/rules/
mkdir -p .heretic/rules
cat > .heretic/rules/style.md <<EOF
Always prefer stdlib over external dependencies.
EOF
```

## License

FSL-1.1-MIT — see [LICENSE.md](./LICENSE.md) and [NOTICE.md](./NOTICE.md).

## Architecture

Heretic is a Go binary. Source layout:

```
.
├── main.go                  # CLI entry point
├── internal/                # packages
│   ├── agent/               # SessionAgent and tool registry
│   ├── cmd/                 # Cobra commands (root, run, server, etc.)
│   ├── config/              # Config schema + load logic
│   ├── session/             # Session CRUD
│   ├── message/             # Message model
│   ├── permission/          # Tool permission checks
│   ├── skills/              # Skill discovery and loading
│   ├── hooks/               # Hook engine
│   ├── lsp/                 # LSP client manager
│   ├── delegate/            # Subagent delegation (OmO port)
│   ├── hashline/            # LINE#ID content hashing (OmO port)
│   ├── intentgate/          # Keyword detector + mode injection (OmO port)
│   └── rules/               # .heretic/rules loader (OmO port)
├── heretic.json             # Default user config
└── Taskfile.yaml            # Build/lint/test tasks
```

## Rules

Heretic reads project-level rules from `.heretic/rules/*.md`. Each `.md` file in that
directory is concatenated into the system prompt before the agent's first turn. See
the `internal/rules/` package and the [`.heretic/rules/`](./.heretic/rules) example.

## Subagent Delegation

Use the `task` tool to spawn a child session. Concurrency is configurable per
provider/model via `delegate.concurrency` in `heretic.json` (default 5).

## Hashline

Every `Read` tool output is tagged with `LINE#ID` hashes (4-character IDs from the
alphabet `ZPMQVRWSNKTXJBYH`). The `hashline_edit` tool requires the hash to match
the current file content; stale edits are rejected.

## IntentGate

The first user message is scanned for keywords:

- `ultrawork` / `ulw` — load the ultrawork mode prompt
- `search` — load the search mode prompt
- `analyze` — load the analyze mode prompt
- `team` — load the team mode prompt

Keywords are detected on the first message of each session only. Subsequent messages
do not re-inject prompts.

## Development

```bash
go test ./... -race -failfast
go vet ./...
gofmt -l .
```

## Acknowledgments

- [Charmbracelet](https://github.com/charmbracelet) for the original crush codebase (see NOTICE.md).
- [code-yeongyu](https://github.com/code-yeongyu) for the oh-my-opencode design patterns.
- The Go community.
