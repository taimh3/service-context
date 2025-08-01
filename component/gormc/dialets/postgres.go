package dialets

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresDB Get Postgres DB connection
// dns string
// Ex: host=myhost port=myport user=gorm dbname=gorm password=mypassword
func PostgresDB(dsn string, gormConfig *gorm.Config) (db *gorm.DB, err error) {
	return gorm.Open(postgres.Open(dsn), gormConfig)
}
