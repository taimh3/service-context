package cmd

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	sctx "github.com/taimaifika/service-context"
	"github.com/taimaifika/service-context/component/ginc"
	"github.com/taimaifika/service-context/component/otelc"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func newServiceCtx() sctx.ServiceContext {
	return sctx.NewServiceContext(
		sctx.WithName("otel-component"),
		sctx.WithComponent(ginc.NewGin("gin")),
		sctx.WithComponent(otelc.NewOtel("otel")),
	)
}

type GINComponent interface {
	GetPort() int
	GetRouter() *gin.Engine
}

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "Start GIN-HTTP service",
	Run: func(cmd *cobra.Command, args []string) {
		serviceCtx := newServiceCtx()

		if err := serviceCtx.Load(); err != nil {
			slog.Error("failed to load service context", slog.Any("error", err))
			os.Exit(1)
		}

		comp := serviceCtx.MustGet("gin").(GINComponent)

		router := comp.GetRouter()

		router.Use(gin.Recovery(), gin.Logger(), otelgin.Middleware(getHostname()))

		router.GET("/ping", func(c *gin.Context) {
			ctx := c.Request.Context()
			slog.InfoContext(ctx, "ping request received",
				slog.String("client_ip", c.ClientIP()),
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
			)
			slog.WarnContext(ctx, "this is a warning test log to verify otel logs")
			slog.ErrorContext(ctx, "this is an error test log to verify otel logs", slog.Int("status_code", http.StatusOK))

			// Set status code on the span
			span := trace.SpanFromContext(ctx)
			span.SetStatus(codes.Ok, "ping success")

			c.JSON(http.StatusOK, gin.H{"data": "pong"})
		})

		if err := router.Run(fmt.Sprintf(":%d", comp.GetPort())); err != nil {
			slog.Error("router execution failed", slog.Any("error", err))
			os.Exit(1)
		}
	},
}

// getHostname returns the hostname of the machine
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		slog.Error("failed to get hostname", slog.Any("error", err))
		os.Exit(1)
	}
	return hostname
}

func Execute() {
	rootCmd.AddCommand(outEnvCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
