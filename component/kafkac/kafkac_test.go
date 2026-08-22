package kafkac

import (
	"testing"
	"time"
)

func TestKafkaComponentLifecycle(t *testing.T) {
	comp := NewKafkaComponent("kafka")
	if comp.ID() != "kafka" {
		t.Fatalf("expected ID 'kafka', got '%s'", comp.ID())
	}

	comp.InitFlags()

	if comp.config.AddrsStr != "localhost:9092" {
		t.Fatalf("expected default AddrsStr 'localhost:9092', got '%s'", comp.config.AddrsStr)
	}
	if comp.config.maxRetries != 3 {
		t.Fatalf("expected maxRetries 3, got %d", comp.config.maxRetries)
	}
	if comp.config.maxWaitTime != 10*time.Second {
		t.Fatalf("expected maxWaitTime 10s, got %v", comp.config.maxWaitTime)
	}

	if comp.GetProducer() != nil {
		t.Fatal("expected nil producer before activate")
	}

	if err := comp.Stop(); err != nil {
		t.Fatalf("Stop error: %v", err)
	}
}
