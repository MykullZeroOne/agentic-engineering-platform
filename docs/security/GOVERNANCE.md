---
id: DES-GOVERNANCE
type: design
tier: 2
status: draft
version: 1
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# Governance and Safety Boundaries

## Human-only or explicit approval domains
Configurable defaults should include:
- production deployment/rollback;
- production secrets;
- destructive data changes;
- legal approval;
- material privacy/risk acceptance;
- consequential architecture decisions;
- security-policy changes;
- billing/payment configuration.

## Specialist legal/compliance agents
These agents provide issue spotting, preflight analysis, requirements suggestions, evidence organization, and questions. They do not represent licensed legal advice or final organizational approval.

## Least privilege
Agents receive role-scoped permissions. A Compliance Agent generally does not need deployment or source-write access. An iOS agent generally does not need infrastructure secrets.

## Auditability
Material decisions, human interventions, permission changes, agent handoffs, and privileged tool usage must be durable events with provenance.
