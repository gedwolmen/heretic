---
description: (builtin) Remove AI-generated code smells from branch changes and critically review the results
argument-hint: ''
---

You are reviewing the recent changes in the working tree (staged, unstaged, and uncommitted files) to remove AI-generated code smells and similar patterns.

## Goals

1. Identify code smells commonly introduced by AI assistants.
2. Refactor or remove them where appropriate.
3. CRITICALLY review the result: every change you make must be evaluated for whether it actually improves the code, or just adds churn.

## What to look for

- Excessive or redundant comments that narrate code.
- Over-abstracted helper functions used only once.
- Defensive null/bounds checks for impossible cases.
- Boilerplate error wrapping that adds no information.
- Re-implementing existing utilities.
- Unused parameters or return values.
- Comments like "// TODO:", "// FIXME:", "// HACK:" that exist only because the AI hedged.
- Test code that exists just to bump coverage numbers.
- Type assertions that ignore errors.
- Hardcoded magic numbers that should be named constants.
- Trivial wrapper functions around single calls.
- Variable names longer than needed (e.g. `resultOfCalculation` → `result`).

## Process

1. Run `git diff` to see all uncommitted changes.
2. Read each file's diff carefully.
3. For each smell you find, decide:
   - Remove it (the simplest fix).
   - Refactor to a more idiomatic form.
   - Leave it (if removal would harm clarity or break an API).
4. After each change, run the project's tests to confirm nothing broke.
5. After all changes, do a final review of the cumulative diff.

## CRITICAL FINAL STEP

Before declaring done, look at the cumulative diff and ask:

- Did I make this code better, or just different?
- Would a human reviewer approve each change?
- Are there any smells I ADDED in this cleanup pass?

If you can't justify a change with a clear "yes" to all three, revert it.
