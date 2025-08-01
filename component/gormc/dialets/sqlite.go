package dialets

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SQLiteDB Get SQLite DB connection
// dsn string
// Ex: /tmp/gorm.db
func SQLiteDB(dsn string, gormConfig *gorm.Config) (db *gorm.DB, err error) {
	return gorm.Open(sqlite.Open(dsn), gormConfig)
}
