# Technical Notes: Failure Postmortems & Production Lessons

Real-world production failure scenarios analyzed and solved by Aswin:

## 1. Silent Zero-Row Truncation on Filtered `pgvector` HNSW Scans (ARDA)
- **Symptom**: On a shared 12,650-row multi-tenant table, vector similarity queries scoped to a 1,000-row agent partition returned **0 rows (0.0% recall)**.
- **Root Cause**: Under default `hnsw.iterative_scan = off` in pgvector 0.8.x, filtered HNSW scans evaluate candidates globally and discard non-matching rows until `ef_search` exhausts without finding enough tenant matches.
- **Remedy**: Set `SET LOCAL hnsw.iterative_scan = strict_order` in `PgVectorReferenceExampleRepository.retrieve_approximate`, restoring `99.1%` Recall@1 and `98.9%` Recall@5 while preserving exact cosine order.

## 2. Ephemeral Port Exhaustion under Load (Tenome)
- **Symptom**: Web crawler worker nodes throwing `dial tcp: bind: address already in use` under high crawler concurrency.
- **Root Cause**: Go HTTP client response bodies were being closed without reading unconsumed payload bytes into `io.Discard`. The underlying TCP socket was placed into `TIME_WAIT` status by the Linux kernel instead of returning immediately to the `http.Transport` connection pool.
- **Remedy**: Standardized all HTTP responses with `_, _ = io.Copy(io.Discard, resp.Body)` followed by `_ = resp.Body.Close()`. Replaced default transport with custom `http.Transport` configuring `MaxIdleConns` and `MaxIdleConnsPerHost`.

## 3. Expired-Hold Razorpay Captures & Split-Fee Reserve Deficits (EVNTX)
- **Symptom**: Payments completing after EVNTX's 1-minute booking reservation hold expired by `BookingExpirationJob` required manual admin refunds, while partial buyer cancellations left platform-fee accounting gaps.
- **Root Cause**: Earlier builds queued late captures in a manual refund workflow and refunded full payments without separating non-refundable platform fees from organizer balances.
- **Remedy**: Refactored `PaymentUsecase.VerifyPayment` (`38991b6`) to detect `ErrBookingExpiredPaymentSuccess` and invoke `RazorpayService.RefundPayment` automatically, and tracked non-refundable fee obligations in organizer `reserve_balance` (`available_balance - abs(reserve_balance)`).

