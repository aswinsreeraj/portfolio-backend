# Technical Skills: Go & Backend Engineering

## Core Competencies
- **Go (Golang)**: Deep mastery of idiomatic Go (versions 1.21 - 1.23). Proficient in memory layouts, pointer semantics, slice reallocation internals, struct field alignment for memory padding optimization, and escape analysis (`go build -gcflags="-m"`).
- **Concurrency & Primitives**: Goroutines, buffered and unbuffered channels, worker pools, fan-out/fan-in patterns, `sync.Mutex`, `sync.RWMutex`, `sync.WaitGroup`, `sync.Once`, `sync/atomic`, and `errgroup.Group`.
- **Context & Lifecycle Management**: Context propagation (`context.Context`), timeout handling, cancellation trees, deadline enforcement, and context leak prevention.
- **Clean Architecture & Design Patterns**: Domain-Driven Design (DDD), Separation of Concerns (Handler -> Service/Use Case -> Repository -> Entity), dependency injection without heavy reflection, interface segregation.
- **RESTful APIs & Web Frameworks**: Gin framework, standard `net/http` server tuning (`ReadTimeout`, `WriteTimeout`, `IdleTimeout`), middleware authoring (rate-limiting, CORS, structured logging, recovery, request ID propagation).
- **RPC & Protocols**: gRPC, Protocol Buffers (protobuf v3), streaming RPCs, Model Context Protocol (MCP).
- **Testing & Benchmarking**: Table-driven unit tests, mock generators (`mockery` / interface stubbing), race condition detection (`go test -race`), CPU and memory profiling with `pprof`.

