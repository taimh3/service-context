package ginc

import (
	"testing"

	sctx "github.com/taimaifika/service-context"
)

func TestGinComponentLifecycle(t *testing.T) {
	g := NewGin("gin")
	if g.ID() != "gin" {
		t.Fatalf("expected ID 'gin', got '%s'", g.ID())
	}

	g.InitFlags()

	sv := sctx.NewServiceContext(sctx.WithName("test-service"))

	if err := g.Activate(sv); err != nil {
		t.Fatalf("Activate error: %v", err)
	}

	if g.GetRouter() == nil {
		t.Fatal("expected non-nil gin router")
	}

	if g.GetPort() != 3000 {
		t.Fatalf("expected default port 3000, got %d", g.GetPort())
	}

	if err := g.Stop(); err != nil {
		t.Fatalf("Stop error: %v", err)
	}
}

func TestGinComponentReleaseMode(t *testing.T) {
	g := NewGin("gin")
	g.Config.ginMode = "release"
	sv := sctx.NewServiceContext(sctx.WithName("test-service"))

	if err := g.Activate(sv); err != nil {
		t.Fatalf("Activate error: %v", err)
	}
}
