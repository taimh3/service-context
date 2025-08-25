package common

import (
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v3"
)

type GinComponent interface {
	GetPort() int
	GetRouter() *gin.Engine
}

type ScyllaComponent interface {
	GetCluster() *gocql.ClusterConfig
	CreateSession() (*gocql.Session, error)
	CreateSessionWithGoCqlX() (*gocqlx.Session, error)
}
