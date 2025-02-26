package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func LogrusLogger() gin.HandlerFunc {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{}) // Format log as JSON

	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate request processing time
		latency := time.Since(startTime)

		// Log with logrus
		logger.WithFields(logrus.Fields{
			"time":       time.Now().Format(time.RFC3339),
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"latency":    latency.Seconds(),
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"error":      c.Errors.ByType(gin.ErrorTypePrivate).String(),
		}).Info("Request log")
	}
}
