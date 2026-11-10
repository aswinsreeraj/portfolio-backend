# Technical Skills: Cloud, Infrastructure & Observability

## Infrastructure & Containers
- **Docker & Compose**: Multi-stage Docker builds (`CGO_ENABLED=0` static non-root `alpine` containers in Tenome Micro), volume mounts, networking bridges, and Docker Compose orchestration across Postgres/pgvector, Temporal, Redpanda, Vault, Kafka KRaft, Dragonfly, and Ollama.
- **Kubernetes & Packaging**: Deployments, Services (ClusterIP, NodePort, LoadBalancer), ConfigMaps, Secrets, PersistentVolumeClaims (`20Gi` Ollama PVC, SQLite PVCs), HPAs, NGINX Gateway Fabric (`Gateway`/`HTTPRoute`), Helm charts (`tenome-gateway`), liveness/readiness probes, and `ServiceMonitor` CRDs.
- **AWS**: Amazon S3 (`aws-sdk-go-v2` media storage in EVNTX with local fallback), EC2 hosting, RDS PostgreSQL, security groups, and IAM least-privilege policies.

## CI/CD & Automation
- **GitHub Actions**: Automated Go test runs (`test.yml`, `tests.yml`), multi-service Docker build pipelines, and Kubernetes/Helm deploy workflows.
- **Tooling**: Protobuf/gRPC code generation scripts, Alembic & `golang-migrate` schema runners.

## Observability & Reliability
- **Structured Logging**: Structured JSON logging (`zerolog`, `zap`, Go `log/slog`) with `RequestID` and `TraceID` log-to-trace correlation via Grafana Alloy -> Loki.
- **Metrics & Health**: Prometheus instrumentation (`/metrics`, `ServiceMonitor` 15s scrapes) alongside gRPC `grpc_health_v1` and HTTP readiness probes.
- **Tracing**: Distributed OpenTelemetry (`OTLP` HTTP `:4318`) propagating W3C `TraceContext` across Gin, gRPC, Kafka message headers, and Ollama HTTP clients into Grafana Tempo (`service-graphs` & `span-metrics`).

