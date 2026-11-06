# Project: ARDA (Auditable, Reliable, Durable Agents)

- **ID**: `arda`
- **Tagline**: AI Agent Execution & Governance Platform
- **Status**: In Progress
- **Key Metrics**: P99 Latency 18ms | Sustained Load 1,200 rps | Failure Recovery 100%
- **GitHub**: https://github.com/aswin/arda
- **Demo / Docs**: https://arda.dev

## Overview & Motivation
ARDA provides a shared Go runtime for multi-step AI agents requiring bounded tool access, versioned execution budgets, and durable workflow guarantees. It replaces fragile in-memory agent loops with immutable state checkpointing and human-in-the-loop review queues.

## Key Architecture & Components
1. **API Gateway & Access Control**: Tenant-aware gRPC / REST API with JWT authentication and role-based access control (RBAC).
2. **Durable Workflow Orchestrator**: Uses the **Temporal** SDK to execute multi-step workflows. Every activity completion is persisted in a deterministic event history, ensuring workflows automatically resume from their last completed step after worker crashes or network timeouts.
3. **Tool Governance Sandbox**: Enforces Model Context Protocol (MCP) tool integration with strict read vs write tool classification, deterministic business-rule guardrails, and JSON Schema validation for tool inputs and outputs.
4. **Reliability & Confidence Scorer**: Implements retrieval-based reliability scoring with embeddings and **pgvector**, comparing agent responses against historic high-confidence reference solutions. Fails closed and routes ambiguous outputs to a human review queue.
5. **Event Streaming & Datastore**: PostgreSQL 16 persistence for tenant configs, immutable execution traces, and audit logs. Redpanda / Kafka event streams driven by the Transactional Outbox pattern to prevent data drift.

## Engineering Decisions & Trade-Offs
- **Temporal Durable Workflows over Ephemeral In-Memory Loops**: In-memory agent loops fail irrevocably when external LLM API calls time out or worker pods restart. Temporal persists every activity completion as a deterministic event history, enabling zero-loss recovery.
- **Transactional Outbox Pattern for Redpanda**: Dual writes to Kafka and PostgreSQL introduce distributed consistency risks if the network fails between operations. The transactional outbox writes events to PostgreSQL atomically, where an outbox worker safely relays them to Redpanda.
- **PostgreSQL + pgvector over SaaS Vector Store**: Colocates relational agent metadata, tenant configurations, and vector embeddings in a single ACID datastore, simplifying transactions, point-in-time recovery, and backups.

## What Went Wrong & Lessons Learned
- **Context Deadline Poisoning on gRPC Streaming Workers**: During stress tests with network jitter, agent tool workers leaked context timers because client cancellation signals did not propagate to nested Temporal activity child contexts. The fix was explicitly binding child contexts with `context.WithCancelCause` and listening on `<-ctx.Done()` in all MCP transport adapters.
- **Key Takeaway**: Never assume context cancellation cascades implicitly across network protocol boundaries without verified propagation.

