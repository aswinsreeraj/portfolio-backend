# Project: Tenome AI (Page Summarization & Semantic Search Service)

- **ID**: `tenome_ai`
- **Tagline**: Semantic Retrieval & Page Summarization Pipeline
- **Status**: Learning / Research Project
- **Key Technologies**: Go, PostgreSQL, pgvector, Ollama / Gemini API, Redis, Kafka
- **GitHub**: https://github.com/aswin/tenome

## Overview & Architecture
Tenome AI is an intelligent extension to the Tenome search engine that transforms raw crawled HTML documents into concise technical summaries and high-dimensional vector embeddings for semantic search.

## Key Capabilities
1. **Asynchronous Ingestion**: Crawled HTML documents published to Kafka's `crawled-pages` topic are consumed by the Tenome AI worker pool.
2. **Text Normalization & Chunking**: Strips script tags, navigation chrome, and styling markup. Chunks long-form articles along semantic paragraph boundaries with overlapping context windows.
3. **Local & Cloud LLM Integration**: Uses Ollama for local inference and fallback to external APIs (Gemini) for batch vector embeddings.
4. **Vector Search with pgvector**: Stores 768-dimensional document chunk embeddings in PostgreSQL using an HNSW index, allowing users to perform hybrid search queries combining BM25 keyword matching with cosine semantic similarity.
5. **Caching Layer**: Frequently queried topic vectors and generated summaries are cached in Redis to minimize redundant model calls.

