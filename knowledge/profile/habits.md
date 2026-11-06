# Engineering Habits & Principles

Aswin Sreeraj follows three foundational engineering principles across all backend and distributed systems development:

## 1. Design Before Coding
- **Principle**: Model database schemas, state machines, and API contracts before touching implementation code.
- **Why**: Concurrency bugs, race conditions, and data corruption occur when systems lack unambiguous domain boundaries and explicit state transition diagrams.
- **Real-World Implementation**: In EVNTX, drafting the multi-party wallet ledger and explicit booking state transitions (`PENDING` -> `RESERVED` -> `CONFIRMED` / `EXPIRED`) eliminated concurrency bugs and race conditions before writing handler logic.

## 2. Instrument From Day One
- **Principle**: Structured JSON logging with Go's `slog`, OpenTelemetry traces, and Prometheus metrics on every synchronous and asynchronous path.
- **Why**: Distributed systems fail in silent and surprising ways. You cannot optimize or debug what you cannot measure.
- **Real-World Implementation**: In ARDA, distributed tracing identified tool execution bottlenecks and token quota exhaustion in real time across multi-step LLM workflows.

## 3. Test Failure Paths First
- **Principle**: Write integration tests for payment verification recovery, expired distributed locks, network partitions, and worker crashes before the happy path.
- **Why**: Happy paths are simple; reliability is determined entirely by how gracefully a system recovers from network timeouts, duplicate webhooks, and unexpected crashes.
- **Real-World Implementation**: In EVNTX, tests verified automated refunds and inventory rollback when Razorpay payment webhooks arrived after reservation lock expiration.

