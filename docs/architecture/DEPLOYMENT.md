# Deployment Architecture

## MVP deployment
- AEP server: single service or modular monolith.
- PostgreSQL: application state, events, memories, graph tables, pgvector.
- Object/artifact storage: optional local/S3-compatible storage for transcripts, test logs, large context snapshots.
- Local worker daemon: runs on developer Mac/Linux hosts and launches authenticated Claude/Codex CLIs.
- GitHub: external control-plane authority for repository workflow.

## Later separation
Split runtime manager, context/memory, event/evaluation, and webhook/control services only when load/operational needs justify it.

## Subscription worker model
A local worker registers capabilities with AEP but credentials remain on the local host. AEP dispatches work to the worker, not credentials to the cloud.
