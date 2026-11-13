package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/aswin/portfolio-rag/internal/api"
	"github.com/aswin/portfolio-rag/internal/api/handler"
	"github.com/aswin/portfolio-rag/internal/database"
	"github.com/aswin/portfolio-rag/internal/domain"
	"github.com/aswin/portfolio-rag/internal/provider/embedding"
	"github.com/aswin/portfolio-rag/internal/provider/llm"
	"github.com/aswin/portfolio-rag/internal/repository"
	"github.com/aswin/portfolio-rag/internal/service/chat"
	"github.com/aswin/portfolio-rag/internal/service/ingestion"
	"github.com/aswin/portfolio-rag/internal/service/rag"
)

func main() {
	// Configure structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load .env if present
	_ = godotenv.Load()

	slog.Info("Starting Aswin's Portfolio RAG Assistant Server...")

	// 1. Database Connection
	pgPort, _ := strconv.Atoi(getEnv("POSTGRES_PORT", "5433"))
	dbCfg := database.Config{
		URL:      getEnv("DATABASE_URL", ""),
		Host:     getEnv("POSTGRES_HOST", "127.0.0.1"),
		Port:     pgPort,
		User:     getEnv("POSTGRES_USER", "arda_user"),
		Password: getEnv("POSTGRES_PASSWORD", "arda_password"),
		DBName:   getEnv("POSTGRES_DB", "portfolio_rag"),
		SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
	}

	db, err := database.ConnectPostgres(dbCfg)
	if err != nil {
		slog.Error("Failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// 2. Schema Migration
	if err := database.RunMigrations(db); err != nil {
		slog.Error("Migration error", "error", err)
		os.Exit(1)
	}

	// 3. Provider Initialization
	geminiKey := os.Getenv("GEMINI_API_KEY")
	openAIKey := os.Getenv("OPENAI_API_KEY")
	llmChoice := getEnv("LLM_PROVIDER", "fallback")
	embChoice := getEnv("EMBEDDING_PROVIDER", "fallback")

	var llmProvider domain.LLMProvider
	var embProvider domain.EmbeddingProvider

	// Select Embedding Provider
	if (embChoice == "gemini" || embChoice == "auto") && geminiKey != "" {
		embProvider = embedding.NewGeminiEmbeddingProvider(geminiKey, os.Getenv("GEMINI_EMBED_MODEL"))
		slog.Info("Using Gemini Embedding Provider", "model", getEnv("GEMINI_EMBED_MODEL", "text-embedding-004"))
	} else if (embChoice == "openai") && openAIKey != "" {
		embProvider = embedding.NewOpenAIEmbeddingProvider(openAIKey, os.Getenv("OPENAI_EMBED_MODEL"))
		slog.Info("Using OpenAI Embedding Provider", "model", getEnv("OPENAI_EMBED_MODEL", "text-embedding-3-small"))
	} else {
		if embChoice == "gemini" && geminiKey == "" {
			slog.Warn("EMBEDDING_PROVIDER is set to 'gemini' but GEMINI_API_KEY is empty. Using deterministic offline fallback until key is added in backend/.env")
		}
		embProvider = embedding.NewFallbackEmbeddingProvider()
		slog.Info("Using Deterministic Offline Fallback Embedding Provider")
	}

	// Select LLM Provider
	if (llmChoice == "gemini" || llmChoice == "auto") && geminiKey != "" {
		llmProvider = llm.NewGeminiLLMProvider(geminiKey, os.Getenv("GEMINI_MODEL"))
		slog.Info("Using Gemini LLM Provider", "model", getEnv("GEMINI_MODEL", "gemini-1.5-flash"))
	} else if (llmChoice == "openai") && openAIKey != "" {
		llmProvider = llm.NewOpenAILLMProvider(openAIKey, os.Getenv("OPENAI_MODEL"))
		slog.Info("Using OpenAI LLM Provider", "model", getEnv("OPENAI_MODEL", "gpt-4o-mini"))
	} else {
		if llmChoice == "gemini" && geminiKey == "" {
			slog.Warn("LLM_PROVIDER is set to 'gemini' but GEMINI_API_KEY is empty. Using deterministic offline fallback until key is added in backend/.env")
		}
		llmProvider = llm.NewFallbackLLMProvider()
		slog.Info("Using Deterministic Grounded Fallback LLM Provider")
	}

	// 4. Repositories & Services
	docRepo := repository.NewDocumentRepository(db)
	vectorRepo := repository.NewVectorRepository(db)
	knowledgeDir := getEnv("KNOWLEDGE_DIR", "./knowledge")

	ingestionWorker := ingestion.NewIngestionWorker(knowledgeDir, docRepo, vectorRepo, embProvider)

	// Perform initial knowledge sync
	syncCtx, syncCancel := context.WithTimeout(context.Background(), 60*time.Second)
	if err := ingestionWorker.SyncKnowledge(syncCtx); err != nil {
		slog.Warn("Initial knowledge sync encountered errors", "error", err)
	}
	syncCancel()

	retriever := rag.NewRetrieverService(vectorRepo, embProvider)
	chatService := chat.NewChatService(retriever, llmProvider)

	// 5. Handlers & Router
	chatHandler := handler.NewChatHandler(chatService)
	healthHandler := handler.NewHealthHandler(db)
	adminHandler := handler.NewAdminHandler(ingestionWorker)

	rpm, _ := strconv.Atoi(getEnv("RATE_LIMIT_RPM", "60"))
	router := api.SetupRouter(api.RouterConfig{
		ChatHandler:   chatHandler,
		HealthHandler: healthHandler,
		AdminHandler:  adminHandler,
		RateLimitRPM:  rpm,
	})

	// 6. HTTP Server with Graceful Shutdown
	port := getEnv("PORT", "8080")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 45 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("Server listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server error", "error", err)
		}
	}()

	// Wait for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exited cleanly")
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
