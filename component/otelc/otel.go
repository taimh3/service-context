package otelc

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"strings"
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
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

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
	defaultVersion          = "1.0.0"
	defaultOtelProtocol     = OtelProtocolGRPC
	defaultIsEnabled        = true
	defaultPrefix           = "otel"
	defaultEnvironment      = "development"
	defaultLanguage         = "go"
)

type config struct {
	// on/off switch
	isEnabled bool

	// otel attributes
	serviceName    string
	serviceVersion string
	environment    string

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

type Option func(*otelComponent)

func WithServiceName(name string) Option {
	return func(oc *otelComponent) {
		oc.serviceName = name
	}
}

func WithServiceVersion(version string) Option {
	return func(oc *otelComponent) {
		oc.serviceVersion = version
	}
}

func WithEnvironment(env string) Option {
	return func(oc *otelComponent) {
		oc.environment = env
	}
}

func WithEndpoint(endpoint string) Option {
	return func(oc *otelComponent) {
		oc.exporterOtlpEndpoint = endpoint
	}
}

func WithProtocol(protocol string) Option {
	return func(oc *otelComponent) {
		oc.exporterOtlpProtocol = protocol
	}
}

func WithPrefix(prefix string) Option {
	return func(oc *otelComponent) {
		oc.prefix = prefix
	}
}

func NewOtel(id string, opts ...Option) *otelComponent {
	oc := &otelComponent{
		config: new(config),
		id:     id,
		ctx:    context.Background(),
		prefix: defaultPrefix,
	}
	for _, opt := range opts {
		opt(oc)
	}
	return oc
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
	// OTEL_ENVIRONMENT
	flag.StringVar(&oc.environment, oc.prefix+"-environment", defaultEnvironment, "The environment name, e.g. development, staging, and production")

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

	// If serviceName is not set from flag/env, fallback to ServiceContext name
	if oc.serviceName == "" && sv != nil && sv.GetName() != "" {
		oc.serviceName = sv.GetName()
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
		oc.serviceVersion = defaultVersion
	}

	// Check if the exporterOtlpEndpoint is empty
	if oc.exporterOtlpEndpoint == "" {
		if oc.exporterOtlpProtocol == OtelProtocolGRPC {
			oc.exporterOtlpEndpoint = defaultOtelEndpointGrpc
		} else {
			oc.exporterOtlpEndpoint = defaultOtelEndpointHttp
		}
	} else if oc.exporterOtlpEndpoint != OtelPrintToConsole && !strings.HasPrefix(oc.exporterOtlpEndpoint, "http://") && !strings.HasPrefix(oc.exporterOtlpEndpoint, "https://") {
		oc.exporterOtlpEndpoint = "http://" + oc.exporterOtlpEndpoint
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

		if oc.isOtlpProtocolEnabled() {
			slog.Info("Using OTLP log exporter")
			slog.SetDefault(slog.New(otelslog.NewHandler(oc.serviceName, otelslog.WithLoggerProvider(loggerProvider))))
		}
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
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
	)
	return traceProvider, nil
}

func (oc *otelComponent) getOtlpHTTPEndpointOptions() (endpoint string, insecure bool) {
	ep := oc.exporterOtlpEndpoint
	if strings.HasPrefix(ep, "https://") {
		return strings.TrimSuffix(strings.TrimPrefix(ep, "https://"), "/"), false
	}
	if strings.HasPrefix(ep, "http://") {
		return strings.TrimSuffix(strings.TrimPrefix(ep, "http://"), "/"), true
	}
	return strings.TrimSuffix(ep, "/"), strings.Contains(ep, "localhost") || strings.Contains(ep, "127.0.0.1")
}

func (oc *otelComponent) getOtlpGRPCEndpointOptions() (endpoint string, insecure bool) {
	ep := oc.exporterOtlpEndpoint
	if strings.HasPrefix(ep, "https://") {
		return strings.TrimSuffix(strings.TrimPrefix(ep, "https://"), "/"), false
	}
	if strings.HasPrefix(ep, "http://") {
		return strings.TrimSuffix(strings.TrimPrefix(ep, "http://"), "/"), true
	}
	return strings.TrimSuffix(ep, "/"), true
}

// newOtlpTraceExporter creates a new OTLP trace exporter. (gRPC or HTTP)
func (oc *otelComponent) newOtlpTraceExporter() (trace.SpanExporter, error) {
	if oc.exporterOtlpProtocol == OtelProtocolHTTP {
		host, isInsecure := oc.getOtlpHTTPEndpointOptions()
		opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(host)}
		if isInsecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		return otlptracehttp.New(oc.ctx, opts...)
	}

	host, isInsecure := oc.getOtlpGRPCEndpointOptions()
	opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(host)}
	if isInsecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	return otlptracegrpc.New(oc.ctx, opts...)
}

