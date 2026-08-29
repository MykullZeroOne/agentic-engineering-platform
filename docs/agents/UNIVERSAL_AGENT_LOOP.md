# Universal Agent Loop

Every lead agent uses the same recursive state machine.

1. **Receive** — accept a work assignment with explicit objective and authority boundary.
2. **Hydrate** — load role identity, task, canonical context, inherited decisions, and role-specific memory.
3. **Analyze** — identify knowns, unknowns, risks, missing evidence, and applicable specialists.
4. **Delegate** — invoke specialist sub-agents where their expertise materially improves confidence.
5. **Resolve** — answer specialist questions from canonical knowledge or durable decisions when possible.
6. **Escalate** — route unresolved questions to parent; parent repeats resolution logic and escalates to human only when required.
7. **Re-evaluate** — incorporate answers/findings and determine whether more questions or specialists are needed.
8. **Validate gate** — evaluate role-specific definition of done/readiness.
9. **Produce output** — create structured artifacts/evidence.
10. **Handoff** — pass output and unresolved approved exceptions to the next owning role.
11. **Consolidate** — record run outcome and memory candidates.

## Loop invariants
- Do not finalize with unresolved critical unknowns.
- Do not ask the human a question already answered by authoritative context.
- Do not silently infer a consequential product/legal/security decision.
- Specialist findings return to their owning parent.
- Every material conclusion has evidence/provenance.
