package cmd

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"

	"github.com/taimaifika/service-context/component/ginc"
	"github.com/taimaifika/service-context/component/ginc/middleware"
	"github.com/taimaifika/service-context/component/gormc"
	"github.com/taimaifika/service-context/component/otelc"
	"github.com/taimaifika/service-context/component/slogc"
	composer "github.com/taimaifika/service-context/examples/gormcomp/components"

	sctx "github.com/taimaifika/service-context"
)

var serviceContextName = "service-context-gorm"

func newServiceCtx() sctx.ServiceContext {
	return sctx.NewServiceContext(
		sctx.WithName(serviceContextName),
		sctx.WithComponent(slogc.NewSlogComponent()),
		sctx.WithComponent(otelc.NewOtel("otel")),
		sctx.WithComponent(ginc.NewGin("gin")),
		sctx.WithComponent(gormc.NewGormDB("postgres", "postgres")),
	)
}

type GINginComponent interface {
	GetPort() int
	GetRouter() *gin.Engine
}

type GormComponent interface {
	GetDB() *gorm.DB
}

type pgRepo struct {
	db *gorm.DB
}

func NewPgRepo(db *gorm.DB) *pgRepo {
	return &pgRepo{db: db}
}

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "Start GIN-HTTP service",
	Run: func(cmd *cobra.Command, args []string) {
		serviceCtx := newServiceCtx()

		if err := serviceCtx.Load(); err != nil {
			log.Fatal(err)
		}
		ginComp := serviceCtx.MustGet("gin").(GINginComponent)

		router := ginComp.GetRouter()

		router.Use(
			gin.Logger(), // format log to text
			middleware.Logger(),
			middleware.Recovery(serviceCtx),
			otelgin.Middleware(
				serviceContextName,
				otelgin.WithTracerProvider(otel.GetTracerProvider()),
			),
		)

		// ping endpoint
		router.GET("/ping", func(c *gin.Context) {
			_, span := otel.Tracer("service-context-gorm").Start(c.Request.Context(), "ping")
			defer span.End()
			slog.InfoContext(c, "This is an info message", slog.String("key", "value"))

			slog.Debug("This is a debug message")
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})

		// test panic endpoint
		router.GET("/panic", func(c *gin.Context) {
			ctx, span := otel.Tracer("service-context-gorm").Start(c.Request.Context(), "ping")
			defer span.End()

			testPanic(ctx)

			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})

		// gorm component
		db := serviceCtx.MustGet("postgres").(GormComponent)
		pgRepo := NewPgRepo(db.GetDB())

		router.GET("/number", func(c *gin.Context) {
			ctx, span := otel.Tracer("service-context-gorm").Start(c.Request.Context(), "get-tasks")
			defer span.End()

			var num int
			if err := pgRepo.db.WithContext(ctx).Raw("SELECT 42").Scan(&num).Error; err != nil {
				panic(err)
			}

			slog.Info("Number", slog.Int("number", num))

			c.JSON(http.StatusOK, gin.H{"data": num})
		})

		// task service
		taskApiService := composer.ComposeTaskApiService(serviceCtx)
		tasks := router.Group("/tasks")
		{
			tasks.GET("", taskApiService.ListTaskHdl())
		}

		// start the server
		if err := router.Run(fmt.Sprintf(":%d", ginComp.GetPort())); err != nil {
			log.Fatal(err)
		}
	},
}

func testPanic(ctx context.Context) {
	_, span := otel.Tracer("service-context-gorm").Start(ctx, "testPanic")
	defer span.End()
	slog.InfoContext(ctx, "This is an info message", slog.String("key", "value"))

	slog.Debug("This is a debug message")
	err := fmt.Errorf("test error")

	// set otel error
	span.SetStatus(codes.Error, "test error")
	// set otel status
	span.RecordError(err)

	panic("test panic")
}

func Execute() {
	rootCmd.AddCommand(outEnvCmd)
	slog.Info("Starting application")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
