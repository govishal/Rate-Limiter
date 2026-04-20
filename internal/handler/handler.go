package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"rate-limiter/internal/models"
	"rate-limiter/internal/ratelimiter"
)

// Handler serves HTTP routes backed by the in-memory sliding-window limiter.
type Handler struct {
	limiter *ratelimiter.SlidingWindow
}

// NewHandler returns an HTTP handler wired to the rate limiter.
func NewHandler(limiter *ratelimiter.SlidingWindow) *Handler {
	return &Handler{limiter: limiter}
}

// RegisterRoutes attaches routes to the given Gin engine.
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/request", h.handleRequest)
	r.GET("/stats", h.handleStats)
}

func (h *Handler) handleRequest(c *gin.Context) {
	var body models.PostRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	if body.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	allowed, count := h.limiter.TryConsume(body.UserID)
	if !allowed {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": fmt.Sprintf(
				"rate limit exceeded: max %d requests per user per %s",
				h.limiter.MaxRequests(),
				h.limiter.Window().String(),
			),
			"user_id":      body.UserID,
			"retry_after":  fmt.Sprintf("%.0fs", h.limiter.Window().Seconds()),
			"window_count": count,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "request accepted",
		"user_id":      body.UserID,
		"window_count": count,
		"payload":      body.Payload,
	})
}

func (h *Handler) handleStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"window_seconds": int(h.limiter.Window().Seconds()),
		"max_requests":   h.limiter.MaxRequests(),
		"users":          h.limiter.SnapshotStats(),
	})
}
