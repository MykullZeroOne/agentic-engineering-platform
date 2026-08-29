---
id: PRINCIPLES
type: principle
tier: 0
status: approved
version: 1
owner: human.cto
human_approved: true
approved_by: MykullZeroOne
approved_on: 2026-08-29
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# Product and Engineering Principles

## Human authority
Human is the highest authority. Agents may recommend, challenge, and request clarification, but they do not silently override human-approved intent or policy.

## Organizational responsibility
Every work item has an owning lead agent. Specialist sub-agents report to that owner. Cross-functional communication is mediated through accountable parents rather than unrestricted agent-to-agent chatter.

## Escalate only when necessary
Before asking the human, an agent must attempt to answer from approved specifications, canonical knowledge, durable decisions, its own memory, and its parent agent.

## Deterministic boundaries
Use hooks, CI, policies, and state machines for deterministic behavior. Use LLM reasoning for ambiguity and synthesis, not as the sole enforcement mechanism.

## Evidence over assertion
Requirements are satisfied by evidence: tests, checks, reviews, artifacts, decisions, and traceable implementation relationships.

## Durable identity, ephemeral runtime
An organizational agent identity persists across sessions and may change runtime/model without losing memory, responsibilities, or history.

## Memory is not truth
Canonical approved knowledge outranks learned memory. Learned memory must carry source, confidence, scope, and supersession information.

## Open interfaces
Favor Git, `gh`, MCP, Agent Skills, CLI adapters, webhooks, standard schemas, and portable files.
