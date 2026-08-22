package sctx

import (
	"testing"
)

type dummyComponent struct {
	id string
}

func (d *dummyComponent) ID() string {
	return d.id
}

func (d *dummyComponent) InitFlags() {}

func (d *dummyComponent) Activate(ServiceContext) error {
	return nil
}

func (d *dummyComponent) Stop() error {
	return nil
}

func TestServiceContextLifecycle(t *testing.T) {
	comp := &dummyComponent{id: "dummy"}

	sv := NewServiceContext(
		WithName("test-app"),
		WithComponent(comp),
	)

	if sv.GetName() != "test-app" {
		t.Fatalf("expected name 'test-app', got '%s'", sv.GetName())
	}

	if sv.EnvName() != DevEnv {
		t.Fatalf("expected default env '%s', got '%s'", DevEnv, sv.EnvName())
	}

	if err := sv.Load(); err != nil {
		t.Fatalf("Load error: %v", err)
	}

	gotComp, ok := sv.Get("dummy")
	if !ok || gotComp == nil {
		t.Fatal("expected to find dummy component")
	}

	mustComp := sv.MustGet("dummy")
	if mustComp == nil {
		t.Fatal("expected non-nil from MustGet")
	}

	if err := sv.Stop(); err != nil {
		t.Fatalf("Stop error: %v", err)
	}
}
