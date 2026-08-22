package mongodbc

import (
	"testing"
	"time"
)

func TestMongoDbComponentLifecycle(t *testing.T) {
	comp := NewMongoDbComponent("mongo")
	if comp.ID() != "mongo" {
		t.Fatalf("expected ID 'mongo', got '%s'", comp.ID())
	}

	comp.InitFlags()

	if comp.config.url != "mongodb://localhost:27017" {
		t.Fatalf("expected default URL 'mongodb://localhost:27017', got '%s'", comp.config.url)
	}
	if comp.config.timeout != 10*time.Second {
		t.Fatalf("expected default timeout 10s, got %v", comp.config.timeout)
	}

	if err := comp.Stop(); err != nil {
		t.Fatalf("Stop error: %v", err)
	}
}

func TestMongoDbComponentGetters(t *testing.T) {
	comp := NewMongoDbComponent("mongo")
	comp.config.authSource = "admin_db"

	if comp.GetDatabaseName() != "admin_db" {
		t.Fatalf("expected database name 'admin_db', got '%s'", comp.GetDatabaseName())
	}
}
