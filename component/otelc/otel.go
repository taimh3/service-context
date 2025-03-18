package otelc

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	sctx "github.com/taimaifika/service-context"
)

// Default values for configuration.
const (
	OtelProtocolHTTP = "http"
	OtelProtocolGRPC = "grpc"

	OtelPrintToConsole = "console"

	defaultOtelEndpointHttp = "http://localhost:4318"
	defaultOtelEndpointGrpc = "http://localhost:4317"
	defaultNameService      = ""
	defaultVersion          = ""
	defaultOtelProtocol     = OtelProtocolGRPC
	defaultIsEnabled        = true
	defaultPrefix           = "otel"
)

type config struct {
	// on/off switch
	isEnabled bool

	// otel attributes
	serviceName    string
	serviceVersion string

	// otel exporter
	exporterOtlpEndpoint string
	exporterOtlpProtocol string

	// otel features
	isEnabledTrace  bool
	isEnabledMetric bool
	isEnabledLog    bool
}

type otelComponent struct {
	*config
	id     string
	prefix string

	ctx context.Context

	shutdown func(context.Context) error
}

func NewOtel(id string) *otelComponent {
	return &otelComponent{
		config: new(config),
		id:     id,
		ctx:    context.Background(),
		prefix: defaultPrefix,
	}
}

func (oc *otelComponent) ID() string {
	return oc.id
}

func (oc *otelComponent) InitFlags() {

	flag.BoolVar(&oc.isEnabled, oc.prefix+"-is-enabled", defaultIsEnabled, "Enable otel service")

	// otel attributes
	// OTEL_SERVICE_NAME
	flag.StringVar(&oc.serviceName, oc.prefix+"-service-name", defaultNameService, "The service name must be the same APP_NAME in .env")
	// OTEL_SERVICE_VERSION
	flag.StringVar(&oc.serviceVersion, oc.prefix+"-service-version", defaultVersion, "The service version must be the same release, e.g. 1.0.0")

	// otel exporter
	// OTEL_EXPORTER_OTLP_PROTOCOL
	flag.StringVar(&oc.exporterOtlpProtocol, oc.prefix+"-exporter-otlp-protocol", defaultOtelProtocol, "Otel protocol, e.g. http or grpc")
	// OTEL_EXPORTER_OTLP_ENDPOINT
	flag.StringVar(&oc.exporterOtlpEndpoint, oc.prefix+"-exporter-otlp-endpoint", "", "Otel otlp endpoint, e.g. http://localhost:4317")

	// otel features
	flag.BoolVar(&oc.isEnabledTrace, oc.prefix+"-is-enabled-trace", true, "Enable otel trace")
	flag.BoolVar(&oc.isEnabledMetric, oc.prefix+"-is-enabled-metric", true, "Enable otel metric")
	flag.BoolVar(&oc.isEnabledLog, oc.prefix+"-is-enabled-log", true, "Enable otel log")
}

func (oc *otelComponent) Activate(sv sctx.ServiceContext) error {
	// otel is not enabled
	if !oc.isEnabled {
		return nil
	}

	// load config
	if err := oc.Configure(); err != nil {
		return err
	}

	// setup otel sdk
	shutdown, err := oc.setupOTelSdk()
	if err != nil {
		return err
	}
	oc.shutdown = shutdown

	return nil
}

func (oc *otelComponent) Stop() error {
	oc.shutdown(oc.ctx)
	return nil
}

// Configure configures the service.
func (oc *otelComponent) Configure() error {
	// Check if the servicename is empty
	if oc.serviceName == "" {
		return errors.New("otel service name is empty")
	}

	// Check if the serviceVersion is empty
	if oc.serviceVersion == "" {
		return errors.New("otel service version is empty")
	}

	// Check if the exporterOtlpEndpoint is empty
	if oc.exporterOtlpEndpoint == "" {
		if oc.exporterOtlpProtocol == OtelProtocolGRPC {
			oc.exporterOtlpEndpoint = defaultOtelEndpointGrpc
		} else {
			oc.exporterOtlpEndpoint = defaultOtelEndpointHttp
		}
	}

	return nil
}

