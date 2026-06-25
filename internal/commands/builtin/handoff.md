---
description: (builtin) Create a detailed context summary for continuing work in a new session
argument-hint: [goal]
---

You are creating a handoff document for a future session. The goal is for
a fresh agent (with no prior context) to pick up exactly where you left off.

## Process

1. **Gather current state**:
   - Read the current session ID and any open todos
   - Check git status (uncommitted changes, branch, last commit)
   - Read boulder.json if it exists
   - Note any pending tool calls or background tasks

2. **Write a handoff document** to `.omo/handoffs/<session-id>.md` with these sections:

   ### Goal
   One paragraph: what we're trying to achieve.

   ### Current State
   - Branch: <branch>
   - Last commit: <sha> <message>
   - Uncommitted changes: <summary or "none">
   - Open todos: <list or "none">

   ### What's Done
   Bullet list of completed work, with file:line references.

   ### What's Next
   Numbered list of the immediate next steps. Each step should be
   concrete enough to start without further clarification.

   ### Key Context
   Anything a fresh agent needs to know that isn't obvious from the
   code:
   - Why we made certain design decisions
   - Known issues / workarounds in place
   - External dependencies (services, env vars, etc.)
   - Gotchas the codebase has

   ### Verification
   How to confirm the next session is on track:
   - `go test ./...` should pass
   - boulder.json status should be "active" (not "completed")
   - specific integration tests that exercise the changed code

3. **Update boulder.json** to point at the handoff document so the next
   session can find it.

4. **Tell the user** the handoff is ready and where it lives. Suggest:
   > To resume: `heretic --session <session-id>` or start a new session
   > and run `/start-work handoff`.

## RULES

- Do NOT include any code in the handoff that isn't already in the repo.
- Do NOT speculate about future design — only document what is and
  what should happen next.
- Keep the handoff terse. A future agent will read it in 30 seconds.
- If the session is in a good state (tests pass, no uncommitted
  changes), say so explicitly. A clean handoff is a valid outcome.
