---
id: DES-QA-REVIEW-INTEGRATION-LOOPS
type: design
tier: 2
status: draft
version: 2
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
approval_record: null
supersedes: null
superseded_by: null
last_reviewed: 2026-09-06
---

# QA, Review, and Integration Loops

Each role below is a configuration of the universal agent loop (ADR-023,
`UNIVERSAL_AGENT_LOOP.md`).

## QA Lead
Starts from requirements and acceptance criteria. Its primary question is: **what evidence proves the requirement is satisfied?**

Specialists execute functional, integration, security, accessibility, performance, and regression verification as required.

A QA finding links:
- requirement/acceptance criterion;
- evidence examined;
- actual behavior;
- expected behavior;
- severity;
- recommended owning role.

### Configuration of ADR-023 (QA Lead)
- **Step-4 specialists**: Functional QA, Integration QA, Security QA,
  Accessibility QA, Performance QA, Regression QA.
- **Step-7 gate**: agent-evaluated. Passes when every requirement/acceptance
  criterion has linked evidence of the form above (requirement/AC, evidence
  examined, actual vs. expected behavior, severity, recommended owning role)
  and no finding carries an unresolved severity.
- **Step-8 artifacts**: QA findings, in the form above.
- **Return sources**: none documented.

## Review Lead
Independently reviews implementation against:
- approved requirement and scope;
- architecture/ADR constraints;
- security and privacy policy;
- standards;
- maintainability;
- test sufficiency;
- known organizational lessons/patterns.

### Configuration of ADR-023 (Review Lead)
- **Step-4 specialists**: none. The review dimensions above (requirement/scope,
  architecture/ADR, security/privacy, standards, maintainability, test
  sufficiency) are step-3 analysis axes, not specialists.
- **Step-7 gate**: agent-evaluated. Passes when the implementation is checked
  against every dimension above with no unresolved finding.
- **Step-8 artifacts**: review findings against the dimensions above.
- **Return sources**: none. Review returns work to Engineering; nothing returns
  work to Review.

## Integration Lead
Validates combined work from parallel branches/PRs. Passing individual work units does not imply passing integration. Integration Lead owns system-level build, conflict resolution routing, cross-component verification, and release candidate evidence.

### Configuration of ADR-023 (Integration Lead)
- **Step-4 specialists**: none.
- **Step-7 gate**: agent-evaluated, unless a release gate triggers on the
  outcome. Its cross-component verification is step 7 Validate: passes when
  system-level build, conflict resolution, and cross-component verification are
  complete with release candidate evidence produced.
- **Step-8 artifacts**: release candidate evidence.
- **Return sources**: none documented.
