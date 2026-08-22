package redisc

import (
	"testing"
)

func TestNewRedisComponent(t *testing.T) {
	rc := NewRedisComponent("redis",
		WithURL("localhost:6379"),
		WithDisableIdentity(true),
		WithDisableMaintNotifications(true),
	)

	if rc.id != "redis" {
		t.Errorf("expected id 'redis', got '%s'", rc.id)
	}

	if rc.url != "localhost:6379" {
		t.Errorf("expected url 'localhost:6379', got '%s'", rc.url)
	}

	if !rc.disableIdentity {
		t.Errorf("expected disableIdentity to be true")
	}

	if !rc.disableMaintNotifications {
		t.Errorf("expected disableMaintNotifications to be true")
	}
}
