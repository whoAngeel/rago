package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/infrastructure/sse"
)

type SSEHandler struct {
	manager *sse.Manager
	logger  ports.Logger
}

func NewSSEHandler(manager *sse.Manager, logger ports.Logger) *SSEHandler {
	return &SSEHandler{
		manager: manager,
		logger:  logger,
	}
}

func (h *SSEHandler) Stream(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt("user_id")

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	client := &ports.SSEClient{
		ID:      uuid.NewString(),
		UserID:  userID,
		Channel: make(chan ports.SSEEvent, 10),
	}

	h.manager.AddClient(userID, client)
	defer h.manager.RemoveClient(userID, client.ID)

	c.Writer.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-heartbeat.C:
			c.Writer.Write([]byte(": ping\n\n"))
			c.Writer.Flush()

		case event := <-client.Channel:
			c.SSEvent(event.Type, event.Data)
			c.Writer.Flush()
		}
	}
}
