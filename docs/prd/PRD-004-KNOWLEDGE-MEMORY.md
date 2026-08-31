---
id: PRD-004
type: prd
tier: 2
status: draft
version: 1
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
approval_record: null
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# PRD-004 — Knowledge, Memory, and Institutional Learning

## Objective
Give every durable agent continuity across ephemeral sessions while maintaining shared organizational knowledge, provenance, and cross-agent learning.

## Memory layers
1. Canonical knowledge — approved PRDs, ADS, ADRs, architecture, standards, policies.
2. Semantic memory — durable facts/conclusions learned from work.
3. Episodic memory — what happened during prior work and why.
4. Procedural memory — reusable methods and preferred procedures.
5. Performance memory — observed effectiveness of models, skills, routes, prompts, and workflows.

## Requirements
- Memories are logically scoped by project, organization, role, agent identity, domain, and visibility.
- An agent's context includes its own memory, inherited parent/org knowledge, task context, and selectively related memories.
- Raw session history must remain distinct from consolidated memory.
- Memories require provenance, confidence, creation source, validity status, and supersession links.
- Approved decisions outrank learned memories.
- Session close hooks consolidate useful outcomes; session start hooks hydrate concise context.
- Cross-agent success may become a shared pattern after validation.

## Acceptance criteria
A future Compliance session can answer a previously resolved compliance question without loading the original chat transcript and can cite the decision/evidence that established the answer.
