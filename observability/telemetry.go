package observability

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Config for telemetry
type Config struct {
	TracerProvider trace.TracerProvider
	MeterProvider  metric.MeterProvider
	ServiceName    string
	ServiceVersion string
}

// Telemetry holds tracing and metrics
type Telemetry struct {
	tracer trace.Tracer
	meter  metric.Meter
	config *Config

	// Metrics
	requestCounter    metric.Int64Counter
	requestDuration   metric.Float64Histogram
	tokenCounter      metric.Int64Counter
	errorCounter      metric.Int64Counter
	streamEventCounter metric.Int64Counter
}

// New creates a new Telemetry instance
func New(config *Config) (*Telemetry, error) {
	if config == nil {
		config = &Config{
			ServiceName:    "llmx",
			ServiceVersion: "1.0.0",
		}
	}

	var tracer trace.Tracer
	if config.TracerProvider != nil {
		tracer = config.TracerProvider.Tracer(
			config.ServiceName,
			trace.WithInstrumentationVersion(config.ServiceVersion),
		)
	}

	var meter metric.Meter
	if config.MeterProvider != nil {
		meter = config.MeterProvider.Meter(
			config.ServiceName,
			metric.WithInstrumentationVersion(config.ServiceVersion),
		)
	}

	tel := &Telemetry{
		tracer: tracer,
		meter:  meter,
		config: config,
	}

	// Initialize metrics
	if meter != nil {
		var err error

		tel.requestCounter, err = meter.Int64Counter(
			"llmx.requests.total",
			metric.WithDescription("Total number of LLM requests"),
			metric.WithUnit("{request}"),
		)
		if err != nil {
			return nil, err
		}

		tel.requestDuration, err = meter.Float64Histogram(
			"llmx.request.duration",
			metric.WithDescription("Duration of LLM requests"),
			metric.WithUnit("ms"),
		)
		if err != nil {
			return nil, err
		}

		tel.tokenCounter, err = meter.Int64Counter(
			"llmx.tokens.total",
			metric.WithDescription("Total number of tokens used"),
			metric.WithUnit("{token}"),
		)
		if err != nil {
			return nil, err
		}

		tel.errorCounter, err = meter.Int64Counter(
			"llmx.errors.total",
			metric.WithDescription("Total number of errors"),
			metric.WithUnit("{error}"),
		)
		if err != nil {
			return nil, err
		}

		tel.streamEventCounter, err = meter.Int64Counter(
			"llmx.stream.events.total",
			metric.WithDescription("Total number of stream events"),
			metric.WithUnit("{event}"),
		)
		if err != nil {
			return nil, err
		}
	}

	return tel, nil
}

// StartSpan starts a new tracing span
func (t *Telemetry) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if t.tracer != nil {
		return t.tracer.Start(ctx, name, opts...)
	}
	return ctx, trace.SpanFromContext(ctx)
}

// RecordRequest records a request metric
func (t *Telemetry) RecordRequest(ctx context.Context, provider, model string, success bool) {
	if t.requestCounter == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("provider", provider),
		attribute.String("model", model),
		attribute.Bool("success", success),
	}

	t.requestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordDuration records request duration
func (t *Telemetry) RecordDuration(ctx context.Context, provider, model string, durationMs float64) {
	if t.requestDuration == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("provider", provider),
		attribute.String("model", model),
	}

	t.requestDuration.Record(ctx, durationMs, metric.WithAttributes(attrs...))
}

// RecordTokens records token usage
func (t *Telemetry) RecordTokens(ctx context.Context, provider, model, tokenType string, count int64) {
	if t.tokenCounter == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("provider", provider),
		attribute.String("model", model),
		attribute.String("type", tokenType), // "prompt", "completion", "total"
	}

	t.tokenCounter.Add(ctx, count, metric.WithAttributes(attrs...))
}

// RecordError records an error
func (t *Telemetry) RecordError(ctx context.Context, provider, model, errorType string) {
	if t.errorCounter == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("provider", provider),
		attribute.String("model", model),
		attribute.String("error_type", errorType),
	}

	t.errorCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordStreamEvent records a stream event
func (t *Telemetry) RecordStreamEvent(ctx context.Context, provider, model, eventType string) {
	if t.streamEventCounter == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("provider", provider),
		attribute.String("model", model),
		attribute.String("event_type", eventType),
	}

	t.streamEventCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}
