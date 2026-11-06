package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/aswin/portfolio-rag/internal/api/handler"
	"github.com/aswin/portfolio-rag/internal/api/middleware"
)

type RouterConfig struct {
	ChatHandler   *handler.ChatHandler
	HealthHandler *handler.HealthHandler
	AdminHandler  *handler.AdminHandler
	RateLimitRPM  int
}

// SetupRouter initializes and configures the Gin engine with all routes and middlewares
func SetupRouter(cfg RouterConfig) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.SlogLogger())
	r.Use(middleware.BodyLimitMiddleware(64 * 1024)) // 64KB max request payload

	// Rate limiter: default 30 req/min per IP
	limit := cfg.RateLimitRPM
	if limit <= 0 {
		limit = 30
	}
	rateLimiter := middleware.NewIPRateLimiter(limit, 1*time.Minute)
	r.Use(rateLimiter.Middleware())

	// Public health probe
	r.GET("/health", cfg.HealthHandler.HandleHealth)

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		v1.POST("/chat", cfg.ChatHandler.HandleChat)
		v1.POST("/admin/reindex", cfg.AdminHandler.HandleReindex)
	}

	return r
}
