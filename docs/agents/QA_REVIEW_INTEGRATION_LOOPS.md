# QA, Review, and Integration Loops

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

## Review Lead
Independently reviews implementation against:
- approved requirement and scope;
- architecture/ADR constraints;
- security and privacy policy;
- standards;
- maintainability;
- test sufficiency;
- known organizational lessons/patterns.

## Integration Lead
Validates combined work from parallel branches/PRs. Passing individual work units does not imply passing integration. Integration Lead owns system-level build, conflict resolution routing, cross-component verification, and release candidate evidence.
