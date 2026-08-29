# Memory Architecture

## Goal
Provide durable institutional continuity without treating a model conversation as long-term memory.

## Storage domains

### Raw event history
Immutable evidence of what happened: sessions, messages, tool calls, commands, Git events, context retrievals, tests, reviews, questions, answers, human interventions, merges, releases.

### Consolidated memory
Derived lessons and reusable knowledge. Memory is not authoritative truth unless promoted/approved.

### Canonical knowledge
Approved PRDs, ADS, ADRs, standards, architecture, policies, and explicit human decisions.

### Evaluation data
Versioned run metadata and outcomes suitable for regression and future training datasets.

## Memory types
- Semantic — facts/conclusions.
- Episodic — prior situations, attempts, outcomes, corrections.
- Procedural — how the organization normally performs work.
- Performance — which agents/models/skills/routes work well for which tasks.

## Scope dimensions
- organization
- project
- organizational function
- role
- durable agent identity
- domain/component
- work item

## Retrieval inheritance
An agent receives a weighted projection of:
1. canonical task knowledge;
2. its own durable memory;
3. role/function memory;
4. parent/org decisions;
5. project shared memory;
6. semantically/graph-related experience.

## Supersession
Memories can be active, challenged, superseded, expired, or promoted to standard. Canonical decisions supersede conflicting memory.
