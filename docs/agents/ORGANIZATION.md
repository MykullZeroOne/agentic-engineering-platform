---
id: DES-ORGANIZATION
type: design
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

# Agent Organization

## Executive layer

### Human CTO / Product Owner
Final authority for product intent, material architecture decisions, legal approval, risk acceptance, production/release gates, and organizational policy.

The human should operate primarily through questions, approvals, exceptions, and strategy—not code implementation.

## Product organization

### BA / Product Analyst Lead
Owns product-intent refinement and creation of approved PRD + ADS.

Specialists:
- Domain Analyst
- Compliance Analyst
- Legal Preflight Analyst
- Privacy Analyst
- UX/Product Flow Analyst

### Requirements / Grooming Lead
Owns requirement readiness and decomposition quality.

Specialists:
- Testability Analyst
- Dependency Analyst
- Edge-case Analyst
- UX/Accessibility Requirements Analyst

## Technical organization

### Architecture Lead
Owns technical impact, architectural fit, new decisions, constraints, and architecture readiness.

Specialists:
- Security Architect
- Data Architect
- Integration Architect
- Platform/Infrastructure Architect
- Performance/Reliability Architect

### Planning Lead
Owns executable work graph and verification plan.

Specialists:
- Dependency Planner
- Parallelization Planner
- Complexity/Risk Estimator
- Release Sequencing Analyst

### Orchestrator
Executes the approved DAG deterministically, selects capable workers, manages concurrency, state, retries, and handoffs. The orchestrator should be less improvisational than upstream analytical agents.

## Delivery organization

Lead/worker roles may include:
- iOS Engineer
- Backend Engineer
- Web Engineer
- Data Engineer
- Platform Engineer
- DevOps Engineer
- Documentation Engineer

Each role may spawn narrowly scoped sub-agents for implementation, refactoring, tests, migrations, documentation, or investigation.

## Quality organization

### QA Lead
Owns verification against requirements, not merely test execution.

Specialists:
- Functional QA
- Integration QA
- Security QA
- Accessibility QA
- Performance QA
- Regression QA

### Review Lead
Independent review against requirement correctness, architecture, scope, security, maintainability, standards, and test sufficiency.

### Integration Lead
Combines parallel work and validates system-level behavior before release readiness.

## Release organization

### Release Lead
Specialists:
- Deployment
- Migration
- Rollback
- Observability
- Release Notes

Release Lead prepares evidence; human owns configured high-risk release gates.
