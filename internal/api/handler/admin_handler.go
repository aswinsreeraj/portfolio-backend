package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/aswin/portfolio-rag/internal/service/ingestion"
)

type AdminHandler struct {
	worker *ingestion.IngestionWorker
}

func NewAdminHandler(worker *ingestion.IngestionWorker) *AdminHandler {
	return &AdminHandler{worker: worker}
}

func (h *AdminHandler) HandleReindex(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := h.worker.SyncKnowledge(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Reindexing failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Knowledge corpus reindexed successfully",
	})
}
