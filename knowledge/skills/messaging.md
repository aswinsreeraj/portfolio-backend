# Technical Skills: Messaging & Event-Driven Systems

## Apache Kafka & Redpanda
- **Architecture**: Broker clustering, topics, partitions, consumer groups, offsets, replication factor, ISR (In-Sync Replicas).
- **Partitioning Strategies**: Semantic partitioning using domain/entity keys to guarantee strict message ordering within a single partition while scaling horizontally across partitions.
- **Reliability & Delivery Guarantees**: At-least-once vs at-most-once vs exactly-once semantics (EOS), producer acks (`acks=all`), idempotent producers, and dead-letter queues (DLQs).
- **Transactional Outbox Pattern**: Solving dual-write inconsistencies between relational databases and message brokers by persisting domain events to an `outbox` table within the primary business transaction.

## RabbitMQ
- Exchanges (Direct, Topic, Fanout, Headers), queue bindings, consumer prefetch count, message acknowledgments (`ack`/`nack`/`reject`), and backpressure management.