// newResource creates a new resource with service.name and service.namespace.
func (oc *otelComponent) newResource() *resource.Resource {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceNameKey.String(oc.serviceName),
			semconv.ServiceVersionKey.String(oc.serviceVersion),
			semconv.DeploymentEnvironmentKey.String(oc.environment),
			semconv.TelemetrySDKLanguageGo.Key.String(defaultLanguage),
			semconv.TelemetrySDKVersionKey.String(otel.Version()),
		),
	)
	if err != nil {
		slog.Error("failed to merge resource attributes", slog.Any("error", err))
		return resource.NewSchemaless(
			semconv.ServiceNameKey.String(oc.serviceName),
			semconv.ServiceVersionKey.String(oc.serviceVersion),
			semconv.DeploymentEnvironmentKey.String(oc.environment),
		)
	}
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

	res := oc.newResource()

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			// Default is 1m. Set to 3s for demonstrative purposes.
			metric.WithInterval(3*time.Second))),
		metric.WithResource(res),
	)
	return meterProvider, nil
}

// newOtlpMetricExporter creates a new OTLP metric exporter. (gRPC or HTTP)
func (oc *otelComponent) newOtlpMetricExporter() (metric.Exporter, error) {
	if oc.exporterOtlpProtocol == OtelProtocolHTTP {
		host, isInsecure := oc.getOtlpHTTPEndpointOptions()
		opts := []otlpmetrichttp.Option{otlpmetrichttp.WithEndpoint(host)}
		if isInsecure {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}
		return otlpmetrichttp.New(oc.ctx, opts...)
	}

	host, isInsecure := oc.getOtlpGRPCEndpointOptions()
	opts := []otlpmetricgrpc.Option{otlpmetricgrpc.WithEndpoint(host)}
	if isInsecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}
	return otlpmetricgrpc.New(oc.ctx, opts...)
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

	res := oc.newResource()

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
		log.WithResource(res),
	)

	return loggerProvider, nil
}

// newOtlpLogExporter creates a new OTLP log exporter. (gRPC or HTTP)
func (oc *otelComponent) newOtlpLogExporter() (log.Exporter, error) {
	if oc.exporterOtlpProtocol == OtelProtocolHTTP {
		host, isInsecure := oc.getOtlpHTTPEndpointOptions()
		opts := []otlploghttp.Option{otlploghttp.WithEndpoint(host)}
		if isInsecure {
			opts = append(opts, otlploghttp.WithInsecure())
		}
		return otlploghttp.New(oc.ctx, opts...)
	}

	host, isInsecure := oc.getOtlpGRPCEndpointOptions()
	opts := []otlploggrpc.Option{otlploggrpc.WithEndpoint(host)}
	if isInsecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}
	return otlploggrpc.New(oc.ctx, opts...)
}

// IsOtlpProtocolEnabled returns true if the otlp protocol is enabled.
func (oc *otelComponent) isOtlpProtocolEnabled() bool {
	return oc.exporterOtlpEndpoint != OtelPrintToConsole
}
