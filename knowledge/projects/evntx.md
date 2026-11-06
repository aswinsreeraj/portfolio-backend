# Project: EVNTX (Event Management & Ticketing Platform)

- **ID**: `evntx`
- **Tagline**: Concurrency-Safe Event Management & Financial Settlement Platform
- **Status**: Completed
- **Key Metrics**: P99 Latency 12ms | Sustained Load 3,500 rps | Failure Tests: 0 double-bookings under 5,000 concurrent buys
- **GitHub**: https://github.com/aswin/evntx
- **Demo**: https://evntx.dev

## Overview & Motivation
EVNTX is a production-grade ticket reservation and event management platform engineered to eliminate double-booking during flash sales and guarantee absolute financial reconciliation across ticket buyers, digital wallets, and organizer payouts.

## Key Features & Architecture
1. **Concurrency-Safe Seat Reservation**: Implemented optimistic concurrency control using version checks and conditional inventory updates:
   ```sql
   UPDATE ticket_tiers SET count = count - 1, version = version + 1
   WHERE id = $1 AND version = $2 AND count > 0;
   ```
   Users receive temporary reservation holds with an explicit TTL, preventing overselling without holding long table locks.
2. **Transactional Multi-Party Ledger**: Built a multi-party double-entry wallet transaction system using PostgreSQL transactions and row-level locking (`SELECT FOR UPDATE`). Coordinates ticket purchases, platform commissions, organizer credits, and wallet withdrawals. Ensures total debits equal total credits.
3. **Razorpay Payments & Automated Late Payment Recovery**: Integrates Razorpay order creation and HMAC-SHA256 webhook signature verification. If a payment webhook arrives after a user's reservation hold has expired, the state machine detects the expired reservation and triggers an automated refund API call.
4. **Clean Architecture in Go**: Handlers, Use Cases, Repositories, and Entities are strictly decoupled with Go interfaces, making business logic 100% unit-testable without database dependencies.
5. **Security & Deployment**: JWT session management, bcrypt password hashing, email OTP authentication, and Google OAuth. Containerized with Docker and deployed on AWS EC2.

## Engineering Decisions & Trade-Offs
- **Optimistic Concurrency Control over Global Pessimistic Locks**: Global mutexes in Redis or DB table locks choke throughput when thousands of buyers purchase tickets across hundreds of separate tiers. Optimistic version checks allow concurrent purchases to proceed in parallel without contention.
- **Double-Entry Ledger over Mutable Balance Columns**: Mutable `user.balance += amount` columns make fraud tracking and partial refund reconciliation impossible. EVNTX writes immutable ledger entry pairs for every monetary movement.
- **Automated Late Refund Engine**: Eliminates manual support tickets by having the webhook handler check the reservation state before creating an order; if expired, it immediately executes a refund.

## What Went Wrong & Lessons Learned
- **Deadlock on Interleaved Multi-Seat Reservation Transactions**: When two users attempted to reserve overlapping seat IDs in opposite alphabetical order within concurrent PostgreSQL transactions, row-level locks produced intermittent transaction deadlocks (`SQLSTATE 40P01`). The solution was enforcing strict deterministic sorting by seat ID in the Go domain layer prior to acquiring database row locks.
- **Key Takeaway**: Always acquire database locks in globally consistent, sorted order across all concurrent transaction paths.

