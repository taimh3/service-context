package redisc

import (
	"testing"
)

func TestNewRedisComponent(t *testing.T) {
	rc := NewRedisComponent("redis",
		WithURL("localhost:6379"),
		WithUsername("myuser"),
		WithPassword("mypass"),
		WithDB(2),
		WithProtocol(2),
		WithIsCluster(false),
		WithDisableIdentity(true),
		WithDisableMaintNotifications(true),
		WithOpenTelemetry(true),
		WithOpenTelemetryTraces(true),
		WithOpenTelemetryMetrics(true),
	)

	if rc.ID() != "redis" {
		t.Errorf("expected id 'redis', got '%s'", rc.ID())
	}

	if rc.url != "localhost:6379" {
		t.Errorf("expected url 'localhost:6379', got '%s'", rc.url)
	}

	if rc.username != "myuser" {
		t.Errorf("expected username 'myuser', got '%s'", rc.username)
	}

	if rc.password != "mypass" {
		t.Errorf("expected password 'mypass', got '%s'", rc.password)
	}

	if rc.db != 2 {
		t.Errorf("expected db 2, got %d", rc.db)
	}

	if rc.protocol != 2 {
		t.Errorf("expected protocol 2, got %d", rc.protocol)
	}

	if rc.isCluster {
		t.Errorf("expected isCluster to be false")
	}

	if !rc.disableIdentity {
		t.Errorf("expected disableIdentity to be true")
	}

	if !rc.disableMaintNotifications {
		t.Errorf("expected disableMaintNotifications to be true")
	}

	if !rc.isOpenTelemetry || !rc.isOpenTelemetryTraces || !rc.isOpenTelemetryMetrics {
		t.Errorf("expected OpenTelemetry flags to be true")
	}

	rc.InitFlags()
}
