package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/aswin/portfolio-rag/internal/domain"
	"github.com/aswin/portfolio-rag/internal/service/chat"
)

type ChatHandler struct {
	chatService *chat.ChatService
}

func NewChatHandler(chatService *chat.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

func (h *ChatHandler) HandleChat(c *gin.Context) {
	var req domain.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request payload. 'messages' array must not be empty.",
		})
		return
	}

	// 30-second context deadline for chat completion
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	resp, err := h.chatService.ProcessChat(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process chat request",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
