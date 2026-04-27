package rest

import (
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/whoAngeel/rago/internal/core/ports"
)

func requestLogger(logger ports.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		logger.Info(
			"incoming request",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"trace_id", requestid.Get(c),
			// "ip", c.ClientIP(),
			"latency", time.Since(start),
		)
	}
}
