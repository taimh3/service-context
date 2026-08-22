package otelc

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	sctx "github.com/taimaifika/service-context"
)

func TestOtelComponentResource(t *testing.T) {
	oc := NewOtel("otel")

	sv := sctx.NewServiceContext(
		sctx.WithName("service-context-otelcomp"),
		sctx.WithComponent(oc),
	)

	oc.serviceName = "service-context-otelcomp"
	oc.serviceVersion = "0.0.1"
	oc.environment = "development"
	oc.exporterOtlpEndpoint = "http://localhost:4318"
	oc.exporterOtlpProtocol = "http"
	oc.isEnabled = true
	oc.isEnabledTrace = true
	oc.isEnabledLog = true

	res := oc.newResource()
	if res == nil {
		t.Fatal("resource is nil")
	}

	t.Logf("Resource SchemaURL: %s", res.SchemaURL())
	for iter := res.Iter(); iter.Next(); {
		attr := iter.Attribute()
		t.Logf("Resource Attribute: %s = %s", attr.Key, attr.Value.AsString())
	}

	err := oc.Activate(sv)
	if err != nil {
		t.Fatalf("Activate error: %v", err)
	}

	tp := otel.GetTracerProvider()
	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")
	span.End()

	t.Logf("Tracer provider initialized successfully, trace ID: %s", span.SpanContext().TraceID())
	_ = ctx
	_ = oc.Stop()
}
