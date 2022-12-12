// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package otelconnect // import "go.opentelemetry.io/contrib/instrumentation/github.com/bufbuild/connect-go/otelgrpc"

import (
	"context"
	"errors"
	"fmt"
	"github.com/bufbuild/connect-go"
	"github.com/stretchr/testify/require"
	v1 "go.opentelemetry.io/contrib/instrumentation/connect/otelconnect/internal/gen/connect/ping/v1"
	"go.opentelemetry.io/contrib/instrumentation/connect/otelconnect/internal/gen/connect/ping/v1/pingv1connect"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

const (
	errorMessage = "RIP"
)

// TODO: add span event assertions
func TestInterceptors(t *testing.T) {
	// Example of any/all propagators
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}, b3.New(), b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)), jaeger.Jaeger{}))

	t.Run("Unary", func(t *testing.T) {
		srvRec := tracetest.NewSpanRecorder()
		clientRec := tracetest.NewSpanRecorder()

		mux := http.NewServeMux()
		mux.Handle(
			pingv1connect.NewPingServiceHandler(
				&pingServer{},
				connect.WithInterceptors(NewOtelInterceptor(
					WithTracerProvider(trace.NewTracerProvider(trace.WithSpanProcessor(srvRec))))),
			),
		)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := pingv1connect.NewPingServiceClient(
			server.Client(),
			server.URL,
			connect.WithInterceptors(NewOtelInterceptor(WithTracerProvider(trace.NewTracerProvider(trace.WithSpanProcessor(clientRec))))),
		)

		_, err := client.Ping(context.Background(), connect.NewRequest(&v1.PingRequest{
			Number: 0,
			Text:   "Ping",
		}))
		require.NoError(t, err)

		outgoingSpan := clientRec.Started()[0]
		incomingSpan := srvRec.Started()[0]
		inspectAndCompareSpans(t,
			incomingSpan,
			[]attribute.KeyValue{
				semconv.RPCMethodKey.String("Ping"),
			},
			outgoingSpan,
			[]attribute.KeyValue{
				semconv.RPCMethodKey.String("Ping"),
			}, false)
	})

	t.Run("Fail", func(t *testing.T) {
		srvRec := tracetest.NewSpanRecorder()
		clientRec := tracetest.NewSpanRecorder()

		mux := http.NewServeMux()
		mux.Handle(
			pingv1connect.NewPingServiceHandler(
				&pingServer{},
				connect.WithInterceptors(NewOtelInterceptor(
					WithTracerProvider(trace.NewTracerProvider(trace.WithSpanProcessor(srvRec))))),
			),
		)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := pingv1connect.NewPingServiceClient(
			server.Client(),
			server.URL,
			connect.WithInterceptors(NewOtelInterceptor(WithTracerProvider(trace.NewTracerProvider(trace.WithSpanProcessor(clientRec))))),
		)

		_, err := client.Fail(context.Background(), connect.NewRequest(&v1.FailRequest{Code: int32(connect.CodeAborted)}))
		require.Error(t, err)

		outgoingSpan := clientRec.Started()[0]
		incomingSpan := srvRec.Started()[0]
		inspectAndCompareSpans(t,
			incomingSpan,
			[]attribute.KeyValue{
				semconv.RPCMethodKey.String("Fail"),
				ConnectStatusCodeKey.Int(int(connect.CodeAborted)),
				ConnectStatusKey.String(connect.CodeAborted.String()),
			},
			outgoingSpan,
			[]attribute.KeyValue{
				semconv.RPCMethodKey.String("Fail"),
				ConnectStatusCodeKey.Int(int(connect.CodeAborted)),
				ConnectStatusKey.String(connect.CodeAborted.String()),
			}, true)
	})

	t.Run("ClientStreaming", func(t *testing.T) {
		srvRec := tracetest.NewSpanRecorder()
		clientRec := tracetest.NewSpanRecorder()

		mux := http.NewServeMux()
		mux.Handle(
			pingv1connect.NewPingServiceHandler(
				&pingServer{},
				connect.WithInterceptors(NewOtelInterceptor(
					WithTracerProvider(trace.NewTracerProvider(trace.WithSpanProcessor(srvRec))))),
			),
		)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := pingv1connect.NewPingServiceClient(
			server.Client(),
			server.URL,
			connect.WithInterceptors(NewOtelInterceptor(WithTracerProvider(trace.NewTracerProvider(trace.WithSpanProcessor(clientRec))))),
		)

		stream := client.Sum(context.Background())
		require.NoError(t, stream.Send(&v1.SumRequest{Number: 1}))
		_, err := stream.CloseAndReceive()
		require.NoError(t, err)

		// Regardless of how many times data is sent/received, we should only have one span
		require.Len(t, clientRec.Started(), 1)
		require.Len(t, srvRec.Started(), 1)

		outgoingSpan := clientRec.Started()[0]
		incomingSpan := srvRec.Started()[0]
		inspectAndCompareSpans(t,
			incomingSpan,
			[]attribute.KeyValue{
				semconv.RPCMethodKey.String("Sum"),
			},
			outgoingSpan,
			[]attribute.KeyValue{
				semconv.RPCMethodKey.String("Sum"),
			}, false)
	})

	t.Run("ServerStreaming", func(t *testing.T) {
		srvRec := tracetest.NewSpanRecorder()
		clientRec := tracetest.NewSpanRecorder()

		mux := http.NewServeMux()
		mux.Handle(
			pingv1connect.NewPingServiceHandler(
				&pingServer{},
				connect.WithInterceptors(NewOtelInterceptor(
					WithTracerProvider(trace.NewTracerProvider(trace.WithSpanProcessor(srvRec))))),
			),
		)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := pingv1connect.NewPingServiceClient(
			server.Client(),
			server.URL,
			connect.WithInterceptors(NewOtelInterceptor(WithTracerProvider(trace.NewTracerProvider(trace.WithSpanProcessor(clientRec))))),
		)

		stream, err := client.CountUp(context.Background(), connect.NewRequest(&v1.CountUpRequest{
			Number: 0,
		}))

		require.NoError(t, err)
		for stream.Receive() {
			// empty the stream
		}

		// Regardless of how many times data is sent/received, we should only have one span
		require.Len(t, clientRec.Started(), 1)
		require.Len(t, srvRec.Started(), 1)

		outgoingSpan := clientRec.Started()[0]
		incomingSpan := srvRec.Started()[0]
		inspectAndCompareSpans(t,
			incomingSpan,
			[]attribute.KeyValue{
				semconv.RPCMethodKey.String("CountUp"),
			},
			outgoingSpan,
			[]attribute.KeyValue{
				semconv.RPCMethodKey.String("CountUp"),
			}, false)
	})

	t.Run("BidiStreaming", func(t *testing.T) {
		srvRec := tracetest.NewSpanRecorder()
		clientRec := tracetest.NewSpanRecorder()

		mux := http.NewServeMux()
		mux.Handle(
			pingv1connect.NewPingServiceHandler(
				&pingServer{},
				connect.WithInterceptors(NewOtelInterceptor(
					WithTracerProvider(trace.NewTracerProvider(trace.WithSpanProcessor(srvRec))))),
			),
		)

		// Bidi requires http/2
		server := httptest.NewUnstartedServer(mux)
		server.EnableHTTP2 = true
		server.StartTLS()
		defer server.Close()

		client := pingv1connect.NewPingServiceClient(
			server.Client(),
			server.URL,
			connect.WithInterceptors(NewOtelInterceptor(WithTracerProvider(trace.NewTracerProvider(trace.WithSpanProcessor(clientRec))))),
		)

		bidi := client.CumSum(context.Background())

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			require.NoError(t, bidi.Send(&v1.CumSumRequest{Number: 1}))
			require.NoError(t, bidi.CloseRequest())
		}()

		go func() {
			defer wg.Done()
			for {
				_, err := bidi.Receive()
				if errors.Is(err, io.EOF) {
					break
				}
				require.NoError(t, err)
			}
			require.NoError(t, bidi.CloseResponse())
		}()
		wg.Wait()

		// Regardless of how many times data is sent/received, we should only have one span
		require.Len(t, clientRec.Started(), 1)
		require.Len(t, srvRec.Started(), 1)

		outgoingSpan := clientRec.Started()[0]
		incomingSpan := srvRec.Started()[0]
		inspectAndCompareSpans(t,
			incomingSpan,
			[]attribute.KeyValue{
				semconv.RPCMethodKey.String("CumSum"),
			},
			outgoingSpan,
			[]attribute.KeyValue{
				semconv.RPCMethodKey.String("CumSum"),
			}, false)
	})

	t.Run("BidiStreaming-Multi", func(t *testing.T) {
		srvRec := tracetest.NewSpanRecorder()
		clientRec := tracetest.NewSpanRecorder()

		mux := http.NewServeMux()
		mux.Handle(
			pingv1connect.NewPingServiceHandler(
				&pingServer{},
				connect.WithInterceptors(NewOtelInterceptor(
					WithTracerProvider(trace.NewTracerProvider(trace.WithSpanProcessor(srvRec))))),
			),
		)

		// Bidi requires http/2
		server := httptest.NewUnstartedServer(mux)
		server.EnableHTTP2 = true
		server.StartTLS()
		defer server.Close()

		client := pingv1connect.NewPingServiceClient(
			server.Client(),
			server.URL,
			connect.WithInterceptors(NewOtelInterceptor(WithTracerProvider(trace.NewTracerProvider(trace.WithSpanProcessor(clientRec))))),
		)

		bidi := client.CumSum(context.Background())

		for i := 0; i < 2; i++ {
			require.NoError(t, bidi.Send(&v1.CumSumRequest{Number: int64(i)}))
			_, err := bidi.Receive()
			require.NoError(t, err)
		}
		require.NoError(t, bidi.CloseRequest())
		require.NoError(t, bidi.CloseResponse())

		// Regardless of how many times data is sent/received, we should only have one span
		require.Len(t, clientRec.Started(), 1)
		require.Len(t, srvRec.Started(), 1)

		// there must be many events
		// the client span should have sent/recv, sent/recv, close/close
		outgoingSpan := clientRec.Started()[0]
		// the server span should have sent/recv, sent/recv
		incomingSpan := srvRec.Started()[0]
		inspectAndCompareSpans(t,
			incomingSpan,
			[]attribute.KeyValue{
				semconv.RPCMethodKey.String("CumSum"),
			},
			outgoingSpan,
			[]attribute.KeyValue{
				semconv.RPCMethodKey.String("CumSum"),
			}, false)
	})
}

