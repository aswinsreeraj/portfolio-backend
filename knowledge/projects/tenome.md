# Project: Tenome (Distributed Search & Web Crawling Platform)

- **ID**: `tenome`
- **Tagline**: Concurrent Web Crawler & Inverted Index Search Engine
- **Status**: Learning Project
- **Key Metrics**: P99 Latency 14ms | Sustained Load 5,000 rps | Scale: 100,000 documents indexed
- **GitHub**: https://github.com/aswin/tenome
- **Demo**: https://tenome.dev

## Overview & Motivation
Tenome is a distributed web crawling platform and full-text search engine written in Go. It fetches web pages concurrently, respects domain politeness rules, and maps tokens to documents via an in-memory inverted index so queries avoid expensive database table scans.

## Key Architecture & Components
1. **Frontier Manager & URL Deduplication**: Seed URLs are pushed into Redis. A memory-efficient Bloom filter checks whether a URL has already been queued or visited, preventing duplicate network requests.
2. **Kafka Domain-Hash Partitioning**: Messages on the `crawl-urls` topic are partitioned by the domain name hash (`hash(url.Hostname()) % num_partitions`). This guarantees all URLs belonging to the same host route to the exact same consumer worker, making local rate-limiting trivial.
3. **Concurrent Go Fetcher Worker Pool**: Spawns worker goroutines that parse `robots.txt` per host, enforce token-bucket crawl delays, reuse HTTP persistent connections, and parse HTML with streaming tokenizers.
4. **In-Memory Inverted Index**: Maps extracted terms to document postings with term frequencies and document lengths. Uses the **BM25 ranking algorithm** to compute relevance scores in sub-15ms for queries over 100k documents.
5. **Decoupled Architecture**: Crawling, indexing, search APIs, and summarization communicate via Kafka event streams, ensuring crawler spikes or failures never compromise search response latency.

## Engineering Decisions & Trade-Offs
- **Domain-Hash Partitioning over Round-Robin Queuing**: Avoids the need for a distributed Redis lock to prevent flooding a single host. Since all URLs for a given domain arrive at one worker, politeness delays are enforced using local Go `time.Ticker` instances.
- **In-Memory Inverted Index with BM25 over SQL LIKE**: Relational queries (`LIKE '%term%'`) degrade linearly with data size. Inverted indexes provide $O(1)$ token lookup and fast posting list intersection.

## What Went Wrong & Lessons Learned
- **Ephemeral Port Exhaustion from Leaked HTTP Response Bodies**: During stress tests crawling 5,000 pages per second, Linux TCP sockets accumulated in `TIME_WAIT`, causing `dial tcp: bind: address already in use`. The root cause was closing response bodies without first reading unconsumed bytes into `io.Discard`, which prevented the Go HTTP Transport from reusing existing TCP connections.
- **Fix**: Always drain HTTP response bodies with `io.Copy(io.Discard, resp.Body)` before calling `resp.Body.Close()` in high-throughput Go clients.

