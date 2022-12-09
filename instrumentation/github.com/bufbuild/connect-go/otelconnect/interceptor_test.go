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
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

const (
	errorMessage = "RIP"
)

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

		res, err := client.Ping(context.Background(), connect.NewRequest(&v1.PingRequest{
			Number: 0,
			Text:   "Ping",
		}))

		// FIXME: at _minimum_ trace should persist
		outgoingSpan := clientRec.Started()[0]
		incomingSpan := srvRec.Started()[0]
		require.Equal(t, outgoingSpan.SpanContext().TraceID(), incomingSpan.SpanContext().TraceID())

		log.Println(res, err)
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
		res, err := stream.CloseAndReceive()
		require.NoError(t, err)

		log.Println(res)

		// FIXME: at _minimum_ trace should persist
		outgoingSpan := clientRec.Started()[0]
		incomingSpan := srvRec.Started()[0]
		require.Equal(t, outgoingSpan.SpanContext().TraceID(), incomingSpan.SpanContext().TraceID())
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
		}

		// FIXME: at _minimum_ trace should persist
		outgoingSpan := clientRec.Started()[0]
		incomingSpan := srvRec.Started()[0]
		require.Equal(t, outgoingSpan.SpanContext().TraceID(), incomingSpan.SpanContext().TraceID())
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

		// FIXME: at _minimum_ trace should persist
		outgoingSpan := clientRec.Started()[0]
		incomingSpan := srvRec.Started()[0]
		require.Equal(t, outgoingSpan.SpanContext().TraceID(), incomingSpan.SpanContext().TraceID())
	})
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
