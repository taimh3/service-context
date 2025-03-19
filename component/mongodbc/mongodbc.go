package mongodbc

import (
	"context"
	"flag"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	sctx "github.com/taimaifika/service-context"
)

type config struct {
	url      string
	username string
	password string
	timeout  time.Duration
}

type mongoDbComponent struct {
	id string

	*config

	mongoClient *mongo.Client
}

func NewMongoDbComponent(id string) *mongoDbComponent {
	return &mongoDbComponent{
		id:     id,
		config: new(config),
	}
}

func (m *mongoDbComponent) GetMongoClient() *mongo.Client {
	return m.mongoClient
}

func (m *mongoDbComponent) ID() string {
	return m.id
}

func (m *mongoDbComponent) InitFlags() {
	flag.StringVar(&m.url, m.id+"-url", "mongodb://localhost:27017", "redis urls. default: mongodb://localhost:27017")
	flag.StringVar(&m.username, m.id+"-username", "", "redis username. default: ''")
	flag.StringVar(&m.password, m.id+"-password", "", "redis password. default: ''")

	flag.DurationVar(&m.timeout, m.id+"-timeout", 10*time.Second, "redis timeout. default: 10s")
}

func (m *mongoDbComponent) Activate(ctx sctx.ServiceContext) error {
	// create mongo client
	opts := options.Client()
	// set url
	opts.ApplyURI(m.url)
	// set timeout
	opts.SetTimeout(m.timeout)

	// set auth
	if m.username != "" && m.password != "" {
		opts.SetAuth(options.Credential{
			Username: m.username,
			Password: m.password,
		})
	}

	slog.Info("Connecting to mongo db")

	client, err := mongo.Connect(opts)
	if err != nil {
		return err
	}

	// health check
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return err
	}

	m.mongoClient = client

	slog.Info("Connect to mongo db success")

	return nil
}

func (m *mongoDbComponent) Stop() error {
	return nil
}
