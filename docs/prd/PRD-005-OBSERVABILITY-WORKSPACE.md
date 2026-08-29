# PRD-005 — Agent Workspace and Observability

## Objective
Eliminate opaque agent execution by giving users a live, inspectable, interruptible view into every agent.

## Agent workspace tabs
- Activity — normalized event timeline.
- Live Session — actual CLI/runtime stream and interactive input.
- Context — exact supplied context and provenance.
- Memory — relevant persistent memory and role knowledge.
- Tools — capabilities and permissions.
- Work — issue, plan, branch/worktree, PR, tests, dependencies.

## Controls
- Message
- Pause
- Resume
- Cancel
- Restart
- Handoff to another runtime/model
- Escalate

## Acceptance criteria
A human can open a running Codex developer session, see what it is reading/running, send a corrective instruction, and preserve that intervention as structured telemetry/memory candidate data.
