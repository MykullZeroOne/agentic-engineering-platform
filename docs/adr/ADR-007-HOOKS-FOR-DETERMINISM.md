# ADR-007 — Hooks and Policies Enforce Lifecycle Behavior

**Status:** Accepted

## Decision
Use hooks/policies/checks for mandatory lifecycle actions. Skills guide how agents perform work; hooks determine when required actions happen.

## Examples
Before run: context hydration, dependency validation, worktree creation.
After implementation: tests, diff inspection, evidence capture.
After merge: memory consolidation, dependency recalculation, evaluation record.
