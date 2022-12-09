package otelconnect

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	// instrumentationName is the name of this instrumentation package.
	instrumentationName = "go.opentelemetry.io/contrib/instrumentation/connect/otelconnect"
)

// Filter is a predicate used to determine whether a given request in
// interceptor info should be traced. A Filter must return true if
// the request should be traced.
type Filter func(*InterceptorInfo) bool

// config is a group of options for this instrumentation.
type config struct {
	Filter         Filter
	Propagators    propagation.TextMapPropagator
	TracerProvider trace.TracerProvider
	Tracer         trace.Tracer

	meter             metric.Meter
	rpcServerDuration syncint64.Histogram
}

// newConfig returns a config configured with all the passed Options.
func newConfig(opts []Option) *config {
	c := &config{
		Propagators: otel.GetTextMapPropagator(),
	}

	for _, o := range opts {
		o.apply(c)
	}

	if c.TracerProvider != nil {
		c.Tracer = c.TracerProvider.Tracer(
			instrumentationName,
			trace.WithInstrumentationVersion(SemVersion()),
		)
	}

	return c
}