func inspectAndCompareSpans(t *testing.T, server trace.ReadWriteSpan, serverAttrs []attribute.KeyValue, client trace.ReadWriteSpan, clientAttrs []attribute.KeyValue, isError bool) {
	// The trace ID must be equivalent
	require.Equal(t, client.SpanContext().TraceID(), server.SpanContext().TraceID())

	// The attributes must be present
	defaultAttrs := []attribute.KeyValue{
		RPCSystemConnect,
		semconv.RPCServiceKey.String("connect.ping.v1.PingService"),
		semconv.NetPeerIPKey.String("127.0.0.1"),
	}

	cattrs := client.Attributes()
	for _, elem := range append(clientAttrs, defaultAttrs...) {
		require.Contains(t, cattrs, elem)
	}

	sattrs := server.Attributes()
	for _, elem := range append(serverAttrs, defaultAttrs...) {
		require.Contains(t, sattrs, elem)
	}

	if isError {
		require.Equal(t, codes.Error, server.Status().Code)
		require.Equal(t, codes.Error, client.Status().Code)
	}
}

// FIXME:  tidy up - lifted straight from connect tests
type pingServer struct {
	pingv1connect.UnimplementedPingServiceHandler
}

func (p pingServer) Ping(ctx context.Context, request *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	if request.Peer().Addr == "" {
		return nil, connect.NewError(connect.CodeInternal, errors.New("no peer address"))
	}

	if request.Peer().Protocol == "" {
		return nil, connect.NewError(connect.CodeInternal, errors.New("no peer protocol"))
	}

	return connect.NewResponse(
		&v1.PingResponse{
			Number: request.Msg.Number,
			Text:   request.Msg.Text,
		},
	), nil
}