// setupOTelSdk bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func (oc *otelComponent) setupOTelSdk() (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(oc.ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	if oc.isEnabledTrace {
		tracerProvider, providerErr := oc.newTraceProvider()
		if providerErr != nil {
			handleErr(providerErr)
			return
		}
		shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
		otel.SetTracerProvider(tracerProvider)
	}

	// Set up meter provider.
	if oc.isEnabledMetric {
		meterProvider, providerErr := oc.newMeterProvider()
		if providerErr != nil {
			handleErr(providerErr)
			return
		}
		shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
		otel.SetMeterProvider(meterProvider)
	}

	// Set up logger provider.
	if oc.isEnabledLog {
		loggerProvider, providerErr := oc.newLoggerProvider()
		if providerErr != nil {
			handleErr(providerErr)
			return
		}
		shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
		global.SetLoggerProvider(loggerProvider)
	}
	return
}

// newPropagator creates a new propagator.
func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// newTraceProvider creates a new trace provider.
func (oc *otelComponent) newTraceProvider() (*trace.TracerProvider, error) {
	var traceExporter trace.SpanExporter

	if oc.isOtlpProtocolEnabled() {
		// Exporter to otlp
		otlpTraceExporter, err := oc.newOtlpTraceExporter()
		if err != nil {
			return nil, err
		}
		traceExporter = otlpTraceExporter
	} else {
		// Exporter to stdout
		stdoutTraceExporter, err := stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			return nil, err
		}
		traceExporter = stdoutTraceExporter
	}

	// Resource attributes
	res := oc.newResource()

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
		trace.WithResource(res),
	)
	return traceProvider, nil
}

// newOtlpTraceExporter creates a new OTLP trace exporter. (gRPC or HTTP)
func (oc *otelComponent) newOtlpTraceExporter() (trace.SpanExporter, error) {
	if oc.exporterOtlpProtocol == OtelProtocolHTTP {
		return otlptracehttp.New(oc.ctx)
	}
	return otlptracegrpc.New(oc.ctx)
}

// newResource creates a new resource with service.name and service.namespace.
func (oc *otelComponent) newResource() *resource.Resource {
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(oc.serviceName),
		semconv.ServiceVersionKey.String(oc.serviceVersion),
	)
	return res
}

// newMeterProvider creates a new meter provider.
func (oc *otelComponent) newMeterProvider() (*metric.MeterProvider, error) {
	var metricExporter metric.Exporter
	if oc.isOtlpProtocolEnabled() {
		// Exporter to otlp
		otlpMetricExporter, err := oc.newOtlpMetricExporter()
		if err != nil {
			return nil, err
		}
		metricExporter = otlpMetricExporter
	} else {
		// Exporter to stdout
		stdoutMetricExporter, err := stdoutmetric.New()
		if err != nil {
			return nil, err
		}
		metricExporter = stdoutMetricExporter
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			// Default is 1m. Set to 3s for demonstrative purposes.
			metric.WithInterval(3*time.Second))),
	)
	return meterProvider, nil
}

// newOtlpMetricExporter creates a new OTLP metric exporter. (gRPC or HTTP)
func (oc *otelComponent) newOtlpMetricExporter() (metric.Exporter, error) {
	if oc.exporterOtlpProtocol == OtelProtocolHTTP {
		return otlpmetrichttp.New(oc.ctx)
	}
	return otlpmetricgrpc.New(oc.ctx)
}

// newLoggerProvider creates a new logger provider.
func (oc *otelComponent) newLoggerProvider() (*log.LoggerProvider, error) {
	var logExporter log.Exporter
	if oc.isOtlpProtocolEnabled() {
		// Exporter to otlp
		otlpLogExporter, err := oc.newOtlpLogExporter()
		if err != nil {
			return nil, err
		}
		logExporter = otlpLogExporter

	} else {
		// Exporter to stdout
		stdoutLogExporter, err := stdoutlog.New()
		if err != nil {
			return nil, err
		}
		logExporter = stdoutLogExporter
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
	)

	// set default slog
	if oc.isOtlpProtocolEnabled() {
		slog.Info("Using OTLP log exporter")
		slog.SetDefault(slog.New(otelslog.NewHandler(oc.serviceName, otelslog.WithLoggerProvider(loggerProvider))))
	}

	return loggerProvider, nil
}

// newOtlpLogExporter creates a new OTLP log exporter. (gRPC or HTTP)
func (oc *otelComponent) newOtlpLogExporter() (log.Exporter, error) {
	if oc.exporterOtlpProtocol == OtelProtocolHTTP {
		return otlploghttp.New(oc.ctx)
	}
	return otlploggrpc.New(oc.ctx)
}

// IsOtlpProtocolEnabled returns true if the otlp protocol is enabled.
func (oc *otelComponent) isOtlpProtocolEnabled() bool {
	return oc.exporterOtlpEndpoint != OtelPrintToConsole
}
