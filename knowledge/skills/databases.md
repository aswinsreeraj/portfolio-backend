# Technical Skills: Databases & Persistence

## PostgreSQL & Relational Systems
- **ACID Transactions & Isolation Levels**: Deep experience with `Read Committed`, `Repeatable Read`, and `Serializable` isolation levels. Handling serialization anomalies and phantom reads.
- **Locking Mechanisms**: Pessimistic row-level locking (`SELECT ... FOR UPDATE`, `SKIP LOCKED`, `NOWAIT`), optimistic concurrency control via integer version counters, and deadlock prevention through deterministic lock sorting.
- **Indexing & Performance**: B-Tree, GIN, GiST, BRIN, and composite indexes. Query analysis with `EXPLAIN (ANALYZE, BUFFERS)`, connection pool sizing (`MaxOpenConns`, `MaxIdleConns`, connection lifetime).
- **pgvector**: High-dimensional vector storage, HNSW and IVFFlat vector index configurations, cosine distance (`<=>`), inner product, and Euclidean distance operations for RAG pipelines.
- **ORM & Drivers**: GORM, raw SQL with `database/sql`, `pgx/v5` driver, schema migrations with versioned SQL scripts.

## In-Memory & NoSQL Stores
- **Redis & Dragonfly**: Caching strategies (Cache-Aside, Write-Through), distributed locks (Redlock algorithm, TTL leases), sliding-window rate limiting, Pub/Sub, Redis Hashes, Sets, Sorted Sets, and Bloom filters for fast membership testing.
- **MongoDB**: Document data modeling, aggregation pipelines, replica sets, and indexing.

