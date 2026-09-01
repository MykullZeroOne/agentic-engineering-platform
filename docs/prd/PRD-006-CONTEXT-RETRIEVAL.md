---
id: PRD-006
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

# PRD-006 — Context Retrieval and Context Packets

## Objective
Deliver concise, role-specific, evidence-backed context to agents automatically before work begins.

## Context packet composition
- agent identity and role contract;
- current work item and parent chain;
- approved PRD/ADS requirements;
- applicable ADRs and policies;
- relevant architecture/components;
- dependency state;
- acceptance criteria and verification requirements;
- role-specific memories and lessons;
- similar prior implementations;
- known pitfalls;
- provenance for each included item;
- the context budget: requested and estimated size;
- **omissions and confidence** — what was considered and left out, and how sure the retriever is
  that what remains is sufficient.

## Reconstruction, not replay

Context is *reconstructed* for the present task, not replayed from a prior session. Replay
returns a previous state; reconstruction assembles the memory this task needs, which is a
different and usually smaller set. Absorbed from Kairo's `docs/v2/12-resume` (ADR-022), which
draws the same line as "resume is not restore".

The distinction has a practical consequence: a packet is not judged by how faithfully it
reproduces a prior context, but by whether the work succeeds with it.

## Declared omissions

**A packet states what it left out.** This is the requirement most likely to be dropped as
overhead and is the one that makes the rest trustworthy.

A packet listing only its contents is indistinguishable from a packet that found nothing else —
so an agent cannot tell "there is no prior attempt at this" from "prior attempts were dropped for
budget". The first is a fact about the work; the second is a fact about the retriever, and only
one of them should change what the agent does.

This is the same discipline as the evidence package's required `open` block
(`docs/spec/EVIDENCE_PACKAGE.md`): absence stated, never implied.

## Retrieval principles
- canonical knowledge first;
- explicit links before semantic similarity;
- graph proximity + semantic relevance + role relevance + recency + confidence;
- concise conclusions first, deeper evidence on demand;
- do not flood the context window;
- record what was retrieved and why.

## Acceptance criteria
Every substantial agent run has a reproducible context manifest showing exactly which knowledge items were supplied and why.
