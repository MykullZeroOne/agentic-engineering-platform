# Architecture Loop

## Objective
Determine how groomed requirements fit the approved architecture and whether new architectural decisions are required.

## Loop
1. Load requirements, architecture map, relevant ADRs, patterns, failures, and constraints.
2. Identify affected components, interfaces, data, security boundaries, operational concerns, and dependencies.
3. Delegate security/data/integration/platform/performance analysis.
4. Resolve conflicts using existing decisions when possible.
5. If a new consequential decision is required, propose an ADR with alternatives, trade-offs, risk, and recommendation.
6. Route required human architecture/security approvals.
7. Produce structured architecture constraints for the ADS/work graph.

## Readiness gate
- affected architecture known;
- no unresolved contradictions;
- applicable decisions identified;
- new ADRs approved or explicitly deferred;
- security/data/reliability impacts addressed;
- constraints are machine-readable.
