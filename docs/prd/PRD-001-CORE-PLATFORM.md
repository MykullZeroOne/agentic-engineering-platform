# PRD-001 — Core Agentic Engineering Platform

## Objective
Provide the reusable platform runtime, UI, configuration, and integration foundation for a vendor-neutral agentic engineering organization.

## Primary capabilities
- Register and manage projects/repositories.
- Model durable agent identities, roles, parents, specialists, skills, permissions, and completion gates.
- Integrate GitHub as the initial control plane.
- Connect runtime adapters for Claude Code, Codex, and future agents.
- Provide agent workspace observability and direct human interaction.
- Provide shared hook, policy, skill, context, and memory services.
- Support new-project bootstrap and existing-project adoption.

## Acceptance criteria
1. A new GitHub project can be initialized with one command or UI flow.
2. An existing repository can be scanned and receive an adoption plan without destructive changes.
3. A durable agent can launch multiple runtime sessions and retain identity and memory across them.
4. A human can observe, message, pause, resume, cancel, and hand off an active agent runtime.
5. The runtime provider can be changed without changing the role contract.
6. GitHub remains authoritative for issue/PR/check/release state.
