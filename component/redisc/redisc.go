package redisc

import (
	"context"
	"flag"
	"log/slog"
	"strings"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"

	sctx "github.com/taimaifika/service-context"
)

type config struct {
	url      string
	username string
	password string
	db       int
	protocol int // 0 (default/auto), 2 (RESP2 for Redis 6/proxies), 3 (RESP3 for Redis 7/8)

	isCluster                 bool
	disableIdentity           bool
	disableMaintNotifications bool

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

	redis            *redis.ClusterClient
	standaloneClient *redis.Client
	universalClient  redis.UniversalClient
}

type Option func(*redisComponent)

func WithURL(url string) Option {
	return func(r *redisComponent) {
		r.url = url
	}
}

func WithUsername(username string) Option {
	return func(r *redisComponent) {
		r.username = username
	}
}

func WithPassword(password string) Option {
	return func(r *redisComponent) {
		r.password = password
	}
}

func WithDB(db int) Option {
	return func(r *redisComponent) {
		r.db = db
	}
}

func WithProtocol(protocol int) Option {
	return func(r *redisComponent) {
		r.protocol = protocol
	}
}

func WithIsCluster(isCluster bool) Option {
	return func(r *redisComponent) {
		r.isCluster = isCluster
	}
}

func WithDisableIdentity(disable bool) Option {
	return func(r *redisComponent) {
		r.disableIdentity = disable
	}
}

func WithDisableMaintNotifications(disable bool) Option {
	return func(r *redisComponent) {
		r.disableMaintNotifications = disable
	}
}

func WithOpenTelemetry(enable bool) Option {
	return func(r *redisComponent) {
		r.isOpenTelemetry = enable
	}
}

func WithOpenTelemetryTraces(enable bool) Option {
	return func(r *redisComponent) {
		r.isOpenTelemetryTraces = enable
	}
}

func WithOpenTelemetryMetrics(enable bool) Option {
	return func(r *redisComponent) {
		r.isOpenTelemetryMetrics = enable
	}
}

