package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	sctx "github.com/taimaifika/service-context"
	"github.com/taimaifika/service-context/component/ginc"
	"github.com/taimaifika/service-context/component/ginc/middleware"
)

type GINComponent interface {
	GetPort() int
	GetRouter() *gin.Engine
}

func main() {
	const compId = "gin"

	serviceCtx := sctx.NewServiceContext(
		sctx.WithName("simple-gin-http"),
		sctx.WithComponent(ginc.NewGin(compId)),
	)

	if err := serviceCtx.Load(); err != nil {
		log.Fatal(err)
	}

	comp := serviceCtx.MustGet(compId).(GINComponent)

	router := comp.GetRouter()
	router.Use(
		gin.Logger(),
		middleware.AllowCORS(),
		gin.Recovery(),
	)

	// Demo serve a handler with service-context
	router.GET("/demo", demoHdl(serviceCtx))

	// Source code from: https://gin-gonic.com/docs/examples/graceful-restart-or-stop/
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", comp.GetPort()),
		Handler: router,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen error", slog.String("error", err.Error()))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server Shutdown", slog.String("error", err.Error()))
	}

	<-ctx.Done()
	slog.Info("timeout of 5 seconds.")

	_ = serviceCtx.Stop()
	slog.Info("Server exited")
}

func demoHdl(serviceCtx sctx.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		slog.Info("Service %s is running with % env\n", serviceCtx.GetName(), serviceCtx.EnvName())

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}
