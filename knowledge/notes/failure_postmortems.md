# Technical Notes: Failure Postmortems & Production Lessons

Real-world production failure scenarios analyzed and solved by Aswin:

## 1. Silent Zero-Row Truncation on Filtered `pgvector` HNSW Scans (ARDA)
- **Symptom**: On a shared 12,650-row multi-tenant table, vector similarity queries scoped to a 1,000-row agent partition returned **0 rows (0.0% recall)**.
- **Root Cause**: Under default `hnsw.iterative_scan = off` in pgvector 0.8.x, filtered HNSW scans evaluate candidates globally and discard non-matching rows until `ef_search` exhausts without finding enough tenant matches.
- **Remedy**: Set `SET LOCAL hnsw.iterative_scan = strict_order` in `PgVectorReferenceExampleRepository.retrieve_approximate`, restoring `99.1%` Recall@1 and `98.9%` Recall@5 while preserving exact cosine order.

## 2. Single-Node Kafka KRaft, Ollama OOM & Duplicate URL Constraint Failures (Tenome Micro)
- **Symptom**: During Minikube cluster bring-up, Kafka failed quorum election, Ollama `14b` models hit node memory limits while synchronous Gateway HTTP requests timed out at `15s`, and re-crawled URLs errored on `UNIQUE(url)` inserts.
- **Root Cause**: `KAFKA_CONTROLLER_QUORUM_VOTERS` pointed to `1@kafka:9093` instead of `1@localhost:9093`, `qwen2.5-coder:14b` exceeded `8192` MB Minikube node RAM while gateway HTTP timeouts closed before CPU LLM responses finished, and `SQLiteStorage.SavePage` executed plain `INSERT`.
- **Remedy**: Bound KRaft voters to `1@localhost:9093` (`cd7b81b`), downsized to `qwen2.5:0.5b` under a 6 GiB cluster profile, extended Gateway HTTP/request timeouts from `15s` to `120s`, and refactored `SavePage` (`3753409`) to an idempotent `INSERT ... ON CONFLICT(url) DO UPDATE ... RETURNING id`.

## 3. Expired-Hold Razorpay Captures & Split-Fee Reserve Deficits (EVNTX)
- **Symptom**: Payments completing after EVNTX's 1-minute booking reservation hold expired by `BookingExpirationJob` required manual admin refunds, while partial buyer cancellations left platform-fee accounting gaps.
- **Root Cause**: Earlier builds queued late captures in a manual refund workflow and refunded full payments without separating non-refundable platform fees from organizer balances.
- **Remedy**: Refactored `PaymentUsecase.VerifyPayment` (`38991b6`) to detect `ErrBookingExpiredPaymentSuccess` and invoke `RazorpayService.RefundPayment` automatically, and tracked non-refundable fee obligations in organizer `reserve_balance` (`available_balance - abs(reserve_balance)`).

