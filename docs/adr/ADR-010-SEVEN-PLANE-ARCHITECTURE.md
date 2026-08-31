---
id: ADR-010
type: adr
tier: 1
status: accepted
version: 1
owner: human.cto
human_approved: true
approved_by: MykullZeroOne
approved_on: 2026-08-29
approval_record: APR-0002
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# ADR-010 — Seven-Plane Architecture

## Context

`docs/architecture/SYSTEM_ARCHITECTURE.md` decomposes AEP into seven planes, and the entire
documentation corpus, the data model, and the component architecture are organized around them.
It is the top-level structural commitment of the system and had no ADR. Recording it matters
because the decomposition is what keeps provider replaceability (ADR-002, ADR-008) from leaking
into organizational state.

## Decision

AEP is decomposed into seven planes: **Experience, Organization, Communication, Control,
Knowledge, Execution, Evaluation**.

Planes are a decomposition of *responsibility*, not a deployment topology. Two rules give the
decomposition force:

1. **No plane owns another plane's state.** Cross-plane access goes through an explicit port.
2. **Providers attach only to Execution and Control.** Adding or replacing a runtime or an SCM
   must not touch Organization or Knowledge — that is what makes "role != model" structural rather
   than aspirational.

## Alternatives considered

**Conventional layering (UI / service / data).** Rejected. It cannot express that GitHub state,
agent memory, and canonical knowledge have different *authority* and different lifecycles.
Collapsing them into one data layer erases the tier distinctions in
`docs/spec/AUTHORITY_MODEL.md`, which is the thing retrieval and conflict resolution depend on.

**A service per agent role.** Rejected. It couples organizational structure to deployment
topology, so adding a role becomes an infrastructure change — directly contrary to ADR-002, where
roles are durable identities and runtimes are interchangeable.

**Fewer planes — fold Communication into Control and Evaluation into Knowledge.** The most
tempting simplification, and worth stating why it fails. Communication carries independent
authority: an open question or escalation must survive control-plane state changes, so it cannot
be a projection of work state. Evaluation must be able to *replay* Knowledge to rebuild derived
memory (ADR-004), which it cannot do from inside it.

## Consequences

- Package boundaries in the implementation follow plane boundaries (see ADR-011).
- A change touching four or more planes is a signal to re-examine the decomposition before writing
  the code.
- The seven planes are the top-level organization of the documentation corpus, so the docs and the
  packages stay legible against each other.
- `docs/architecture/COMPONENTS.md` names the components; this ADR names the boundaries they may
  not cross.

## Risks

- **Seven planes is a lot of structure for an MVP.** Mitigated by ADR-011: planes are packages in
  a single deployable, not services. The cost is discipline, not operations.
- **A boundary may prove wrong under load.** Accepted. Superseding this ADR is the mechanism, and
  the port-based rule means moving a boundary is a refactor rather than a rewrite.
