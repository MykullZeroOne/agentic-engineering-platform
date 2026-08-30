---
id: ADR-018
type: adr
tier: 1
status: accepted
version: 1
owner: human.cto
human_approved: true
approved_by: MykullZeroOne
approved_on: 2026-08-30
supersedes: null
superseded_by: null
last_reviewed: 2026-08-30
---

# ADR-018 — A GitHub App Is the Control-Plane Adapter

## Context

ADR-001 made GitHub the control plane and said AEP would wrap it behind a provider interface. It
did not say how AEP authenticates. The default path — a personal access token, or the `gh` CLI —
has a defect that only becomes visible once agents act autonomously:

**AEP operating on a human's token is indistinguishable from the human.** Every commit, comment,
check and review in GitHub's audit trail is attributed to the token holder. There is then no way,
from the system of record, to separate "the human approved this" from "an agent acted on the
human's credentials." Principle 1 and the approval provenance of ADR-013 exist precisely to keep
those separable, and the SCM layer would collapse them again.

A second constraint is concrete: the Checks API is available to GitHub Apps and not to bare
tokens. Any ambition to surface AEP's gates as check runs that branch protection enforces
requires an App.

## Decision

**A GitHub App is the primary control-plane adapter. Token and `gh` authentication remain a
supported local-development implementation of the same interface.**

1. **AEP acts as itself.** The App is a distinct actor, so agent action and human judgment stay
   distinguishable in GitHub's own record.
2. **Permissions are installation-scoped and least-privilege**, per repository, mapping onto the
   `tools.allow` / `deny` model in `agent-role.example.yaml` and the least-privilege requirement in
   `GOVERNANCE.md`. A token that can do everything the human can makes role-scoped permissions a
   fiction at the GitHub boundary.
3. **Webhooks are the event source**, feeding the immutable event log of ADR-004 rather than
   polling.
4. **Gates may be published as check runs.** This is what moves gate enforcement from `prose` to
   `check` on ADR-014's enforcement ladder, and it is unavailable without an App.
5. **The App lives inside the adapter.** GitHub concepts do not leak into the Organization,
   Knowledge or Communication planes. Replacing GitHub means writing another adapter, per ADR-010.
6. **Not built yet.** There is no service to hold a private key or receive a webhook. This decides
   the target so the adapter interface is written for it, with token/`gh` as the first
   implementation.

### ADR-008 does not conflict with this

ADR-008 prefers official CLIs over raw APIs, which appears to point at `gh`. It does not apply
here. ADR-008 governs **model runtimes** and exists to avoid unpredictable API-metered LLM spend;
it says nothing about SCM access, and `gh` carries the identity defect above. Stated explicitly so
the preference is not misapplied later.

## Alternatives considered

**Personal access token or `gh` CLI as the primary path.** Rejected on identity. Simplest to build
and adequate for a single human running things by hand, but it makes agent action indistinguishable
from human action in the audit trail, forecloses the Checks API, offers no per-repo permission
scoping, and consumes the human's rate limit. It stays as the local-development implementation
because it needs no server.

**OAuth App.** Rejected. It acts on behalf of a user, which is the same identity collapse in a
different wrapper, and it is designed for user login rather than autonomous service action.

**GitHub Actions as the integration surface.** Rejected as the primary mechanism. Actions run
inside GitHub's CI and are well suited to deterministic checks, but they cannot host a
long-running orchestrator, cannot supervise provider CLIs, and would invert the control
relationship AEP requires. Actions remain useful for the checks themselves.

**Deep GitHub integration with no abstraction.** Rejected as a direct contradiction of ADR-001 and
ADR-010, however tempting the ergonomics.

## Consequences

- **Webhooks require a reachable endpoint**, which `DEPLOYMENT.md` does not currently provide — it
  describes a local worker daemon. Either a hosted control service arrives earlier than planned, a
  relay is used for development, or the adapter falls back to polling and loses event immediacy.
  This is a real architectural pull, not a detail.
- Private key custody and installation-token minting become operational concerns.
- The adapter interface must be authored so both App and token implementations satisfy it, with no
  caller aware of which is active.
- If AEP is ever installed in another organization, the App is already the distribution mechanism.
- Rate limits become the App's rather than the human's.

## Risks

- **Behavioral lock-in.** Technical portability is preserved by the adapter, but the richer the
  integration, the more the team's habits assume GitHub semantics. That erosion is invisible until
  a second provider is attempted, and no interface prevents it.
- **The hosted-endpoint pull may compromise the local-first model.** ADR-008 keeps provider
  credentials on local workers; a webhook receiver wants a public address. Those are not
  contradictory but they push the deployment shape in opposite directions.
- **Not every GitHub surface supports App auth equally.** Projects v2 and some GraphQL operations
  have had gaps. Each capability the adapter depends on needs verifying rather than assuming.
- **Two implementations to keep honest.** A token path that drifts from the App path produces
  behavior that works locally and fails in production, which is the worst failure shape.
