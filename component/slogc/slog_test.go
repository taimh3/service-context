package slogc

import (
	"log/slog"
	"testing"
)

func TestSlogComponentLifecycle(t *testing.T) {
	comp := NewSlogComponent()
	if comp.ID() != "slog" {
		t.Fatalf("expected ID 'slog', got '%s'", comp.ID())
	}

	comp.InitFlags()

	comp.SetLogLevel("debug")
	if comp.opts.Level != slog.LevelDebug {
		t.Fatalf("expected LevelDebug, got %v", comp.opts.Level)
	}

	comp.SetLogLevel("info")
	if comp.opts.Level != slog.LevelInfo {
		t.Fatalf("expected LevelInfo, got %v", comp.opts.Level)
	}

	comp.SetLogLevel("warn")
	if comp.opts.Level != slog.LevelWarn {
		t.Fatalf("expected LevelWarn, got %v", comp.opts.Level)
	}

	comp.SetLogLevel("error")
	if comp.opts.Level != slog.LevelError {
		t.Fatalf("expected LevelError, got %v", comp.opts.Level)
	}

	comp.SetLogFormat("json")
	if comp.handler == nil {
		t.Fatal("expected non-nil handler after SetLogFormat(json)")
	}

	comp.SetLogFormat("text")
	if comp.handler == nil {
		t.Fatal("expected non-nil handler after SetLogFormat(text)")
	}

	if err := comp.Activate(nil); err != nil {
		t.Fatalf("Activate error: %v", err)
	}

	if err := comp.Stop(); err != nil {
		t.Fatalf("Stop error: %v", err)
	}
}
