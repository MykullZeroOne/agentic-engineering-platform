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
6. Working memory — the current task's context: active work item, branch, files, recent
   decisions, failing tests, immediate goal. Distinct from episodic because it is *current*,
   and distinct from a context packet because it outlives a single retrieval.
7. Sensory memory — short-lived raw observation: tool output, terminal logs, file changes,
   webhook payloads. Mostly discarded. It is named because "raw session history must remain
   distinct from consolidated memory" needs a name for the thing being kept distinct.
8. Organizational memory — reviewers, ownership, team conventions, external-system workflows,
   policy constraints. Facts about how *this organization* works rather than how the code works.

Layers 6–8 are absorbed from Kairo's `docs/v2/08-memory` (ADR-022). Its typology has six types
where this had five; neither was a superset, and the union is the better model. Performance
memory is AEP's and has no Kairo equivalent.

## Memory lifecycle

```
Observe → Filter → Capture → Summarize → Index → Retrieve → Refresh → Retain/Delete
```

Filter precedes capture deliberately. A secret that is captured and then redacted has already
been written down.

## Decay, weight, and retention

Not every memory deserves equal retrieval weight forever, and this was the half of memory this
document did not specify. Five mechanisms, which are independent and combine:

- **Recency decay** — retrieval weight falls with age, per layer. Canonical knowledge does not
  decay; sensory memory decays fastest.
- **Importance scoring** — an explicit weight, set at consolidation and revisable.
- **Pinning** — a human may hold a memory at full weight indefinitely. Pinning is a human act
  and is recorded as one.
- **Retention policy** — how long a layer is kept at all, distinct from how heavily it ranks.
  Deletion must be possible without invalidating an artifact that points at the memory.
- **Archival** — removal from retrieval without removal from the record.

Supersession already exists in this document and is not a decay mechanism: a superseded memory
is wrong, not old.

## Context budget

Model context is scarce, and retrieval that ignores that is retrieval that crowds out the task.
The retriever ranks candidates by expected utility and includes only what fits a declared
budget. What it *excluded* is reported, not silently dropped — see PRD-006.

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

A second criterion, for the half added above: a memory that has decayed, been archived, or been
excluded for budget can be shown to have been *considered and not selected*, distinguishable from
one that was never a candidate. A retriever that cannot tell those apart cannot be tuned.
