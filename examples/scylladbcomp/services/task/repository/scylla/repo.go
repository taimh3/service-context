package scylla

import (
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v3"
)

type scyllaRepo struct {
	cluster           *gocql.ClusterConfig // ScyllaDB cluster configuration
	session           *gocql.Session       // ScyllaDB session
	sessionWithGocqlX *gocqlx.Session      // ScyllaDB session with gocqlx
}

func NewScyllaRepo(
	cluster *gocql.ClusterConfig,
	session *gocql.Session,
	sessionWithGocqlX *gocqlx.Session,
) *scyllaRepo {
	return &scyllaRepo{
		cluster:           cluster,
		session:           session,
		sessionWithGocqlX: sessionWithGocqlX,
	}
}
