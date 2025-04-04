package middleware

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

type PrintInfo struct {
	ClientIP  string        `json:"client_ip"`
	Status    int           `json:"status"`
	Method    string        `json:"method"`
	Path      string        `json:"path"`
	Latency   time.Duration `json:"latency"`
	UserAgent string        `json:"user_agent"`
	Header    string        `json:"header"`
}

// Logger is a custom logger middleware for the Gin framework
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Process request
		c.Next()

		// Stop timer
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		// Get http status code
		httpCode := c.Writer.Status()

		// Get client IP
		clientIP := c.ClientIP()

		// Get request method
		method := c.Request.Method

		// Get request path
		path := c.Request.URL.Path

		// parsing header to object
		slog.Info("Request",
			slog.String("client_ip", clientIP),
			slog.Int("http_code", httpCode),
			slog.String("method", method),
			slog.String("path", path),
			slog.String("latency", fmt.Sprintf("%v", latency)),
			slog.String("header", fmt.Sprintf("%v", c.Request.Header)),
		)
	}
}
