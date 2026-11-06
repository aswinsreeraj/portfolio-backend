# Technical Skills: Cloud, Infrastructure & Observability

## Infrastructure & Containers
- **Docker & Compose**: Multi-stage Docker builds for minimal Go production images (`scratch` / `alpine`), volume mounts, networking bridges, compose files for local orchestration.
- **Kubernetes**: Deployments, StatefulSets, Services (ClusterIP, NodePort, LoadBalancer), ConfigMaps, Secrets, Ingress, liveness and readiness health probes.
- **AWS**: Amazon EC2 instance provisioning, S3 bucket lifecycle management and presigned URLs, RDS PostgreSQL, VPC networking, security groups, and IAM least-privilege policies.

## CI/CD & Automation
- **GitHub Actions**: Automated linting (`golangci-lint`), race detector testing, container build pipelines, multi-arch image publishing (`amd64`/`arm64`).
- **Makefiles & Taskfiles**: Reproducible local developer environments, migration scripts, and test runners.

## Observability & Reliability
- **Structured Logging**: Go 1.21+ `log/slog` structured JSON logs with correlation IDs (`request_id`, `trace_id`, `tenant_id`).
- **Metrics**: Prometheus instrumentation (`counter`, `gauge`, `histogram` for latency percentiles p50/p95/p99), Grafana dashboards.
- **Tracing**: OpenTelemetry (OTel) context propagation across microservices.

