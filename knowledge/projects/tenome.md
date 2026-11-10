# Project: Tenome Micro (tenome-micro)

- **ID**: `tenome`
- **Tagline**: Event-Driven Distributed Web Crawler, Search Engine & Local-LLM Summarization Microservices in Go & Kubernetes
- **Status**: In Progress (4 Go microservices, Docker Compose + observability stack, Kubernetes/Helm manifests, and CI/CD pipelines implemented; unit/integration test suites, full JWT validation, `CrawlBatch` contract parity, and Grafana dashboard JSONs remain incomplete)
- **GitHub**: https://github.com/aswinsreeraj/tenome-micro
- **Core Stack**: Go 1.26, Gin v1.12, gRPC/Protobuf, Apache Kafka (KRaft / `kafka-go`), Dragonfly DB (`go-redis/v9`), Pure-Go SQLite (`modernc.org/sqlite` + GORM), Ollama (`qwen2.5:0.5b`), OpenTelemetry (`otelgin`, `otelgrpc`, `otelhttp`, `otlptracehttp`), Prometheus (`ServiceMonitor`), Grafana Tempo, Loki + Grafana Alloy, Kubernetes + Helm, NGINX Gateway Fabric, Docker.

## Overview & Problem Solved
Tenome Micro decomposes crawling, full-text inverted-index search, AI page summarization, and downstream notification dispatch across four independent Go modules (`tenome-gateway`, `tenome-search`, `tenome-ai`, `tenome-notify`). It keeps synchronous REST and gRPC query/submission endpoints responsive while offloading slow HTML scraping, full-page hydration, and local LLM inference onto an asynchronous Kafka event pipeline with W3C OpenTelemetry trace propagation and Dead Letter Topic (`summary-generated-dlt`) fault isolation.

## 4-Service Architecture & Execution Flow
1. **`tenome-gateway` (Gin HTTP `:8080` -> gRPC Client)**:
   - Runs `otelgin` -> `RequestID` (`X-Request-ID` HTTP header + `x-request-id` gRPC metadata) -> structured Zap logging -> CORS -> panic Recovery -> `120s` Timeout -> placeholder `Authenticate` middleware.
   - Maps gRPC status codes to HTTP responses (`codes.ResourceExhausted` -> `HTTP 429`), aggregates backend gRPC health checks (`grpc_health_v1`) at `GET /health`, and exports `/metrics`.
   - Packaged with a dedicated Helm chart (`k8s/charts/tenome-gateway`).
2. **`tenome-search` (gRPC `:8051` + HTTP `:8050`)**:
   - Enqueues crawl URLs into a fixed `100`-slot buffered channel (`jobs := make(chan worker.CrawlJob, 100)`) via non-blocking `select`, returning `ErrQueueFull` cleanly when saturated.
   - Runs `4` worker goroutines (`15s` timeout per job) that scrape `<title>` and visible HTML text (`golang.org/x/net/html`), upsert into pure-Go SQLite `pages` via `INSERT ... ON CONFLICT(url) DO UPDATE ... RETURNING id`, update an in-memory `InvertedIndex` (`GetAllPages` rehydrated at boot), and asynchronously publish lightweight claim-check metadata (`PageIndexedEvent`) to Kafka topic `page-processed` (`5s` publish timeout + W3C trace headers).
   - Serves queries via Dragonfly cache (`search:<term>`, `5m` TTL), falling back to `InvertedIndex.Search` + SQLite `GetPagesByIDs`.
3. **`tenome-ai` (Kafka Consumer + gRPC `:50051` + Metrics `:8081`)**:
   - Consumes `page-processed` (`ai-summary-group`), extracts OTel trace headers, and handles producer casing/type skew cleanly via `FlexString` (`int64`/`string`/`bool`).
   - Fetches page body back from `tenome-search` over gRPC (`SearchService.GetPage`) with up to `3` exponential retries (`500ms` base backoff), invokes self-hosted Ollama (`/api/generate`, `qwen2.5:0.5b`, `60s` timeout, `3` retries `1s` base backoff), persists results into GORM SQLite `summaries` (`uniqueIndex idx_summaries_page_id`), and emits `summary-generated`.
4. **`tenome-notify` (Kafka Consumer + Health/Metrics `:8080`)**:
   - Consumes `summary-generated` (`notification-service-group`), validates schemas upfront (committing malformed poison pills immediately so partitions don't stall), retries failures with `±20%` jittered exponential backoff (`CalculateBackoff`), and forwards exhausted messages to Dead Letter Topic `summary-generated-dlt` with `dlt_reason` and `dlt_timestamp` headers before committing offsets. Currently wired with `ConsoleSender` stdout/Zap sink.

## Key Engineering Trade-Offs & Honest Gaps
- **Claim-Check Choreography vs. Fat Kafka Payloads / Sync Fan-Out**: Emitting `pageId`/`url` metadata instead of scraped HTML keeps Kafka topics lightweight and isolates crawl workers from 60s CPU Ollama inference latency.
- **Pure-Go Embedded SQLite + Dragonfly vs. Postgres + Elasticsearch**: Enables static `CGO_ENABLED=0` Alpine containers (non-root UID `10001`) within a 6 GiB Minikube memory envelope. Known trade-off noted in architecture: `ReadWriteOnce` SQLite PVCs and process-local `InvertedIndex` state prevent safe multi-replica HPA scaling on `tenome-search`, and `InvertedIndex.Search` currently reads `i.index[term]` without `i.mu.Lock()`.
- **Local Ollama Sizing (`14b -> 0.5b`) & Timeouts**: Commit `cd7b81b` downsized Ollama from `qwen2.5-coder:14b` (`8192` MB Minikube) to `qwen2.5:0.5b` (`6144` MB), raised gateway timeouts from `15s` to `120s`, and fixed single-node KRaft `KAFKA_CONTROLLER_QUORUM_VOTERS` (`1@localhost:9093`).
- **Known Gaps Explicitly Tracked**: `0` `*_test.go` automated tests across the 4 modules, stubbed `Authenticate` JWT middleware (parses Bearer header without cryptographic verification), `CrawlBatch` proto/query-param (`?query=` vs README `?q=`) drift between Gateway and Search, `RegenerateSummary` calling `Save()` with a fresh UUID risking `idx_summaries_page_id` conflict on existing `page_id`, mock echoes in `deploy.yml` alongside `production-deploy.yml`, and empty `grafana/dashboards/`.
