# Standards, Policies, Skills, and Hooks

## Standard
Defines what good engineering means. Example: all external API calls require bounded timeout behavior.

## Policy
Defines a mandatory rule. Example: changes under `migrations/**` require rollback plan and human approval.

## Skill
Defines the reusable procedure for performing work. Example: `implement-api-endpoint`.

## Hook
Defines the lifecycle boundary at which a policy/procedure is invoked.

## Agent
Executes the skill within policy and context.

## Example
Standard: authorization enforced server-side.
Policy: API changes require integration authorization tests.
Skill: backend-endpoint.
Hook: before_pr -> run authorization policy checks.
Agent: Backend Engineer (Codex/Claude runtime).
