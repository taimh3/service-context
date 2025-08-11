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
	"github.com/spf13/cobra"
	"github.com/taimaifika/service-context/component/ginc"
	"github.com/taimaifika/service-context/component/ginc/middleware"
	"github.com/taimaifika/service-context/component/mongodbc"
	"github.com/taimaifika/service-context/component/otelc"
	"github.com/taimaifika/service-context/component/slogc"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	sctx "github.com/taimaifika/service-context"
)

func newServiceCtx() sctx.ServiceContext {
	return sctx.NewServiceContext(
		// logging
		sctx.WithComponent(slogc.NewSlogComponent()),
		// observability (traces/metrics/log export)
		sctx.WithComponent(otelc.NewOtel("otel")),
		// http server (gin) - optional for this example, initialized but not started
		sctx.WithComponent(ginc.NewGin("gin")),
		// mongodb component
		sctx.WithComponent(mongodbc.NewMongoDbComponent("mongodb")),
	)
}

type MongoDbComponent interface {
	GetMongoClient() *mongo.Client
	GetDatabaseName() string
}

type GINComponent interface {
	GetPort() int
	GetRouter() *gin.Engine
}

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "Start mongodb api service",
	Run: func(cmd *cobra.Command, args []string) {
		serviceCtx := newServiceCtx()

		if err := serviceCtx.Load(); err != nil {
			slog.Error("load service context error", "error", err)
			panic(err)
		}

		mongoComp := serviceCtx.MustGet("mongodb").(MongoDbComponent)
		ginComp := serviceCtx.MustGet("gin").(GINComponent)

		router := ginComp.GetRouter()
		// middlewares
		router.Use(
			gin.Logger(),
			middleware.AllowCORS(),
			gin.Recovery(),
		)

		db := mongoComp.GetMongoClient().Database(mongoComp.GetDatabaseName())
		coll := db.Collection("test")

		type TestDoc struct {
			ID   interface{} `json:"id" bson:"_id,omitempty"`
			Test string      `json:"test" bson:"test"`
		}

		// health check
		router.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// list documents
		router.GET("/tests", func(c *gin.Context) {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			defer cancel()
			cursor, err := coll.Find(ctx, bson.M{})
			if err != nil {
				slog.Error("find documents", "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer cursor.Close(ctx)
			var docs []TestDoc
			for cursor.Next(ctx) {
				var d TestDoc
				if err := cursor.Decode(&d); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				docs = append(docs, d)
			}
			c.JSON(http.StatusOK, gin.H{"data": docs})
		})

		// create document
		router.POST("/tests", func(c *gin.Context) {
			var body struct {
				Test string `json:"test"`
			}
			if err := c.ShouldBindJSON(&body); err != nil || body.Test == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
				return
			}
			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			defer cancel()
			res, err := coll.InsertOne(ctx, bson.M{"test": body.Test})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"inserted_id": res.InsertedID})
		})

		// get one document by id
		router.GET("/tests/:id", func(c *gin.Context) {
			idHex := c.Param("id")
			objID, err := bson.ObjectIDFromHex(idHex)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
				return
			}
			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			defer cancel()
			var doc TestDoc
			if err := coll.FindOne(ctx, bson.M{"_id": objID}).Decode(&doc); err != nil {
				if err == mongo.ErrNoDocuments {
					c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": doc})
		})

		// delete one document by id
		router.DELETE("/tests/:id", func(c *gin.Context) {
			idHex := c.Param("id")
			objID, err := bson.ObjectIDFromHex(idHex)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
				return
			}
			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			defer cancel()
			res, err := coll.DeleteOne(ctx, bson.M{"_id": objID})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if res.DeletedCount == 0 {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"deleted": res.DeletedCount})
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
