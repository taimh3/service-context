package common

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RedisComponent interface for redis component
type RedisComponent interface {
	GetRedis() *redis.ClusterClient
}

// GINComponent interface for gin component
type GINComponent interface {
	GetPort() int
	GetRouter() *gin.Engine
}
