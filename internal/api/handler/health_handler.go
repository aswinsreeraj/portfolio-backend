package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db        *sql.DB
	startTime time.Time
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		db:        db,
		startTime: time.Now(),
	}
}

func (h *HealthHandler) HandleHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	dbStatus := "healthy"
	if err := h.db.PingContext(ctx); err != nil {
		dbStatus = "degraded: " + err.Error()
	}

	uptime := time.Since(h.startTime).Round(time.Second).String()

	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"database": dbStatus,
		"uptime":   uptime,
		"version":  "1.0.0",
		"time":     time.Now().UTC().Format(time.RFC3339),
	})
}

