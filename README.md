# Heretic Ultimate

A terminal-first AI assistant forked from
[charmbracelet/crush](https://github.com/charmbracelet/crush) (FSL-1.1-MIT)
and extended with the full om-my-opencode feature set, reimplemented in
Go.

## What's new vs crush

This is the **Ultimate** edition. It includes everything in crush plus
a port of the major om-my-opencode features:

- **11 built-in agents** — sisyphus (orchestrator), hephaestus (plan
  builder), oracle (advisor), librarian (external research), explore
  (codebase), multimodal-looker (images), metis (intent clarifier),
  momus (plan reviewer), atlas (multi-agent orchestrator),
  sisyphus-junior (delegated coder), prometheus (plan generator).
- **54 lifecycle hooks** across 5 tiers (Session, ToolGuard, Transform,
  Continuation, Skill) firing on 12 events.
- **Team mode** — parallel multi-agent coordination with storage,
  mailbox, worktrees, and 12 team tools.
- **3-tier MCP** — built-in (5 servers), Claude Code `.mcp.json` parser,
  skill-embedded with OAuth 2.0 + PKCE + DCR.
- **Background agents** — concurrent LLM calls with FIFO queueing and
  parent-wake notification.
- **Boulder state** — JSON-backed work tracker that survives sessions.
- **8 task categories** — quick, ultrabrain, deep, artistry,
  visual-engineering, writing, unspecified-low, unspecified-high.
- **Hashline edit + read + diff enhancer** — LINE#ID content addressing.
- **IntentGate keyword detector** — first-message mode injection.
- **Rules injector** — `.heretic/rules/*.md` → system prompt.
- **Tool catalog with gating** — bit-set gating flags (team_mode,
  hashline, task_system, etc.).

## Install

```bash
go install github.com/gedwolmen/heretic@latest
```

## Build

```bash
go build ./...
./heretic --help
```

## Quick Start

```bash
# Interactive
heretic

# Non-interactive
heretic run "Summarize the README"

# Use a specific agent
heretic --agent hephaestus run "Read plan.md and ship it"

# Subagent delegation
heretic run "Delegate refactoring of all .go files to a subagent"

# Team mode (in heretic.json)
# { "team_mode": { "enabled": true, "max_parallel_members": 4 } }
heretic run "Use team mode to refactor auth"
```

## Configuration

```jsonc
{
  "team_mode": {
    "enabled": true,
    "max_parallel_members": 4,
    "max_members": 8
  },
  "background_agent": {
    "model_concurrency": 5,
    "queue_on_full": true
  },
  "boulder_state": {
    "enabled": true,
    "auto_advance": true,
    "worktree_enabled": true
  },
  "delegate": {
    "concurrency": 5
  }
}
```

## Architecture

```
internal/
  agent/builtin/        # 11 built-in agents + embedded prompts
    prompts/*.md        # ~50KB of system prompts
  hookengine/           # 5-tier hook engine
    hooks/
      session/          # Session-tier hooks
      toolguard/        # ToolGuard-tier hooks
      transform/        # Transform-tier hooks
      continuation/     # Continuation-tier hooks
      skill/            # Skill-tier hooks
  team/                  # Team mode (storage, mailbox, 12 tools, worktrees)
  mcp/                   # 3-tier MCP + OAuth
  background/            # Background agent (concurrency, parent wake)
  boulder/              # Boulder state (work tracker)
  hashline/              # LINE#ID + diff enhancer
  intentgate/            # Keyword detector + mode injection
  rules/                 # .heretic/rules/*.md loader
  delegate/              # Subagent delegation tool
  toolregistry/          # Tool catalog with gating flags
  config/                # Config schema
  agent/tools/           # 20+ tool implementations
  ... (the original crush packages)
```

## License

FSL-1.1-MIT. See [LICENSE.md](./LICENSE.md) and [NOTICE.md](./NOTICE.md).

## Acknowledgments

- [Charmbracelet](https://github.com/charmbracelet) for the original
  crush codebase.
- [code-yeongyu](https://github.com/code-yeongyu) for the
  oh-my-opencode design patterns.
- The Go community.
