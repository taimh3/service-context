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
	GetDatabaseName() string
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
		databaseName := mongoDbComponent.GetDatabaseName()

		// insert data
		res, err := mongoClient.Database(databaseName).Collection("test").InsertOne(context.Background(), map[string]string{"test": "test"})
		if err != nil {
			slog.Error("insert data error", "error", err)
		}
		fmt.Println(res)

		// get all data
		var result []map[string]interface{}
		cursor, err := mongoClient.Database(databaseName).Collection("test").Find(
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

		// find one data
		var resultOne map[string]interface{}
		idStr := "67ed19ea38a0ce1af87c4f5a"
		objID, err := bson.ObjectIDFromHex(idStr)
		if err != nil {
			slog.Error("convert id to object id error", "error", err)
		}
		filter := bson.M{"_id": objID}
		err = mongoClient.Database(databaseName).Collection("test").FindOne(
			context.Background(), filter).Decode(&resultOne)
		if err != nil {
			slog.Error("find one data error", "error", err)
		}
		fmt.Println(resultOne)

		// delete data
		_, err = mongoClient.Database(databaseName).Collection("test").DeleteOne(context.Background(), bson.M{"test": "test"})
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
