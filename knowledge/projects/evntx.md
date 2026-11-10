# Project: EVNTX (Multi-Role Event Discovery, Ticketing & Escrow Wallet Platform)

- **ID**: `evntx`
- **Tagline**: Multi-Role Event Discovery, Concurrency-Safe Ticketing & Escrow Wallet Platform
- **Status**: Completed (289 commits, 65+ PRs across Clean Architecture Go backend and React 19 TypeScript SPA)
- **GitHub**: https://github.com/aswinsreeraj/evntx
- **Live Demo**: https://evntx.aswinsreeraj.online (Vercel frontend + live CORS origins `evntx.aswinsreeraj.online` / `evntx-self.vercel.app`)
- **Core Stack**: Go 1.25.5, Gin v1.12.0, GORM v1.31.1 + `pgx/v5`, PostgreSQL, Razorpay Go SDK v1.4.0, AWS SDK v2 S3, JWT (`HS256`) + Google OAuth, `robfig/cron/v3`, `zerolog`, React 19.2, TypeScript 5.9, TanStack Query v5, Zustand v5, Tailwind CSS v4, Recharts, `html5-qrcode`, Docker.

## Overview & Problem Solved
EVNTX serves attendees (`goer`), event hosts (`organizer`), and platform operators (`admin`) through a Clean Architecture Go monolith (`internal/domain` -> `internal/repository` -> `internal/usecase` -> `internal/infrastructure` -> `internal/delivery/http`) backed by 29 PostgreSQL tables (`AutoMigrate`) and a feature-modular React 19 SPA (`frontend/src/modules/`). It solves high-concurrency ticket tier overselling, short-lived 1-minute checkout hold expirations, escrow-style organizer fund settlement, and split-fee refund accounting across external Razorpay payments and an internal multi-bucket wallet ledger.

## Key Architecture & Implementation Details
1. **Hybrid Concurrency Control (`ReserveTickets` vs. `PayWithWallet` & `CheckInTicket`)**:
   - **Inventory Reservations (`booking_gorm.go`)**: Optimistic compare-and-swap on `ticket_type_models.version` and `available_quantity >= qty` up to 3 retry attempts, creating a **1-minute reservation hold** (`expires_at`) and generating `TKT-<uuid[:8]>` codes in `ticket_models`. Avoids blocking concurrent shoppers browsing popular tiers.
   - **Wallet Ledger & QR Check-In**: `PayWithWallet`, `ApplyPlatformTransaction`, and `CheckInTicket` acquire explicit PostgreSQL row-level locks (`clause.Locking{Strength: "UPDATE"}`) inside ACID transactions so user/organizer/platform wallet balances stay consistent and single-use QR tickets transition `valid -> used` at most once.
2. **Escrow Settlement & Split-Fee Reserve Accounting**:
   - Wallets track three distinct buckets: `available_balance`, `pending_balance`, and `reserve_balance`, alongside a singleton platform treasury (`platform_wallet_models`, UUID `00000000-0000-0000-0000-000000000001`).
   - Booking revenues credit organizer `pending_balance` while platform fees credit `platform_wallet_models`. Only after event completion does `SettleEventEarnings` deduct the 5% organizer fee and release funds to `available_balance`.
   - Buyer cancellations obey `RefundWindowDays`, calculate ticket-only refunds (`payment.Amount - booking.TotalAmount`), and track non-refundable platform fee obligations in `reserve_balance`; `RequestPayout` blocks withdrawals exceeding `available_balance - abs(reserve_balance)`.
3. **Razorpay Verification & Late-Capture Auto-Refunds**:
   - Supports free-ticket bypass (`FREE_<bookingID>`), Razorpay order creation (`POST /payments/razorpay/order`), cryptographic signature verification (`POST /payments/razorpay/verify`), and `FetchOrder` validation on wallet top-ups.
   - If a Razorpay payment verifies *after* the 1-minute reservation hold was already released by `BookingExpirationJob`, `PaymentUsecase.VerifyPayment` catches `ErrBookingExpiredPaymentSuccess` (`EVT_015`) and triggers `RazorpayService.RefundPayment` automatically back to the buyer's source method.
