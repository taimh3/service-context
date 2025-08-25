package cmd

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"

	"github.com/taimaifika/service-context/examples/scylladbcomp/common"
	"github.com/taimaifika/service-context/examples/scylladbcomp/composer"

	sctx "github.com/taimaifika/service-context"
	"github.com/taimaifika/service-context/component/ginc"
	"github.com/taimaifika/service-context/component/ginc/middleware"
	"github.com/taimaifika/service-context/component/otelc"
	"github.com/taimaifika/service-context/component/scylladbc"
	"github.com/taimaifika/service-context/component/slogc"
)

const serviceName = "simple-clean-architecture"

func newServiceCtx() sctx.ServiceContext {
	return sctx.NewServiceContext(
		sctx.WithName(serviceName),
		sctx.WithComponent(slogc.NewSlogComponent()),
		sctx.WithComponent(otelc.NewOtel(common.KeyCompOtel)),
		sctx.WithComponent(ginc.NewGin(common.KeyCompGin)),
		sctx.WithComponent(scylladbc.NewScyllaDbComponent(common.KeyCompScylla)),

		sctx.WithComponent(NewConfig()),
	)
}

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "Start service",
	Run: func(cmd *cobra.Command, args []string) {
		serviceCtx := newServiceCtx()

		if err := serviceCtx.Load(); err != nil {
			slog.Error("Service context load error", "error", err)
			panic(err)
		}

		// Load Gin component
		ginComp := serviceCtx.MustGet(common.KeyCompGin).(common.GinComponent)

		router := ginComp.GetRouter()
		router.Use(
			middleware.AllowCORS(),
			otelgin.Middleware(
				serviceCtx.GetName(),
				otelgin.WithTracerProvider(otel.GetTracerProvider()),
				otelgin.WithFilter(func(req *http.Request) bool {
					// Skip tracing for /ping endpoints
					return req.URL.Path != "/ping"
				}),
			),
		)

		router.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})

		// Example routes
		v1 := router.Group("/v1")
		exampleRoutes(v1, serviceCtx)

		// Start the server
		if err := router.Run(fmt.Sprintf(":%d", ginComp.GetPort())); err != nil {
			slog.Error("Service start error", "error", err)
			panic(err)
		}
	},
}

// exampleRoutes is an example of how to define routes (task management)
func exampleRoutes(router *gin.RouterGroup, serviceCtx sctx.ServiceContext) {

	taskApiService := composer.ComposeTaskApiService(serviceCtx)

	// scylla
	scyllaTaskHandler := router.Group("/scylla/tasks")
	{
		scyllaTaskHandler.GET("", taskApiService.ScyllaListTaskHdl())
		scyllaTaskHandler.POST("", taskApiService.ScyllaCreateTaskHdl())
		scyllaTaskHandler.GET("/:task-id", taskApiService.ScyllaGetTaskHdl())
		scyllaTaskHandler.PATCH("/:task-id", taskApiService.ScyllaUpdateTaskHdl())
		scyllaTaskHandler.DELETE("/:task-id", taskApiService.ScyllaDeleteTaskHdl())
	}

	scyllaPersonHandler := router.Group("/scylla/persons")
	{
		scyllaPersonHandler.POST("", taskApiService.ScyllaCreatePersonHdl())
		scyllaPersonHandler.GET("", taskApiService.ScyllaListPersonHdl())
	}
}

func Execute() {
	rootCmd.AddCommand(outEnvCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
