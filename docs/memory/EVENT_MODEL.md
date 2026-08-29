---
id: DES-EVENT-MODEL
type: design
tier: 2
status: draft
version: 1
owner: human.cto
human_approved: false
approved_by: null
approved_on: null
supersedes: null
superseded_by: null
last_reviewed: 2026-08-29
---

# Event Model

AEP stores immutable domain events so derived knowledge can be replayed and rebuilt.

## Representative events
- ProjectRegistered
- AgentIdentityCreated
- RoleVersionChanged
- AgentSessionStarted
- ContextItemRetrieved
- ContextPacketBuilt
- QuestionAsked
- QuestionEscalated
- AnswerReceived
- DecisionRecorded
- SpecialistDelegated
- ToolInvoked
- FileRead
- FileChanged
- CommandExecuted
- TestCompleted
- CIFailed
- ReviewFindingCreated
- HumanIntervened
- RuntimeHandedOff
- PullRequestOpened
- PullRequestMerged
- MemoryCandidateCreated
- MemoryPromoted
- MemorySuperseded
- RegressionRunCompleted

## Event envelope
Every event should include:
- event ID/type/version;
- timestamp;
- organization/project;
- durable agent identity and runtime session if applicable;
- correlation IDs for work item/run;
- source system;
- payload;
- provenance/integrity metadata.
