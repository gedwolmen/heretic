---
description: (builtin) Start Sisyphus work session from Prometheus plan
argument-hint: [plan-name]
---

You are starting a Sisyphus work session.

## ARGUMENTS

- `/start-work [plan-name]`
  - `plan-name` (optional): name or partial match of the plan to start

## WHAT TO DO

1. **Find available plans**: Search for Prometheus-generated plan files at `.omo/plans/`

2. **Check for active boulder state**: Read `.omo/boulder.json` if it exists

3. **Decision logic**:
   - If multiple active works are listed in your context:
     - This means boulder.json has more than one work with status: `active` or `paused`
     - Use the Question tool to ask the user which plan to resume
     - Resume by running `/start-work {plan-name}` for the selected plan
     - If the user says "start a new plan", continue with cold-start auto-selection logic
   - If exactly one active work is listed and the user did not name a plan:
     - Auto-resume that single active work
   - If no active plan OR plan is complete:
     - List available plan files
     - If ONE plan: auto-select it
     - If MULTIPLE plans: show list with timestamps, ask user to select

4. **Worktree Setup** (if requested):
   - `git worktree list --porcelain` - see available worktrees
   - Create: `git worktree add <absolute-path> <branch-or-HEAD>`
   - Update boulder.json to add `"worktree_path": "<absolute-path>"`
   - All work happens inside that worktree directory

5. **Create/Update boulder.json**:
   ```json
   {
     "active_plan": "/absolute/path/to/plan.md",
     "status": "active",
     "started_at": "<ISO timestamp>"
   }
   ```

6. **Read the plan** end-to-end before touching any code.

7. **Execute the plan step by step**, marking progress in boulder.json after each step.

8. **After completion**:
   - Run full test suite
   - Update boulder.json status to "completed"
   - If a worktree was used, leave it in place (user can clean up later)

## RULES

- Do NOT skip the plan. If the user says "just start", find or generate a plan first.
- Do NOT add features that are not in the plan.
- If a step cannot be completed, STOP and report. Do not improvise.
- Do NOT commit unless the user explicitly asks.

## OUTPUT FORMAT

After each step, output:

```
[boulder/step-N] <what you did>
  verification: <test result>
  next: <what's next>
```

When complete:

```
[boulder/done] all steps finished
  verification: full test suite passed
  plan: /absolute/path/to/plan.md
```
