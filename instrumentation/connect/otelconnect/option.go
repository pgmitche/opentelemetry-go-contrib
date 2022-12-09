package otelconnect

import (
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Option applies an option value for a config.
type Option interface {
	apply(*config)
}

type propagatorsOption struct{ p propagation.TextMapPropagator }

func (o propagatorsOption) apply(c *config) {
	if o.p != nil {
		c.Propagators = o.p
	}
}

// WithPropagators returns an Option to use the Propagators when extracting
// and injecting trace context from requests.
func WithPropagators(p propagation.TextMapPropagator) Option {
	return propagatorsOption{p: p}
}

type tracerProviderOption struct{ tp trace.TracerProvider }

func (o tracerProviderOption) apply(c *config) {
	if o.tp != nil {
		c.TracerProvider = o.tp
	}
}

// WithInterceptorFilter returns an Option to use the request filter.
func WithInterceptorFilter(f Filter) Option {
	return interceptorFilterOption{f: f}
}

type interceptorFilterOption struct {
	f Filter
}

func (o interceptorFilterOption) apply(c *config) {
	if o.f != nil {
		c.Filter = o.f
	}
}

// WithTracerProvider returns an Option to use the TracerProvider when
// creating a Tracer.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return tracerProviderOption{tp: tp}
}
