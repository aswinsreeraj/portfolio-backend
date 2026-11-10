# Service Deep Dive: tenome-ai (Part of Tenome Micro)

- **ID**: `tenome_ai`
- **Repository**: https://github.com/aswinsreeraj/tenome-micro (directory `tenome-ai/`)
- **Role**: Asynchronous AI Summarization Consumer & Internal gRPC Service (`:50051` gRPC, `:8081` Prometheus metrics)
- **Core Stack**: Go 1.26, Apache Kafka (`kafka-go`), gRPC/Protobuf (`otelgrpc`), Ollama (`qwen2.5:0.5b` / configurable via `OLLAMA_MODEL`), Pure-Go SQLite + GORM (`modernc.org/sqlite`), OpenTelemetry (`otlptracehttp`, `otelhttp`).

## Overview
`tenome-ai` decouples local LLM page summarization cleanly from web crawling and search indexing:
1. **Kafka Claim-Check Consumption (`ai-summary-group`)**: Listens on topic `page-processed` for lightweight `PageIndexedEvent` payloads (`pageId`, `url`, `indexedAt`) instead of streaming raw HTML documents across Kafka.
2. **Schema-Skew Tolerance (`FlexString`)**: Normalizes `int64`, `string`, and `bool` fields across camelCase and snake_case JSON payloads without crashing consumer loops.
3. **gRPC Hydration & Backoff Retries**: Extracts W3C OpenTelemetry trace context from Kafka headers, fetches full parsed page content from `tenome-search` via `SearchService.GetPage` (3 attempts, `500ms` initial exponential backoff skipping non-retryable gRPC status codes), and calls Ollama (`POST /api/generate`, `stream: false`, `60s` HTTP timeout, `3` retries `1s` base backoff).
4. **Persistence & Downstream Fan-Out**: Stores summaries in GORM SQLite (`summaries` table with unique index `idx_summaries_page_id`) and emits `SummaryGeneratedEvent` to Kafka topic `summary-generated` (consumed by `tenome-notify`), propagating OpenTelemetry traceparent headers end to end into Grafana Tempo.
