# ADR-001 — GitHub as Initial Control Plane

**Status:** Accepted

## Decision
Use GitHub Issues, Projects, Pull Requests, Actions/checks, branches, commits, and Releases as the default authoritative execution state.

## Rationale
GitHub provides open Git semantics, `gh`, APIs/webhooks, CI, issue/PR traceability, broad developer adoption, and compatibility with multiple coding agents. AEP will wrap GitHub behind an SCM/control-provider interface to avoid architectural lock-in.

## Consequences
- Project intent and agent memory remain outside GitHub-specific constructs.
- GitHub adapters must be replaceable.
- Deterministic policy enforcement should use Actions/checks/rules where practical.
