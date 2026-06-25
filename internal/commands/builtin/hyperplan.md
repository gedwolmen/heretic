---
description: (builtin) Adversarial multi-agent planning via team-mode (5 hostile category members cross-critique, lead synthesizes)
argument-hint: [planning-request]
---

You are leading an adversarial planning session. Your goal is to produce
a plan that has been stress-tested by 5 hostile reviewers from different
perspectives.

## Process

1. **Read the planning request carefully.** What is the user actually
   trying to achieve? Restate it in 1-2 sentences.

2. **Spawn 5 subagents in parallel** (use the `task` tool with category):
   - **explore** (codebase): What existing code is relevant? What
     patterns does this codebase follow?
   - **librarian** (research): What external libraries, RFCs, or
     documentation apply? Are there established solutions?
   - **oracle** (architecture): What's the simplest viable design?
     What are the trade-offs?
   - **metis** (intent): What did the user actually mean? What
     constraints did they not mention but matter?
   - **momus** (review): What could go wrong with the obvious plan?
     What's vague? What needs verification?

   Each subagent returns a short, focused critique or contribution.

3. **Synthesize** the 5 outputs into a single plan. Where they
   disagree, pick the option that best matches the user's stated goal
   and surface the disagreement in the plan.

4. **Write the plan** to `.omo/plans/<slug>.md` using the standard plan
   format (Goal, Steps with verification, Open Questions, Out of Scope).

5. **Self-critique** by re-reading the plan and asking: "If I were
   the user, would I trust this plan to be correct? Would I be
   surprised by anything in it?" If yes, revise.

## RULES

- The 5 subagents must be HOSTILE to the obvious plan. Their job is to
  find holes, not rubber-stamp.
- The synthesized plan must address every criticism raised, even if
  the answer is "considered and rejected because X".
- Do NOT skip steps. If a step is unclear, say so in Open Questions
  rather than guessing.
- The plan must be self-contained. A different agent reading only
  the plan should be able to execute it.

## Output

End with a one-paragraph summary:

> Plan written to <path>. <N> steps, <N> open questions, <N> rejected
> alternatives documented. Confidence: <low/medium/high>.
