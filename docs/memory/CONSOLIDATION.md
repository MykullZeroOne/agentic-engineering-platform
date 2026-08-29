# Memory Consolidation

## Session-start hook
1. Load agent identity and role version.
2. Load current work and parent chain.
3. Load active/open questions and decisions.
4. Retrieve role-specific memory and related project knowledge.
5. Build concise context packet.

## Session-end hook
Summarize:
- decisions made;
- assumptions created/invalidated;
- questions opened/resolved;
- attempts and outcomes;
- failures and corrections;
- useful procedures/patterns;
- human interventions;
- remaining next actions.

Do not promote every observation into durable memory.

## Post-merge consolidation
Compare:
- intended requirements;
- implementation plan;
- actual diff;
- test/CI failures;
- QA/review findings;
- human corrections;
- final merged implementation.

Generate memory candidates such as:
- reusable lesson;
- anti-pattern;
- preferred procedure;
- obsolete memory;
- candidate standard;
- candidate skill improvement.

## Promotion ladder
Observation -> Candidate Memory -> Validated Pattern -> Procedure -> Project/Org Standard/Skill.

Promotion criteria may include repeat successful use, absence of counterexamples, review approval, and human approval for high-impact standards.
