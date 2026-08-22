package scylladbc

import (
	"testing"
	"time"
)

func TestScyllaDbComponentLifecycle(t *testing.T) {
	comp := NewScyllaDbComponent("scylla")
	if comp.ID() != "scylla" {
		t.Fatalf("expected ID 'scylla', got '%s'", comp.ID())
	}

	comp.InitFlags()

	if comp.config.timeout != 10*time.Second {
		t.Fatalf("expected default timeout 10s, got %v", comp.config.timeout)
	}

	if err := comp.Stop(); err != nil {
		t.Fatalf("Stop error: %v", err)
	}
}

func TestScyllaDbComponentActivateValidation(t *testing.T) {
	t.Run("Empty keyspace", func(t *testing.T) {
		comp := NewScyllaDbComponent("scylla")
		comp.hostsStr = "localhost:9042"
		comp.config.ks = ""
		if err := comp.Activate(nil); err == nil {
			t.Fatal("expected error for empty keyspace")
		}
	})

	t.Run("Empty hosts", func(t *testing.T) {
		comp := NewScyllaDbComponent("scylla")
		comp.hostsStr = ""
		comp.config.ks = "my_keyspace"
		if err := comp.Activate(nil); err == nil {
			t.Fatal("expected error for empty hosts")
		}
	})
}
