package scylladbc

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v3"
	sctx "github.com/taimaifika/service-context"
)

type config struct {
	// ScyllaDB configuration
	hosts    []string
	hostsStr string // Comma-separated list of hosts
	username string
	password string

	// Timeouts
	timeout        time.Duration // Timeout for queries
	connectTimeout time.Duration // Timeout for establishing a connection

	// Keyspace
	ks                         string
	ksClass                    string // Class for keyspace replication, e.g., SimpleStrategy, NetworkTopologyStrategy
	ksReplicationFactor        int    // Replication factor for the keyspace, e.g., 1 for SimpleStrategy
	ksDisableInitialHostLookup bool   // Disable initial host lookup
	ksNumConns                 int    // Number of connections to use
}

type scyllaDbComponent struct {
	id string
	*config

	cluster *gocql.ClusterConfig // ScyllaDB cluster configuration
}

func NewScyllaDbComponent(id string) *scyllaDbComponent {
	return &scyllaDbComponent{
		id:     id,
		config: new(config),
	}
}

func (s *scyllaDbComponent) ID() string {
	return s.id
}

func (s *scyllaDbComponent) InitFlags() {
	flag.StringVar(&s.hostsStr, s.id+"-hosts", "localhost:9042,localhost:9043,localhost:9044", "List of ScyllaDB hosts, not empty (e.g. localhost:9042,localhost:9043,localhost:9044)")
	flag.StringVar(&s.config.username, s.id+"-username", "", "ScyllaDB username for authentication")
	flag.StringVar(&s.config.password, s.id+"-password", "", "ScyllaDB password for authentication")

	flag.DurationVar(&s.config.timeout, s.id+"-timeout", 10*time.Second, "Timeout for ScyllaDB queries, e.g. 10s")
	flag.DurationVar(&s.config.connectTimeout, s.id+"-connect-timeout", 10*time.Second, "Timeout for establishing a connection to ScyllaDB, e.g. 10s")

	flag.StringVar(&s.config.ks, s.id+"-keyspace", "", "ScyllaDB keyspace to use, not empty (e.g. catalog, admin, etc.)")
	flag.StringVar(&s.config.ksClass, s.id+"-keyspace-class", "NetworkTopologyStrategy", "ScyllaDB keyspace replication class (e.g. SimpleStrategy, NetworkTopologyStrategy)")
	flag.IntVar(&s.config.ksReplicationFactor, s.id+"-keyspace-replication-factor", 1, "ScyllaDB keyspace replication factor (e.g. 1 for SimpleStrategy)")
	flag.BoolVar(&s.config.ksDisableInitialHostLookup, s.id+"-keyspace-disable-initial-host-lookup", true, "Disable initial host lookup for ScyllaDB keyspace")
	flag.IntVar(&s.config.ksNumConns, s.id+"-keyspace-num-conns", 10, "Number of connections to use for ScyllaDB keyspace")
}

func (s *scyllaDbComponent) Activate(ctx sctx.ServiceContext) error {
	if s.hostsStr == "" || s.config.ks == "" {
		return fmt.Errorf("hosts or keyspace not configured: hosts=%s, keyspace=%s", s.hostsStr, s.config.ks)
	}

	// parse the hosts from the comma-separated string
	s.hosts = strings.Split(s.hostsStr, ",")

	// Create a new ScyllaDB cluster configuration
	cluster := gocql.NewCluster(s.config.hosts...)
	cluster.Keyspace = s.config.ks
	cluster.Consistency = gocql.LocalQuorum
	cluster.Timeout = s.config.timeout
	cluster.ConnectTimeout = s.config.connectTimeout
	cluster.NumConns = s.config.ksNumConns

	// Note: TokenAwareHostPolicy will be set per session to avoid sharing between sessions
	// This follows best practice to avoid "sharing token aware host selection policy between sessions is not supported" error

	// Disable initial host lookup, it helps with faster startup
	cluster.DisableInitialHostLookup = s.config.ksDisableInitialHostLookup

	// If username and password are provided, set the authenticator
	// This is optional, if not provided, it will connect without authentication
	if s.config.username != "" && s.config.password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: s.config.username,
			Password: s.config.password,
		}
	}

	// Set the cluster configuration to the component
	s.cluster = cluster

	// Log successful activation
	log.Printf("ScyllaDB component %s activated successfully", s.id)

	return nil
}

func (s *scyllaDbComponent) Stop() error {
	return nil
}

// GetCluster returns the ScyllaDB cluster configuration
func (s *scyllaDbComponent) GetCluster() *gocql.ClusterConfig {
	return s.cluster
}

// CreateSession creates a new ScyllaDB session
func (s *scyllaDbComponent) CreateSession() (*gocql.Session, error) {
	if s.cluster == nil {
		return nil, fmt.Errorf("ScyllaDB cluster is not initialized")
	}

	// Create a copy of the cluster config to avoid sharing policies between sessions
	clusterCopy := *s.cluster

	// Set TokenAwareHostPolicy for this session to enhance performance
	// It routes requests to the replica that is most likely to have the data,
	// reducing latency and improving throughput.
	clusterCopy.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())

	session, err := clusterCopy.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create ScyllaDB session: %w", err)
	}

	return session, nil
}

// CreateSessionWithGoCqlX creates a new ScyllaDB session using gocqlx v3
func (s *scyllaDbComponent) CreateSessionWithGoCqlX() (*gocqlx.Session, error) {
	if s.cluster == nil {
		return nil, fmt.Errorf("ScyllaDB cluster is not initialized")
	}

	// Create a copy of the cluster config to avoid sharing policies between sessions
	clusterCopy := *s.cluster

	// Set TokenAwareHostPolicy for this session to enhance performance
	// It routes requests to the replica that is most likely to have the data,
	// reducing latency and improving throughput.
	clusterCopy.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())

	session, err := gocqlx.WrapSession(clusterCopy.CreateSession())
	if err != nil {
		return nil, fmt.Errorf("failed to create ScyllaDB session with gocqlx v3: %w", err)
	}

	return &session, nil
}
