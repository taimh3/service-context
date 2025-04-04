package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	sctx "github.com/taimaifika/service-context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type CanGetStatusCode interface {
	StatusCode() int
}

func Recovery(serviceCtx sctx.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {

			if err := recover(); err != nil {
				// OpenTelemetry
				ctx, span := otel.Tracer("gin.middleware.recovery").Start(c, "Recovery")
				span.SetStatus(codes.Error, "Gin middleware recovered")
				span.SetAttributes(attribute.String("gin.middleware.recovery.url", c.Request.URL.String()))
				span.SetAttributes(attribute.String("gin.middleware.recovery.method", c.Request.Method))
				span.SetAttributes(attribute.String("gin.middleware.recovery.header", fmt.Sprintf("%+v\n", c.Request.Header)))
				span.SetAttributes(attribute.String("gin.middleware.recovery.error", fmt.Sprintf("%+v\n", err)))
				defer span.End()

				// Response with error
				if appErr, ok := err.(CanGetStatusCode); ok {
					slog.ErrorContext(ctx, "Gin middleware recovered", "error", appErr)
					c.AbortWithStatusJSON(appErr.StatusCode(), appErr)
				} else {
					// General panic cases
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
						"code":    http.StatusInternalServerError,
						"status":  "internal server error",
						"message": "something went wrong, please try again or contact supporters",
					})
				}

				slog.ErrorContext(ctx, "Gin middleware recovered", "error", err)
			}
		}()
		c.Next()
	}
}
