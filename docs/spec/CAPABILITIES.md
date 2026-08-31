---
id: SPEC-CAPABILITIES
type: spec
tier: 0
status: approved
version: 2
owner: human.cto
human_approved: true
approved_by: MykullZeroOne
approved_on: 2026-08-31
approval_record: APR-0014
supersedes: null
superseded_by: null
last_reviewed: 2026-08-31
enforcement: check
---

# Capabilities and Grant Semantics

What an agent may do, and how a grant is evaluated. The namespace itself is the `capability`
vocabulary in `.agentic/registries/vocabularies.yaml`; this document defines the rules.

## Why this was urgent

`production_secrets` and `billing_configuration` both triggered on *"an agent requests a
capability"* against a namespace with no defined membership. **Neither security gate could fire**,
because nothing said which requests counted. `agent-role.example.yaml` had been granting
`github.read`, `repository.write` and `deployment.*` against a vocabulary that did not exist.

## The namespace is closed

A capability is `<domain>.<action>`, and both parts come from the registry. A grant naming a token
that is not there is an **error**, never a silently ignored no-op — silent tolerance of unknown
capabilities is how a typo becomes an unnoticed permission gap.

Domains: `github` · `repository` · `workspace` · `runtime` · `knowledge` · `memory` · `reference` ·
`communication` · `deployment` · `secrets` · `billing` · `data`

## Principals

A capability is held by a **principal**. This document specified grant *evaluation* without
saying what holds a grant, which left "role != model" (Principle 2) expressed in prose and
nowhere in the data. Absorbed from Kairo's `docs/v2/17-security/auth-rbac.md` per ADR-022,
narrowed: Kairo's organization and billing tiers are dropped, because AEP has no tenancy.

| Principal | Represents |
| --- | --- |
| **User** | A human account. |
| **Team** | A named group of users. |
| **Token** | A machine credential held by a CLI, an agent runtime, or an integration. |

**An agent is not a principal type.** It is a **token** whose kind is `agent`, carrying an
effective role and a scope. This is the structural form of Principle 2: a role is a durable
identity, a runtime session is transient, and a token is what a transient thing holds.

### `on_behalf_of`

A token may act `on_behalf_of` a user. That attribution is **carried on every event and checked
for audit — never for authorization.** A token acting for an owner does not thereby gain the
owner's permissions; it keeps its own.

Separating the two is the whole point. Collapsing them is precisely the identity collapse
ADR-018 rejects a personal access token for, one layer up.

**This is the missing half of the independent-review problem.** CLAUDE.md records that review is
suspended because "this repository has one human identity, and agents act as that identity, so
GitHub cannot tell author from reviewer" (WI-0016). With a principal model, *AEP's own records*
can tell them apart: author is a token, reviewer is a different token or a user, and
`on_behalf_of` says which human stands behind each. It does not fix GitHub's view — ADR-018's
App is what makes GitHub see a distinct actor. The two together are the answer; neither alone is.

### Scope

A grant is assigned at a scope, and the effective grant for an action is the **most specific
assignment covering the target**. Absent an assignment there is no grant, per rule 3 below.

Scope is also what answers the third known limit at the foot of this document: whether reading
another agent's memory across a namespace boundary needs its own capability. It does not need a
new capability; it needs the read to be outside the token's scope, which the existing
`memory.*` tokens then evaluate normally.

## Grant evaluation

A role or agent identity carries `tools.allow` and `tools.deny`, each a list of tokens or domain
wildcards.

**1. Deny beats allow, always.** Not by specificity, not by order. A capability in `deny` is
denied even if `allow` names it exactly. There is no override, because the alternative is a
precedence puzzle nobody can evaluate under pressure.

**2. Wildcards expand at grant time, not at check time.** `deployment.*` resolves to the concrete
tokens present in the registry *when the grant is written*, and the expansion is stored.

This is the rule that matters most and it is the least obvious. Expanding at check time would mean
that **adding a capability to the registry silently widens every existing wildcard grant** — a new
`deployment.destroy` would be granted to everyone already holding `deployment.*`, with no change to
any role and nothing to review. Expanding at grant time makes that a visible diff instead.

**3. Absent is denied.** No grant means no capability. There is no implicit baseline.

**4. A gated capability requires its gate.** A capability with `gated_by` may not be exercised
until that gate is closed by an approval record, *even when granted*. The grant makes it
reachable; the gate makes it permitted. Seven capabilities are gated:

| Capability | Gate |
| --- | --- |
| `deployment.deploy`, `deployment.rollback` | `production_release` |
| `secrets.read`, `secrets.write` | `production_secrets` |
| `billing.read`, `billing.write` | `billing_configuration` |
| `data.destructive` | `destructive_data_change` |

That table is what makes those gate triggers evaluable. Before it, they named a condition no
system could test.

## Least privilege is the default

`GOVERNANCE.md` requires role-scoped permissions: a Compliance agent needs no deployment access, an
iOS agent no infrastructure secrets. With an open namespace that was aspiration. With a closed one
it is checkable — a role's grants can be compared against the capabilities its work actually
requires.

## Known limits

- **Grants are not yet enforced at runtime.** There is no runtime. This defines the semantics an
  implementation must honour; today it constrains what a role definition may *say*, which is
  enforcement stage `check` rather than `hook`.
- **`reference.lookup` names a capability whose content has no authority tier.** External
  documentation is not project knowledge and must not be conflated with it. Recorded as a known
  gap in `REGISTRIES.md`.
- ~~**No capability covers reading another agent's memory across a namespace boundary.**~~
  Resolved by scope, above: a cross-namespace read is the same `memory.*` capability evaluated
  against a scope that does not cover the target. No new capability is required.

- **Principals are specified and not stored.** There is no user, team or token record anywhere in
  `.agentic/`, so the model above constrains what a role definition may *say* and nothing more.
  Where those records live is a `platform_config` decision this document does not make.

- **A role's `kind` is not yet in the vocabulary registry.** `agent`, `mcp` and the human kinds
  are named here and are not tokens in `vocabularies.yaml`, which makes them prose. Closing that
  is a `platform_config` change.
