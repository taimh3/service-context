package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/taimaifika/service-context/component/mongodbc"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	sctx "github.com/taimaifika/service-context"
)

func newServiceCtx() sctx.ServiceContext {
	return sctx.NewServiceContext(
		sctx.WithComponent(mongodbc.NewMongoDbComponent("mongodb")),
	)
}

type MongoDbComponent interface {
	GetMongoClient() *mongo.Client
}

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "Start mongodb service",
	Run: func(cmd *cobra.Command, args []string) {
		serviceCtx := newServiceCtx()

		if err := serviceCtx.Load(); err != nil {
			slog.Error("load service context error", "error", err)
			panic(err)
		}

		// get mongodb component
		mongoDbComponent := serviceCtx.MustGet("mongodb").(MongoDbComponent)
		mongoClient := mongoDbComponent.GetMongoClient()

		// insert data
		res, err := mongoClient.Database("test").Collection("test").InsertOne(context.Background(), map[string]string{"test": "test"})
		if err != nil {
			slog.Error("insert data error", "error", err)
		}
		fmt.Println(res)

		// get data
		var result []map[string]interface{}
		cursor, err := mongoClient.Database("test").Collection("test").Find(
			context.Background(), bson.M{
				"test": "test",
			})
		if err != nil {
			slog.Error("find data error", "error", err)
		}
		defer cursor.Close(context.Background())
		if cursor.Next(context.Background()) {
			var item map[string]interface{}
			if err := cursor.Decode(&item); err != nil {
				slog.Error("decode data error", "error", err)
			}
			result = append(result, item)
		}
		fmt.Println(result)

		// delete data
		_, err = mongoClient.Database("test").Collection("test").DeleteOne(context.Background(), bson.M{"test": "test"})
		if err != nil {
			slog.Error("delete data error", "error", err)
		}
	},
}

func Execute() {
	rootCmd.AddCommand(outEnvCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
