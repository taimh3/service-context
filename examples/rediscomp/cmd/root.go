package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	sctx "github.com/taimaifika/service-context"
	"github.com/taimaifika/service-context/component/ginc"
	"github.com/taimaifika/service-context/component/ginc/middleware"
	"github.com/taimaifika/service-context/component/otelc"
	"github.com/taimaifika/service-context/component/redisc"
	"github.com/taimaifika/service-context/component/slogc"
)

var serviceContextName = "service-context-redis"

func newServiceCtx() sctx.ServiceContext {
	return sctx.NewServiceContext(
		sctx.WithComponent(slogc.NewSlogComponent()),
		sctx.WithComponent(otelc.NewOtel("otel")),
		sctx.WithComponent(ginc.NewGin("gin")),
		sctx.WithComponent(redisc.NewRedisComponent("redis")),
	)
}

type RedisComponent interface {
	GetRedis() *redis.ClusterClient
}

type GINComponent interface {
	GetPort() int
	GetRouter() *gin.Engine
}

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "Start redis service",
	Run: func(cmd *cobra.Command, args []string) {
		serviceCtx := newServiceCtx()

		if err := serviceCtx.Load(); err != nil {
			slog.Error("load service context error", "error", err)
			panic(err)
		}

		redisc := serviceCtx.MustGet("redis").(RedisComponent)
		ginComp := serviceCtx.MustGet("gin").(GINComponent)
		redis := redisc.GetRedis()

		router := ginComp.GetRouter()
		// middlewares
		router.Use(
			middleware.Logger(),
			middleware.AllowCORS(),
			middleware.Recovery(serviceCtx),
			otelgin.Middleware(serviceContextName),
		)

		// health check
		router.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// test redis connection
		router.GET("/redis/test", func(c *gin.Context) {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			defer cancel()

			// set test data
			err := redis.Set(ctx, "test_connection_key", "data 123123", 0).Err()
			if err != nil {
				slog.Error("set data error", "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// get test data
			result, err := redis.Get(ctx, "test_connection_key").Result()
			if err != nil {
				slog.Error("get data error", "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"result": result})
		})

		// set key-value
		router.POST("/redis/set", func(c *gin.Context) {
			var body struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			}
			if err := c.ShouldBindJSON(&body); err != nil || body.Key == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
				return
			}

			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			defer cancel()

			err := redis.Set(ctx, body.Key, body.Value, 0).Err()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "key set successfully"})
		})

		// get value by key
		router.GET("/redis/get/:key", func(c *gin.Context) {
			key := c.Param("key")
			if key == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
				return
			}

			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			defer cancel()

			result, err := redis.Get(ctx, key).Result()
			if err != nil {
				if err.Error() == "redis: nil" {
					c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"key": key, "value": result})
		})

		// delete key
		router.DELETE("/redis/del/:key", func(c *gin.Context) {
			key := c.Param("key")
			if key == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
				return
			}

			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			defer cancel()

			deleted, err := redis.Del(ctx, key).Result()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if deleted == 0 {
				c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"deleted": deleted})
		})

		srv := &http.Server{Addr: fmt.Sprintf(":%d", ginComp.GetPort()), Handler: router}

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("listen error", "error", err)
			}
		}()

		slog.Info("server started", "port", ginComp.GetPort())

		// graceful shutdown
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		slog.Info("shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
		_ = serviceCtx.Stop()
		slog.Info("server exited")
	},
}

func Execute() {
	rootCmd.AddCommand(outEnvCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
