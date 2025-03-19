package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"

	sctx "github.com/taimaifika/service-context"
	"github.com/taimaifika/service-context/component/redisc"
)

func newServiceCtx() sctx.ServiceContext {
	return sctx.NewServiceContext(
		sctx.WithComponent(redisc.NewRedisComponent("redis")),
	)
}

type RedisComponent interface {
	GetRedis() *redis.ClusterClient
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
		redis := redisc.GetRedis()

		// set data redis
		_, err := redis.Set(context.Background(), "test_connection_key", "data 123123", 0).Result()
		if err != nil {
			slog.Error("set data error", "error", err)
			return
		}
		// get data redis
		result, err := redis.Get(context.Background(), "test_connection_key").Result()
		if err != nil {
			slog.Error("get data error", "error", err)
			return
		}

		fmt.Println(result)

	},
}

func Execute() {
	rootCmd.AddCommand(outEnvCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
