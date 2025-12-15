package slogc

import (
	"log/slog"
	"testing"
)

func TestSetLogFormat_UsesArgument(t *testing.T) {
	comp := NewSlogComponent()
	// Manually set internal config to "text" to simulate flag default
	comp.logFormat = "text"

	// Try to set to json via method argument
	comp.SetLogFormat("json")

	// Check if handler is JSONHandler
	_, ok := comp.handler.(*slog.JSONHandler)
	if !ok {
		t.Errorf("Expected JSONHandler, got %T. SetLogFormat should use the argument.", comp.handler)
	}
}

func TestSetLogLevel_Receiver(t *testing.T) {
	comp := NewSlogComponent()
	comp.opts.Level = slog.LevelInfo

	// Calling on *comp. The receiver is now pointer, so it should work regardless.
	comp.SetLogLevel("error")

	if comp.opts.Level != slog.LevelError {
		t.Errorf("Expected LevelError, got %v.", comp.opts.Level)
	}
}

func TestAddSource(t *testing.T) {
	comp := NewSlogComponent()
	comp.addSource = true

	comp.Activate(nil)

	if !comp.opts.AddSource {
		t.Errorf("Expected AddSource to be true")
	}
}
