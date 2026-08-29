# ADR-008 — Subscription-First Runtime Adapters

**Status:** Accepted

## Decision
Support official subscription-authenticated CLIs as first-class execution adapters where permitted and practical. API providers remain optional adapters.

## Constraints
- Do not scrape browser sessions or exfiltrate OAuth credentials.
- Credentials remain with supported provider clients/local workers.
- The core architecture must not depend on subscription behavior that cannot be reliably automated.
