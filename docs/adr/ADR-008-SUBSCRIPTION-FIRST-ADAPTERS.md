---
id: ADR-008
type: adr
tier: 1
status: accepted
version: 1
owner: human.cto
human_approved: true
approved_by: human.cto
approved_on: 2026-08-29
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# ADR-008 — Subscription-First Runtime Adapters

## Decision
Support official subscription-authenticated CLIs as first-class execution adapters where permitted and practical. API providers remain optional adapters.

## Constraints
- Do not scrape browser sessions or exfiltrate OAuth credentials.
- Credentials remain with supported provider clients/local workers.
- The core architecture must not depend on subscription behavior that cannot be reliably automated.
