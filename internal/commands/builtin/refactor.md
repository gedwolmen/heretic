---
description: (builtin) Intelligent refactoring command with LSP, AST-grep, architecture analysis, codemap, and TDD verification.
argument-hint: <refactoring-target> [--scope=<file|module|project>] [--strategy=<safe|aggressive>]
---

# Intelligent Refactor Command

## Usage

```
/refactor <refactoring-target> [--scope=<file|module|project>] [--strategy=<safe|aggressive>]

Arguments:
  refactoring-target: What to refactor. Can be:
    - File path: src/auth/handler.ts
    - Symbol name: "AuthService class"
    - Pattern: "all functions using deprecated API"
    - Description: "extract validation logic into separate module"

Options:
  --scope: Refactoring scope (default: module)
    - file: Single file only
    - module: Module/directory scope
    - project: Entire codebase

  --strategy: Risk tolerance (default: safe)
    - safe: Conservative, maximum test coverage required
    - aggressive: Allow broader changes with adequate coverage
```

## What This Command Does

Performs intelligent, deterministic refactoring with full codebase awareness. Unlike blind search-and-replace, this command:

1. **Understands your intent** - Analyzes what you actually want to achieve
2. **Maps the codebase** - Builds a definitive codemap before touching anything
3. **Assesses risk** - Evaluates test coverage and determines verification strategy
4. **Plans meticulously** - Creates a detailed plan
5. **Executes precisely** - Step-by-step refactoring with LSP and AST-grep
6. **Verifies constantly** - Runs tests after each change to ensure zero regression

---

# PHASE 0: INTENT GATE (MANDATORY FIRST STEP)

**BEFORE ANY ACTION, classify and validate the request.**

## Step 0.1: Parse Request Type

| Signal | Classification | Action |
|--------|----------------|--------|
| Specific file/symbol | Explicit | Proceed to codebase analysis |
| "Refactor X to Y" | Clear transformation | Proceed to codebase analysis |
| "Improve", "Clean up" | Open-ended | MUST ask: "What specific improvement?" |
| Ambiguous scope | Uncertain | MUST ask: "Which modules/files?" |
| Missing context | Incomplete | MUST ask: "What's the desired outcome?" |

## Step 0.2: Validate Understanding

Before proceeding, confirm:
- [ ] Target is clearly identified
- [ ] Desired outcome is understood
- [ ] Scope is defined (file/module/project)
- [ ] Success criteria can be articulated

**If ANY of above is unclear, ASK CLARIFYING QUESTION.**

## Step 0.3: Create Initial Todos

Immediately after understanding the request, create todos:

- [ ] PHASE 1: Codebase Analysis
- [ ] PHASE 2: Build Codemap
- [ ] PHASE 3: Test Assessment
- [ ] PHASE 4: Plan Generation
- [ ] PHASE 5: Execute Refactoring
- [ ] PHASE 6: Final Verification

---

# PHASE 1: CODEBASE ANALYSIS

## 1.1: Launch Parallel Explore Agents

Fire all of these in background:

- Agent 1: Find all occurrences and definitions of the target
- Agent 2: Find all code that imports/uses/depends on the target
- Agent 3: Find similar code patterns in the codebase
- Agent 4: Find all test files related to the target
- Agent 5: Find architectural patterns around the target

## 1.2: Direct Tool Exploration

While background agents run, use direct tools:

### LSP Tools

- LspGotoDefinition - find definitions
- LspFindReferences - find all usages
- LspDocumentSymbols - file structure
- LspWorkspaceSymbols - search by name
- lsp_diagnostics - errors/warnings before we start

### AST-Grep

```
sg --pattern 'function $NAME($$$) { $$$ }' --lang go .
sg --pattern '[old_pattern]' --rewrite '[new_pattern]' --lang go .  # preview
```

### Grep

```
grep(pattern="<term>", path="src/")
```

## 1.3: Collect Results

Wait for background agents and merge their output.

---

# PHASE 2: BUILD CODEMAP

Build a complete map of the target's dependency surface:

1. **Callers** - every function/file that invokes the target
2. **Callees** - every dependency the target relies on
3. **Tests** - existing test coverage
4. **Types** - signatures in/out
5. **Side effects** - filesystem, network, global state

---

# PHASE 3: TEST ASSESSMENT

1. Run `go test ./...` to establish a baseline
2. For the target, find:
   - Direct unit tests
   - Integration tests
   - End-to-end tests
3. Note any uncovered branches
4. If coverage < 60%, write characterization tests BEFORE refactoring

---

# PHASE 4: PLAN GENERATION

Use the plan tool (or `task(category="plan")`) to write a detailed
step-by-step refactoring plan. Each step must include:

- What changes
- Which file
- Verification (test command)

---

# PHASE 5: EXECUTE

Execute the plan step by step. After each step:

1. Run the relevant tests
2. Run `go build ./...` to ensure no compile errors
3. Run `go vet ./...` to check for issues
4. If anything breaks, revert and replan

---

# PHASE 6: FINAL VERIFICATION

```
go test ./... -race
go vet ./...
go build ./...
```

Then manually exercise the affected code paths and confirm:
- No behavioral regressions
- No new warnings
- The refactor achieves its stated goal

---

**Begin PHASE 0 now.**
