package scylladbc

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gocql/gocql"
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
	ks                  string
	ksClass             string // Class for keyspace replication, e.g., SimpleStrategy, NetworkTopologyStrategy
	ksReplicationFactor int    // Replication factor for the keyspace, e.g., 1 for SimpleStrategy
}

type scyllaDbComponent struct {
	id string
	*config

	session *gocql.Session
}

func NewScyllaDbComponent(id string) *scyllaDbComponent {
	return &scyllaDbComponent{
		id: id,
		config: &config{
			hosts:    []string{"127.0.0.1"},
			ks:       "catalog",
			username: "",
			password: "",
		},
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
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = s.config.timeout
	cluster.ConnectTimeout = s.config.connectTimeout

	// If username and password are provided, set the authenticator
	// This is optional, if not provided, it will connect without authentication
	if s.config.username != "" && s.config.password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: s.config.username,
			Password: s.config.password,
		}
	}

	// Create a session to the ScyllaDB cluster
	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	s.session = session

	// Log successful activation
	log.Printf("ScyllaDB component %s activated successfully", s.id)

	// Create the keyspace if it does not exist
	if s.config.ks != "" {
		createKeyspaceQuery := fmt.Sprintf(`
			CREATE KEYSPACE IF NOT EXISTS %s
			WITH REPLICATION = {'class': '%s', 'replication_factor': %d}
		`, s.config.ks, s.config.ksClass, s.config.ksReplicationFactor)
		if err := s.session.Query(createKeyspaceQuery).Exec(); err != nil {
			return fmt.Errorf("failed to create keyspace: %w", err)
		}
	}

	return nil
}

func (s *scyllaDbComponent) Stop() error {
	if s.session != nil {
		s.session.Close()
		log.Printf("ScyllaDB component %s stopped", s.id)
	}
	return nil
}

// GetScyllaDb returns the ScyllaDB session
func (s *scyllaDbComponent) GetSession() *gocql.Session {
	if s.session == nil {
		log.Fatal("ScyllaDB session is not initialized")
	}
	return s.session
}
