# Question and Escalation Workflow

## Resolution order
1. Can the child agent answer from canonical approved context?
2. Can it answer from validated durable memory?
3. Ask parent agent.
4. Parent repeats canonical/memory resolution.
5. If still unresolved and materially required, escalate to parent/human according to authority.
6. Human answer is stored as a durable Decision when consequential.
7. Decision propagates to affected work/spec/context.

## Question object
A question includes source, target, work item, type, severity, blocking flag, question text, reason, affected artifacts, suggested choices, status, timestamps, and resolution/decision link.

## Human notification quality
A human escalation should explain:
- what is being asked;
- why it matters;
- what the system currently knows;
- what work is blocked;
- suggested options/trade-offs;
- which agent/function requested it.
