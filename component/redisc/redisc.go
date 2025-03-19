package redisc

import (
	"context"
	"flag"
	"log/slog"
	"strings"

	"github.com/redis/go-redis/v9"

	sctx "github.com/taimaifika/service-context"
)

type config struct {
	url      string
	username string
	password string
}

type RedisComponent struct {
	id string

	*config

	redis *redis.ClusterClient
}

func NewRedisComponent(id string) *RedisComponent {
	return &RedisComponent{
		id:     id,
		config: new(config),
	}
}

func (r *RedisComponent) healthCheck() error {
	// health check
	// ping redis
	_, err := r.redis.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	// set data redis
	_, err = r.redis.Set(context.Background(), "test", "test", 0).Result()
	if err != nil {
		return err
	}
	// get data redis
	_, err = r.redis.Get(context.Background(), "test").Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisComponent) GetRedis() *redis.ClusterClient {
	return r.redis
}

func (r *RedisComponent) ID() string {
	return r.id
}

func (r *RedisComponent) InitFlags() {
	flag.StringVar(&r.url, r.id+"-url", "localhost-0:6379,localhost-1:6379,localhost-2:6379", "redis urls. default: localhost-0:6379,localhost-1:6379,localhost-2:6379")
	flag.StringVar(&r.username, r.id+"-username", "", "redis username. default: ''")
	flag.StringVar(&r.password, r.id+"-password", "", "redis password. default: ''")
}

func (r *RedisComponent) Activate(ctx sctx.ServiceContext) error {
	opts := &redis.ClusterOptions{
		Addrs: strings.Split(r.url, ","),
	}

	// set username and password
	if r.username != "" && r.password != "" {
		opts.Username = r.username
		opts.Password = r.password
	}

	r.redis = redis.NewClusterClient(opts)

	slog.Info("Connect to redis...")

	// health check
	err := r.healthCheck()
	if err != nil {
		return err
	}

	slog.Info("Connect to redis success")

	return nil
}

func (r *RedisComponent) Stop() error {
	return nil
}