4. **Embedded Cron Scheduler, In-Process Caching & Storage**:
   - **No Redis or Kafka**: Uses a single Go monolith with an in-process `sync.RWMutex` TTL cache (`60s` TTL on `GET /events/:slug`, invalidated on edits), per-IP token-bucket rate limiting (`golang.org/x/time/rate` on OTP & signup routes), and `robfig/cron/v3` background workers wrapped in 2-minute context timeouts and 3-attempt retries logged to `job_log_models`:
     - `BookingExpirationJob` (`*/10 * * * *`): Expires stale `reserved` bookings (batch limit `100`) and restores `available_quantity`.
     - `AutoProcessCompletedEventsJob` (`*/45 * * * *`): Transitions ended live events to `completed` and runs `SettleEventEarnings`.
     - `ProcessPayoutSettlementsJob` (`*/30 * * * *`): Transitions `approved` `payout_requests` to `completed`.
   - Pluggable `storage.Uploader` interface supporting AWS S3 (SDK v2) and local filesystem fallback.
5. **Authentication, Cryptography, Funnel Telemetry & SPA**:
   - Passwordless 6-digit CSPRNG bcrypt OTP (5m TTL), Google ID Token OAuth, 15m HS256 JWT access tokens with transparent Axios 401 refresh via SHA-256 hashed 7d refresh sessions (`user_session_models`), and **AES-256-GCM encryption** (`pkg/encryption/encryption.go`) for organizer bank account numbers, IFSC codes, and UPI IDs.
   - Captures anonymous `visitor_sessions` and `engagement_events` (excluding admin/organizer actors) and upserts daily conversion funnels into `event_engagement_daily` via `ON CONFLICT (event_id, date) DO UPDATE`, paired with `zerolog`, `pprof`, Recharts admin dashboards, RFC-4180 CSV export, and browser camera QR scanning (`html5-qrcode`).

## Testing, Benchmarks & Real Bugs Solved
- **Test Suites & Statement Coverage**:
  - `go test -race -v ./internal/usecase/...` passes all **6 top-level tests (11 subtests)** in `5.757s` using `testify` mocks (`internal/repository/mocks`), with `14.0%` total statement coverage across `internal/usecase` (`coverage.out`).
  - Live-DB integration tests (`tests/integration/{auth,booking,wallet}_int_test.go`) and full Gin router E2E tests (`tests/e2e/workflow_e2e_test.go`) run against PostgreSQL (`AutoMigrate` + `TRUNCATE ... CASCADE`). `benchmark_test.go` defines 4 HTTP benchmarks without checked-in run results.
- **Key Engineering Fixes From Commit History**:
  - **`38991b6` — Automated Razorpay Refunds on Late Capture**: Eliminated the manual admin refund queue by automatically invoking `RazorpayService.RefundPayment` when payment verification arrives after the 1-minute hold expired.
  - **`28feb23` & `f5eec29` — Split-Fee Cancellation Refunds & `reserve_balance`**: Prevented organizers from over-withdrawing when buyer cancellations left non-refundable platform fee deficits.
  - **`20d6157` — Preserving Unaffected Ticket Tiers During Event Edits**: Fixed `UpdateEvent` so editing event metadata preserves existing ticket inventory via explicit `WHERE event_id = ? AND id NOT IN ?` cleanup inside a transaction.
  - **`71d54e8` & `925d915` — Cron Interval Tuning, DB Pooling & Indexes**: Adjusted cron frequencies (`*/10`, `*/30`, `*/45`), configured `PrepareStmt: true`, `MaxOpenConns(25)` / `MaxIdleConns(10)`, and added GORM B-Tree indexes after `pprof` profiling (`1c2b564`) and 60s caching (`363b0eb`).
  - **`9915484` & `bce6507` — Ambiguous Join Columns & RFC-4180 CSV**: Fixed ambiguous `status` columns in joined `GetUserBookings` queries and hardened CSV export quote escaping.
