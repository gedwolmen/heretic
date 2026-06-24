Use this tool to delegate a discrete subtask to a specialized subagent.

Each subagent runs in its own session with its own provider/model selection.
Use the `category` parameter to route the task to the right specialist:

- `explore` — read-only investigation of the codebase
- `librarian` — research external libraries and documentation
- `search` — fast keyword search across files
- `edit` — focused code edits

The parent session is blocked until the subagent completes, so prefer
delegation for tasks that would otherwise consume the parent's context window.

Concurrency is bounded per provider/model (default 5 simultaneous spawns).
Excess calls queue; they do not fail.
