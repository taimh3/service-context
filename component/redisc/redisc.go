package redisc

import (
	"context"
	"flag"
	"log/slog"
	"strings"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"

	sctx "github.com/taimaifika/service-context"
)

type config struct {
	url      string
	username string
	password string

	// Enable OpenTelemetry instrumentation.
	isOpenTelemetry bool
	// Enable tracing instrumentation.
	isOpenTelemetryTraces bool
	// Enable metrics instrumentation.
	isOpenTelemetryMetrics bool
}

type redisComponent struct {
	id string

	*config

	redis *redis.ClusterClient
}

func NewRedisComponent(id string) *redisComponent {
	return &redisComponent{
		id:     id,
		config: new(config),
	}
}

func (r *redisComponent) healthCheck() error {
	// health check
	// ping redis
	_, err := r.redis.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *redisComponent) GetRedis() *redis.ClusterClient {
	return r.redis
}

func (r *redisComponent) ID() string {
	return r.id
}

func (r *redisComponent) InitFlags() {
	flag.StringVar(&r.url, r.id+"-url", "localhost-0:6379,localhost-1:6379,localhost-2:6379", "redis urls. default: localhost-0:6379,localhost-1:6379,localhost-2:6379")
	flag.StringVar(&r.username, r.id+"-username", "", "redis username. default: ''")
	flag.StringVar(&r.password, r.id+"-password", "", "redis password. default: ''")

	// OpenTelemetry flags
	flag.BoolVar(&r.isOpenTelemetry, r.id+"-is-otel", false, "enable OpenTelemetry instrumentation. default: false")
	flag.BoolVar(&r.isOpenTelemetryTraces, r.id+"-is-otel-traces", false, "enable OpenTelemetry tracing instrumentation. default: false")
	flag.BoolVar(&r.isOpenTelemetryMetrics, r.id+"-is-otel-metrics", false, "enable OpenTelemetry metrics instrumentation. default: false")
}

func (r *redisComponent) Activate(ctx sctx.ServiceContext) error {
	opts := &redis.ClusterOptions{
		Addrs: strings.Split(r.url, ","),
	}

	// set username and password
	if r.username != "" && r.password != "" {
		opts.Username = r.username
		opts.Password = r.password
	}

	r.redis = redis.NewClusterClient(opts)

	// OpenTelemetry instrumentation
	// Just ensure the OpenTelemetry SDK is initialized in your application.
	if r.isOpenTelemetry {
		slog.Info("OpenTelemetry instrumentation enabled")
		if r.isOpenTelemetryTraces {
			slog.Info("Tracing instrumentation enabled")
			if err := redisotel.InstrumentTracing(r.redis); err != nil {
				return err
			}
		}

		if r.isOpenTelemetryMetrics {
			slog.Info("Metrics instrumentation enabled")
			if err := redisotel.InstrumentMetrics(r.redis); err != nil {
				return err
			}
		}
	}

	slog.Info("Connect to redis...")

	// health check
	err := r.healthCheck()
	if err != nil {
		return err
	}

	slog.Info("Connect to redis success")

	return nil
}

func (r *redisComponent) Stop() error {
	return nil
}
