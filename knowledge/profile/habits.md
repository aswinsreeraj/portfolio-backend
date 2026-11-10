# Engineering Habits & Principles

Aswin Sreeraj follows three foundational engineering principles across all backend and distributed systems development:

## 1. Design Before Coding
- **Principle**: Model database schemas, state machines, and API contracts before touching implementation code.
- **Why**: Concurrency bugs, race conditions, and balance discrepancies occur when systems lack unambiguous domain boundaries and explicit state transition diagrams.
- **Real-World Implementation**: In EVNTX, drafting the multi-bucket escrow ledger (`available`, `pending`, `reserve` + platform treasury) and explicit booking/ticket state transitions (`reserved` -> `paid` / `expired`, `valid` -> `used`) prevented split-fee reserve deficits and unguarded mutations.

## 2. Instrument From Day One
- **Principle**: Structured logging, latency/allocation profiling (`go test -bench`, `pprof`), and verifiable execution audit trails on critical paths.
- **Why**: Distributed and concurrent systems fail in subtle performance ways. You cannot optimize or trust what you cannot measure.
- **Real-World Implementation**: In EVNTX, `pprof` profiling guided PostgreSQL pool sizing (`25` max open / `10` idle), `PrepareStmt: true`, `60s` `sync.RWMutex` event caching, and relaxed cron intervals (`*/10`, `*/30`, `*/45`). In ARDA, HNSW benchmarking caught a 0% recall multi-tenant scan bug before production.

## 3. Test Failure Paths First
- **Principle**: Verify failure recovery, expired locks, missing precedent, and security boundaries as first-class paths rather than afterthoughts.
- **Why**: Happy paths are simple; reliability is determined by how safely a system behaves when payments clear late, external services time out, or LLMs hallucinate.
- **Real-World Implementation**: In EVNTX, `VerifyPayment` automatically triggers `RazorpayService.RefundPayment` when captures verify after the 1-minute booking hold expired (`ErrBookingExpiredPaymentSuccess`). In ARDA, the Tool Gateway and Reliability Gate both fail closed on policy breaches or missing HNSW precedent.