func NewRedisComponent(id string, opts ...Option) *redisComponent {
	r := &redisComponent{
		id: id,
		config: &config{
			isCluster:                 true,
			disableIdentity:           true,
			disableMaintNotifications: true,
			protocol:                  0,
			db:                        0,
		},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *redisComponent) healthCheck() error {
	if r.universalClient != nil {
		_, err := r.universalClient.Ping(context.Background()).Result()
		return err
	}
	return nil
}

// GetRedis returns the ClusterClient (for backwards compatibility).
func (r *redisComponent) GetRedis() *redis.ClusterClient {
	return r.redis
}

// GetStandaloneClient returns the standalone Client.
func (r *redisComponent) GetStandaloneClient() *redis.Client {
	return r.standaloneClient
}

// GetUniversalClient returns the UniversalClient interface supporting both Cluster and Standalone clients.
func (r *redisComponent) GetUniversalClient() redis.UniversalClient {
	return r.universalClient
}

func (r *redisComponent) ID() string {
	return r.id
}

func (r *redisComponent) InitFlags() {
	if flag.Lookup(r.id+"-url") == nil {
		flag.StringVar(&r.url, r.id+"-url", "localhost-0:6379,localhost-1:6379,localhost-2:6379", "redis urls. default: localhost-0:6379,localhost-1:6379,localhost-2:6379")
	}
	if flag.Lookup(r.id+"-username") == nil {
		flag.StringVar(&r.username, r.id+"-username", "", "redis username. default: ''")
	}
	if flag.Lookup(r.id+"-password") == nil {
		flag.StringVar(&r.password, r.id+"-password", "", "redis password. default: ''")
	}
	if flag.Lookup(r.id+"-db") == nil {
		flag.IntVar(&r.db, r.id+"-db", 0, "redis database index (standalone mode only). default: 0")
	}
	if flag.Lookup(r.id+"-protocol") == nil {
		flag.IntVar(&r.protocol, r.id+"-protocol", 0, "redis protocol (0: auto, 2: RESP2, 3: RESP3). default: 0")
	}
	if flag.Lookup(r.id+"-is-cluster") == nil {
		flag.BoolVar(&r.isCluster, r.id+"-is-cluster", true, "redis cluster mode (true for cluster, false for standalone). default: true")
	}
	if flag.Lookup(r.id+"-disable-identity") == nil {
		flag.BoolVar(&r.disableIdentity, r.id+"-disable-identity", true, "disable CLIENT SETINFO command on connection init (for Redis < 7.2). default: true")
	}
	if flag.Lookup(r.id+"-disable-maint-notifications") == nil {
		flag.BoolVar(&r.disableMaintNotifications, r.id+"-disable-maint-notifications", true, "disable CLIENT MAINT_NOTIFICATIONS command on connect. default: true")
	}

	// OpenTelemetry flags
	if flag.Lookup(r.id+"-is-otel") == nil {
		flag.BoolVar(&r.isOpenTelemetry, r.id+"-is-otel", false, "enable OpenTelemetry instrumentation. default: false")
	}
	if flag.Lookup(r.id+"-is-otel-traces") == nil {
		flag.BoolVar(&r.isOpenTelemetryTraces, r.id+"-is-otel-traces", false, "enable OpenTelemetry tracing instrumentation. default: false")
	}
	if flag.Lookup(r.id+"-is-otel-metrics") == nil {
		flag.BoolVar(&r.isOpenTelemetryMetrics, r.id+"-is-otel-metrics", false, "enable OpenTelemetry metrics instrumentation. default: false")
	}
}

func (r *redisComponent) Activate(ctx sctx.ServiceContext) error {
	addrs := strings.Split(r.url, ",")

	if r.isCluster {
		opts := &redis.ClusterOptions{
			Addrs:           addrs,
			DisableIdentity: r.disableIdentity,
			Protocol:        r.protocol,
		}

		if r.disableMaintNotifications {
			opts.MaintNotificationsConfig = &maintnotifications.Config{
				Mode: maintnotifications.ModeDisabled,
			}
		}

		// Support password-only authentication (common in Redis 6.x) as well as ACL username+password
		if r.password != "" {
			opts.Password = r.password
			if r.username != "" {
				opts.Username = r.username
			}
		}

		r.redis = redis.NewClusterClient(opts)
		r.universalClient = r.redis
	} else {
		addr := "localhost:6379"
		if len(addrs) > 0 && strings.TrimSpace(addrs[0]) != "" {
			addr = strings.TrimSpace(addrs[0])
		}

		opts := &redis.Options{
			Addr:            addr,
			DB:              r.db,
			DisableIdentity: r.disableIdentity,
			Protocol:        r.protocol,
		}

		if r.disableMaintNotifications {
			opts.MaintNotificationsConfig = &maintnotifications.Config{
				Mode: maintnotifications.ModeDisabled,
			}
		}

		// Support password-only authentication (common in Redis 6.x) as well as ACL username+password
		if r.password != "" {
			opts.Password = r.password
			if r.username != "" {
				opts.Username = r.username
			}
		}

		r.standaloneClient = redis.NewClient(opts)
		r.universalClient = r.standaloneClient
	}

	// OpenTelemetry instrumentation
	if r.isOpenTelemetry {
		slog.Info("OpenTelemetry instrumentation enabled")
		if r.isOpenTelemetryTraces {
			slog.Info("Tracing instrumentation enabled")
			if r.isCluster && r.redis != nil {
				if err := redisotel.InstrumentTracing(r.redis); err != nil {
					return err
				}
			} else if r.standaloneClient != nil {
				if err := redisotel.InstrumentTracing(r.standaloneClient); err != nil {
					return err
				}
			}
		}

		if r.isOpenTelemetryMetrics {
			slog.Info("Metrics instrumentation enabled")
			if r.isCluster && r.redis != nil {
				if err := redisotel.InstrumentMetrics(r.redis); err != nil {
					return err
				}
			} else if r.standaloneClient != nil {
				if err := redisotel.InstrumentMetrics(r.standaloneClient); err != nil {
					return err
				}
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
	if r.universalClient != nil {
		return r.universalClient.Close()
	}
	return nil
}

