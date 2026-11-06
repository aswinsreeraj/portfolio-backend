package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type clientLimit struct {
	count       int
	windowStart time.Time
}

type IPRateLimiter struct {
	mu      sync.Mutex
	clients map[string]*clientLimit
	limit   int
	window  time.Duration
}

func NewIPRateLimiter(limit int, window time.Duration) *IPRateLimiter {
	limiter := &IPRateLimiter{
		clients: make(map[string]*clientLimit),
		limit:   limit,
		window:  window,
	}

	// Periodic cleanup of stale clients
	go func() {
		for {
			time.Sleep(window * 2)
			limiter.mu.Lock()
			now := time.Now()
			for ip, entry := range limiter.clients {
				if now.Sub(entry.windowStart) > window*2 {
					delete(limiter.clients, ip)
				}
			}
			limiter.mu.Unlock()
		}
	}()

	return limiter
}

func (l *IPRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		l.mu.Lock()
		entry, exists := l.clients[ip]
		now := time.Now()

		if !exists || now.Sub(entry.windowStart) >= l.window {
			l.clients[ip] = &clientLimit{
				count:       1,
				windowStart: now,
			}
			l.mu.Unlock()
			c.Next()
			return
		}

		if entry.count >= l.limit {
			l.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please wait a moment before sending more messages.",
			})
			return
		}

		entry.count++
		l.mu.Unlock()
		c.Next()
	}
}

