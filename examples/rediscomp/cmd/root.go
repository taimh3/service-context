package cmd

import (
	"fmt"
	"log"
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
			log.Fatal(err)
		}

		// redisc := serviceCtx.MustGet("redis").(RedisComponent)
		// redis := redisc.GetRedis()

	},
}

func Execute() {
	rootCmd.AddCommand(outEnvCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
