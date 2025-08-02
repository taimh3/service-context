package common

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GINComponent interface {
	GetPort() int
	GetRouter() *gin.Engine
}

type GormComponent interface {
	GetDB() *gorm.DB
}