func (p pingServer) Fail(ctx context.Context, request *connect.Request[v1.FailRequest]) (*connect.Response[v1.FailResponse], error) {
	if request.Peer().Addr == "" {
		return nil, connect.NewError(connect.CodeInternal, errors.New("no peer address"))
	}

	if request.Peer().Protocol == "" {
		return nil, connect.NewError(connect.CodeInternal, errors.New("no peer protocol"))
	}

	return nil, connect.NewError(connect.Code(request.Msg.Code), errors.New(errorMessage))
}

func (p pingServer) Sum(ctx context.Context, stream *connect.ClientStream[v1.SumRequest]) (*connect.Response[v1.SumResponse], error) {
	if stream.Peer().Addr == "" {
		return nil, connect.NewError(connect.CodeInternal, errors.New("no peer address"))
	}

	if stream.Peer().Protocol == "" {
		return nil, connect.NewError(connect.CodeInternal, errors.New("no peer protocol"))
	}

	var sum int64
	for stream.Receive() {
		sum += stream.Msg().Number
	}

	if stream.Err() != nil {
		return nil, stream.Err()
	}

	return connect.NewResponse(&v1.SumResponse{Sum: sum}), nil
}

func (p pingServer) CountUp(ctx context.Context, request *connect.Request[v1.CountUpRequest], stream *connect.ServerStream[v1.CountUpResponse]) error {
	if request.Peer().Addr == "" {
		return connect.NewError(connect.CodeInternal, errors.New("no peer address"))
	}

	if request.Peer().Protocol == "" {
		return connect.NewError(connect.CodeInternal, errors.New("no peer protocol"))
	}

	if request.Msg.Number <= 0 {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf(
			"number must be positive: got %v",
			request.Msg.Number,
		))
	}

	for i := int64(1); i <= request.Msg.Number; i++ {
		if err := stream.Send(&v1.CountUpResponse{Number: i}); err != nil {
			return err
		}
	}

	return nil
}

func (p pingServer) CumSum(ctx context.Context, stream *connect.BidiStream[v1.CumSumRequest, v1.CumSumResponse]) error {
	var sum int64

	if stream.Peer().Addr == "" {
		return connect.NewError(connect.CodeInternal, errors.New("no peer address"))
	}

	if stream.Peer().Protocol == "" {
		return connect.NewError(connect.CodeInternal, errors.New("no peer address"))
	}

	for {
		msg, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return err
		}
		sum += msg.Number
		if err := stream.Send(&v1.CumSumResponse{Sum: sum}); err != nil {
			return err
		}
	}
}
