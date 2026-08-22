package gormc

import (
	"testing"
)

type sampleModel struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:255"`
}

func TestGormDBTypeParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected GormDBType
	}{
		{"mysql", GormDBTypeMySQL},
		{"MySQL", GormDBTypeMySQL},
		{"postgres", GormDBTypePostgres},
		{"POSTGRES", GormDBTypePostgres},
		{"sqlite", GormDBTypeSQLite},
		{"SQLite", GormDBTypeSQLite},
		{"mssql", GormDBTypeMSSQL},
		{"unknown", GormDBTypeNotSupported},
	}

	for _, tc := range tests {
		got := getDBType(tc.input)
		if got != tc.expected {
			t.Errorf("getDBType(%s): expected %v, got %v", tc.input, tc.expected, got)
		}
	}
}

func TestGormComponentLifecycleWithSQLite(t *testing.T) {
	dbComp := NewGormDB("db", "test")
	if dbComp.ID() != "db" {
		t.Fatalf("expected ID 'db', got '%s'", dbComp.ID())
	}

	dbComp.InitFlags()

	dbComp.GormOpt.dbType = "sqlite"
	dbComp.GormOpt.dsn = "file::memory:?cache=shared"
	dbComp.GormOpt.logLevel = "debug"
	dbComp.GormOpt.isPluginOpenTelemetry = false

	if err := dbComp.Activate(nil); err != nil {
		t.Fatalf("Activate error: %v", err)
	}

	db := dbComp.GetDB()
	if db == nil {
		t.Fatal("expected non-nil *gorm.DB")
	}

	// Auto-migrate test table
	if err := db.AutoMigrate(&sampleModel{}); err != nil {
		t.Fatalf("AutoMigrate error: %v", err)
	}

	// Create record
	item := sampleModel{Name: "test-item"}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("Create error: %v", err)
	}

	// Read record
	var found sampleModel
	if err := db.First(&found, item.ID).Error; err != nil {
		t.Fatalf("First error: %v", err)
	}
	if found.Name != "test-item" {
		t.Fatalf("expected name 'test-item', got '%s'", found.Name)
	}

	if err := dbComp.Stop(); err != nil {
		t.Fatalf("Stop error: %v", err)
	}
}

func TestGormComponentUnsupportedDB(t *testing.T) {
	dbComp := NewGormDB("db", "")
	dbComp.GormOpt.dbType = "unsupported-driver"
	if err := dbComp.Activate(nil); err == nil {
		t.Fatal("expected error for unsupported db type")
	}
}
