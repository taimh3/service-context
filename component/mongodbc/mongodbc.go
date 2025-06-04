package mongodbc

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	sctx "github.com/taimaifika/service-context"
)

type config struct {
	url                    string
	username               string
	password               string
	authMechanism          string
	authSource             string
	maxPoolSize            uint64
	minPoolSize            uint64
	timeout                time.Duration
	connectionTimeout      time.Duration
	ServerSelectionTimeout time.Duration
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

// GetMongoClient returns the mongo client
func (m *mongoDbComponent) GetMongoClient() *mongo.Client {
	return m.mongoClient
}

// GetMongoClientWithName returns the mongo client with the given name in config
func (m *mongoDbComponent) GetDatabase() *mongo.Database {
	return m.mongoClient.Database(m.config.authSource)
}

// GetCollection returns the mongo client with the given collection name
func (m *mongoDbComponent) GetCollection(collectionName string) *mongo.Collection {
	return m.mongoClient.Database(m.config.authSource).Collection(collectionName)
}

// GetDatabaseName returns the mongo client with the given name in config
func (m *mongoDbComponent) GetDatabaseName() string {
	return m.config.authSource
}

// GetDatabaseWithName returns the mongo client with the given name
func (m *mongoDbComponent) GetDatabaseWithName(dbName string) *mongo.Database {
	return m.mongoClient.Database(dbName)
}

// GetCollection returns the mongo client with the given name
func (m *mongoDbComponent) GetCollectionWithDatabase(dbName, collectionName string) *mongo.Collection {
	return m.mongoClient.Database(dbName).Collection(collectionName)
}

func (m *mongoDbComponent) ID() string {
	return m.id
}

func (m *mongoDbComponent) InitFlags() {
	flag.StringVar(&m.url, m.id+"-url", "mongodb://localhost:27017", "redis urls. default: mongodb://localhost:27017")

	flag.StringVar(&m.username, m.id+"-username", "", "mongodb username. default: ''")
	flag.StringVar(&m.password, m.id+"-password", "", "mongodb password. default: ''")
	flag.StringVar(&m.authMechanism, m.id+"-auth-mechanism", "SCRAM-SHA-256", "AuthMechanism supported values include SCRAM-SHA-256, SCRAM-SHA-1, MONGODB-CR, PLAIN, GSSAPI, MONGODB-X509, and MONGODB-AWS. default: 'SCRAM-SHA-256'")
	flag.StringVar(&m.authSource, m.id+"-auth-source", "", "AuthSource. default: ''")

	flag.Uint64Var(&m.minPoolSize, m.id+"-min-pool-size", 10, "mongodb min pool size. default: 10")
	flag.Uint64Var(&m.maxPoolSize, m.id+"-max-pool-size", 100, "mongodb max pool size. default: 100")

	flag.DurationVar(&m.timeout, m.id+"-timeout", 10*time.Second, "mongodb timeout. default: 10s")
	flag.DurationVar(&m.connectionTimeout, m.id+"-connection-timeout", 30*time.Second, "mongodb connection timeout. default: 30s")
	flag.DurationVar(&m.ServerSelectionTimeout, m.id+"-server-selection-timeout", 30*time.Second, "mongodb server selection timeout. default: 30s")

}

func (m *mongoDbComponent) Activate(ctx sctx.ServiceContext) error {
	// create mongo client
	opts := options.Client()
	// set url
	opts.ApplyURI(m.url)

	// set min pool size
	opts.SetMinPoolSize(m.minPoolSize)

	// set max pool size
	opts.SetMaxPoolSize(m.maxPoolSize)

	// set timeout
	opts.SetTimeout(m.timeout)

	// set connection timeout
	opts.SetConnectTimeout(m.connectionTimeout)

	// set ConnectTimeout
	opts.SetServerSelectionTimeout(m.ServerSelectionTimeout)

	// set auth database
	if m.username != "" && m.password != "" {
		opts.SetAuth(options.Credential{
			Username:      m.username,
			Password:      m.password,
			AuthMechanism: m.authMechanism,
			AuthSource:    m.authSource,
		})
	}
	// validate auth source, return err
	if opts.Auth.AuthSource == "" {
		return errors.New("auth source is empty")
	}

	slog.Info("Connecting to mongo db ...")

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

	slog.Info("Connect to mongo db success !!!")

	return nil
}

func (m *mongoDbComponent) Stop() error {
	return nil
}
