package gormc

import (
	"flag"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/pkg/errors"
	sctx "github.com/taimaifika/service-context"
	"github.com/taimaifika/service-context/component/gormc/dialets"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
)

type GormDBType int

const (
	GormDBTypeMySQL GormDBType = iota + 1
	GormDBTypePostgres
	GormDBTypeSQLite
	GormDBTypeMSSQL
	GormDBTypeNotSupported
)

type GormOpt struct {
	dsn                   string
	dbType                string
	maxOpenConnections    int
	maxIdleConnections    int
	maxConnectionIdleTime int
	logLevel              string

	isKeepDefaultTransaction bool
	isPrepareStmt            bool

	// Plugin
	// OpenTelemetry tracing plugin
	isPluginOpenTelemetry        bool
	isPluginOpenTelemetryMetrics bool
}

type gormDB struct {
	id     string
	prefix string
	db     *gorm.DB
	*GormOpt
}

func NewGormDB(id, prefix string) *gormDB {
	return &gormDB{
		GormOpt: new(GormOpt),
		id:      id,
		prefix:  strings.TrimSpace(prefix),
	}
}

func (gdb *gormDB) ID() string {
	return gdb.id
}

func (gdb *gormDB) InitFlags() {
	prefix := gdb.prefix
	if gdb.prefix != "" {
		prefix += "-"
	}

	flag.StringVar(
		&gdb.dsn,
		fmt.Sprintf("%sdb-dsn", prefix),
		"",
		"Database dsn",
	)

	flag.StringVar(
		&gdb.dbType,
		fmt.Sprintf("%sdb-driver", prefix),
		"mysql",
		"Database driver (mysql, postgres, sqlite, mssql) - Default mysql",
	)

	flag.IntVar(
		&gdb.maxOpenConnections,
		fmt.Sprintf("%sdb-max-conn", prefix),
		30,
		"maximum number of open connections to the database - Default 30",
	)

	flag.IntVar(
		&gdb.maxIdleConnections,
		fmt.Sprintf("%sdb-max-ide-conn", prefix),
		10,
		"maximum number of database connections in the idle - Default 10",
	)

	flag.IntVar(
		&gdb.maxConnectionIdleTime,
		fmt.Sprintf("%sdb-max-conn-ide-time", prefix),
		3600,
		"maximum amount of time a connection may be idle in seconds - Default 3600",
	)

	flag.StringVar(
		&gdb.logLevel,
		fmt.Sprintf("%sdb-log-level", prefix),
		"info",
		"Log level info | debug | trace - Default info ; debug and trace will log all SQL queries",
	)

	flag.BoolVar(
		&gdb.isKeepDefaultTransaction,
		fmt.Sprintf("%sdb-keep-default-transaction", prefix),
		false,
		"Keep default transaction - Default false",
	)

	flag.BoolVar(
		&gdb.isPrepareStmt,
		fmt.Sprintf("%sdb-prepare-stmt", prefix),
		false,
		"Use prepared statement - Default false",
	)

	flag.BoolVar(
		&gdb.isPluginOpenTelemetry,
		fmt.Sprintf("%sdb-plugin-open-telemetry", prefix),
		false,
		"Enable OpenTelemetry tracing plugin - Default false",
	)

	flag.BoolVar(
		&gdb.isPluginOpenTelemetryMetrics,
		fmt.Sprintf("%sdb-plugin-open-telemetry-metrics", prefix),
		true,
		"Enable OpenTelemetry metrics plugin - Default true",
	)
}

func (gdb *gormDB) Activate(_ sctx.ServiceContext) error {
	dbType := getDBType(gdb.dbType)
	if dbType == GormDBTypeNotSupported {
		return errors.WithStack(errors.New("Database type not supported."))
	}

	slog.Info("Connecting to database...")

	var err error
	gdb.db, err = gdb.getDBConn(dbType)

	if err != nil {
		slog.Error("Cannot connect to database", "error", err.Error())
		return err
	}

	return nil
}

func (gdb *gormDB) Stop() error {
	return nil
}

func (gdb *gormDB) GetDB() *gorm.DB {
	var newSessionDB *gorm.DB
	if gdb.logLevel == "debug" || gdb.logLevel == "trace" {
		newSessionDB = gdb.db.Session(&gorm.Session{NewDB: true}).Debug()
	} else {
		newSessionDB = gdb.db.Session(&gorm.Session{NewDB: true, Logger: gdb.db.Logger.LogMode(logger.Silent)})

		if db, err := newSessionDB.DB(); err == nil {
			db.SetMaxOpenConns(gdb.maxOpenConnections)
			db.SetMaxIdleConns(gdb.maxIdleConnections)
			db.SetConnMaxIdleTime(time.Second * time.Duration(gdb.maxConnectionIdleTime))
		}
	}

	// add plugin
	if gdb.isPluginOpenTelemetry {
		opts := []tracing.Option{}
		if !gdb.isPluginOpenTelemetryMetrics {
			opts = append(opts, tracing.WithoutMetrics())
		}
		newSessionDB.Use(
			tracing.NewPlugin(
				opts...,
			),
		)
	}

	return newSessionDB
}

func getDBType(dbType string) GormDBType {
	switch strings.ToLower(dbType) {
	case "mysql":
		return GormDBTypeMySQL
	case "postgres":
		return GormDBTypePostgres
	case "sqlite":
		return GormDBTypeSQLite
	case "mssql":
		return GormDBTypeMSSQL
	}

	return GormDBTypeNotSupported
}

func (gdb *gormDB) getDBConn(t GormDBType) (dbConn *gorm.DB, err error) {
	// GORM config
	gormConfig := &gorm.Config{
		SkipDefaultTransaction: gdb.isKeepDefaultTransaction,
		PrepareStmt:            gdb.isPrepareStmt,
	}
	switch t {
	case GormDBTypeMySQL:
		return dialets.MySqlDB(gdb.dsn, gormConfig)
	case GormDBTypePostgres:
		return dialets.PostgresDB(gdb.dsn, gormConfig)
	case GormDBTypeSQLite:
		return dialets.SQLiteDB(gdb.dsn, gormConfig)
	case GormDBTypeMSSQL:
		return dialets.MSSqlDB(gdb.dsn, gormConfig)
	}

	return nil, nil
}
