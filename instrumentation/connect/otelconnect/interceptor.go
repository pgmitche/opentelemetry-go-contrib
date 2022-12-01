package otelconnect

import (
	"context"
	"github.com/bufbuild/connect-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"net"
	"strings"
)

const (
	// ConnectStatusCodeKey is convention for numeric status code of a connect request.
	ConnectStatusCodeKey = attribute.Key("rpc.connect.status_code")
	// ConnectStatusKey is convention for statuses of a connect request.
	ConnectStatusKey = attribute.Key("rpc.connect.status")
)

var _ connect.Interceptor = &otelInterceptor{}

type otelInterceptor struct {
	cfg    *config
	tracer trace.Tracer
}

func NewOtelInterceptor(opts ...Option) *otelInterceptor {
	cfg := newConfig(opts)
	tracer := cfg.TracerProvider.Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(SemVersion()),
	)

	connect.WithInterceptors()

	return &otelInterceptor{
		cfg:    cfg,
		tracer: tracer,
	}
}

func (i *otelInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		info := &InterceptorInfo{Method: req.Spec().Procedure}
		if i.cfg.Filter != nil && !i.cfg.Filter(info) {
			return next(ctx, req)
		}

		if req.Spec().IsClient {
			name, attrs := buildSpanInfo(req.Spec().Procedure, req.Peer().Addr)
			ctx, span := i.tracer.Start(ctx, name, trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attrs...))
			defer span.End()

			// propagate the span in the outgoing request
			inject(ctx, req, i.cfg.Propagators)
			res, err := next(ctx, req)
			if err != nil {
				// TODO: unpack more error info as attributes?
				// var connectErr *connect.Error
				// if errors.As(err, &connectErr) {
				//
				// }

				code := connect.CodeOf(err)
				span.SetStatus(codes.Error, code.String())
				span.SetAttributes(
					statusCodeAttr(code),
					statusAttr(code),
				)
			}

			return res, err
		} else {
			ctx = extract(ctx, req, i.cfg.Propagators)
			name, attrs := buildSpanInfo(req.Spec().Procedure, req.Peer().Addr)
			ctx, span := i.tracer.Start(
				trace.ContextWithRemoteSpanContext(ctx, trace.SpanContextFromContext(ctx)),
				name,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(attrs...),
			)

			defer span.End()
			res, err := next(ctx, req)
			if err != nil {
				// TODO: unpack more error info as attributes?
				// var connectErr *connect.Error
				// if errors.As(err, &connectErr) {
				//
				// }

				code := connect.CodeOf(err)
				span.SetStatus(codes.Error, code.String())
				span.SetAttributes(
					statusCodeAttr(code),
					statusAttr(code),
				)
			}

			return res, err
		}
	}
}

func (i *otelInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		info := &InterceptorInfo{Method: conn.Spec().Procedure}
		if i.cfg.Filter != nil && !i.cfg.Filter(info) {
			return next(ctx, conn)
		}

		// TODO:

	}
}

func (i *otelInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		info := &InterceptorInfo{Method: spec.Procedure}
		if i.cfg.Filter != nil && !i.cfg.Filter(info) {
			conn := next(ctx, spec)
			return conn
		}

		// TODO:
	}
}

// buildSpanInfo returns a span name and all appropriate attributes from the procedure
// and peer address.
func buildSpanInfo(fullMethod, peerAddress string) (string, []attribute.KeyValue) {
	attrs := []attribute.KeyValue{}
	name, mAttrs := parseFullMethod(fullMethod)
	attrs = append(attrs, mAttrs...)
	attrs = append(attrs, peerAttr(peerAddress)...)
	return name, attrs
}

// parseFullMethod returns a span name following the OpenTelemetry semantic
// conventions as well as all applicable span attribute.KeyValue attributes based
// on a procedure.
func parseFullMethod(fullMethod string) (string, []attribute.KeyValue) {
	name := strings.TrimLeft(fullMethod, "/")
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 {
		// Invalid format, does not follow `/package.service/method`.
		return name, []attribute.KeyValue(nil)
	}

	var attrs []attribute.KeyValue
	if service := parts[0]; service != "" {
		attrs = append(attrs, semconv.RPCServiceKey.String(service))
	}

	if method := parts[1]; method != "" {
		attrs = append(attrs, semconv.RPCMethodKey.String(method))
	}

	return name, attrs
}

// peerAttr returns attributes about the peer address.
func peerAttr(addr string) []attribute.KeyValue {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return []attribute.KeyValue(nil)
	}

	if host == "" {
		host = "127.0.0.1"
	}

	return []attribute.KeyValue{
		semconv.NetPeerIPKey.String(host),
		semconv.NetPeerPortKey.String(port),
	}
}

// statusCodeAttr returns status code attribute of connect error code
func statusCodeAttr(c connect.Code) attribute.KeyValue {
	return ConnectStatusCodeKey.Int(int(c))
}

// statusAttr returns status attribute of connect error
func statusAttr(c connect.Code) attribute.KeyValue {
	return ConnectStatusKey.String(c.String())
}

func extract(ctx context.Context, req connect.AnyRequest, propagators propagation.TextMapPropagator) context.Context {
	return propagators.Extract(ctx, propagation.HeaderCarrier(req.Header()))
}

func inject(ctx context.Context, req connect.AnyRequest, propagators propagation.TextMapPropagator) {
	propagators.Inject(ctx, propagation.HeaderCarrier(req.Header()))
}
