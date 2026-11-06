# Technical Notes: Failure Postmortems & Production Lessons

Real-world production failure scenarios analyzed and solved by Aswin:

## 1. Concurrency Deadlocks on Multi-Row Updates
- **Symptom**: Intermittent database queries crashing with `pq: deadlock detected (SQLSTATE 40P01)` under high-concurrency ticket reservations.
- **Root Cause**: Two transactions reserving adjacent seats acquired locks in opposing orders: Transaction A locked seat 10 then attempted seat 11; Transaction B locked seat 11 then attempted seat 10.
- **Remedy**: Implemented strict deterministic sorting by primary key (`sort.Slice(seats, ...)`) in the application layer before executing `SELECT ... FOR UPDATE`. With identical lock acquisition order, deadlocks became mathematically impossible.

## 2. Ephemeral Port Exhaustion under Load
- **Symptom**: Web crawler worker nodes throwing `dial tcp: bind: address already in use` at ~5,000 requests/second.
- **Root Cause**: Go HTTP client response bodies were being closed without reading unconsumed payload bytes into `io.Discard`. The underlying TCP socket was placed into `TIME_WAIT` status by the Linux kernel instead of returning immediately to the `http.Transport` connection pool.
- **Remedy**: Standardized all HTTP responses with `_, _ = io.Copy(io.Discard, resp.Body)` followed by `_ = resp.Body.Close()`. Replaced default transport with custom `http.Transport` configuring `MaxIdleConns: 1000` and `MaxIdleConnsPerHost: 100`.

## 3. Webhook Double-Delivery & Out-of-Order Execution
- **Symptom**: Payment gateway retried payment notification webhooks due to momentary network delays, causing duplicate balance credits.
- **Root Cause**: Handler lacked atomic idempotency checking and processed the payload twice.
- **Remedy**: Added unique constraint on `(provider, event_id)` in PostgreSQL, coupled with atomic state machine transitions verifying that the booking state was strictly `PENDING` before transitioning to `PAID`.

