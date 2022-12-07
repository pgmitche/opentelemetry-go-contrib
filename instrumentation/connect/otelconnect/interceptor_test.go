package otelconnect

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
	"testing"
)

const (
	errorMessage = "RIP"
)

func TestInterceptors(t *testing.T) {
	// Example of any/all propagators
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}, b3.New(), b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)), jaeger.Jaeger{}))

	tests := []struct {
		name          string
		srvRec        *tracetest.SpanRecorder
		clientRec     *tracetest.SpanRecorder
		interceptorFn func(...Option) *otelInterceptor
	}{
		{
			name:          "Unary",
			srvRec:        tracetest.NewSpanRecorder(),
			clientRec:     tracetest.NewSpanRecorder(),
			interceptorFn: NewOtelInterceptor,
		},
	}

	// FIXME: No propagation, nil propagators
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.Handle(
				pingv1connect.NewPingServiceHandler(
					&pingServer{},
					connect.WithInterceptors(test.interceptorFn(
						WithTracerProvider(trace.NewTracerProvider(trace.WithSpanProcessor(test.srvRec))))),
				),
			)

			server := httptest.NewServer(mux)
			defer server.Close()

			client := pingv1connect.NewPingServiceClient(
				server.Client(),
				server.URL,
				connect.WithInterceptors(test.interceptorFn(WithTracerProvider(trace.NewTracerProvider(trace.WithSpanProcessor(test.clientRec))))),
			)

			res, err := client.Ping(context.Background(), connect.NewRequest(&v1.PingRequest{
				Number: 0,
				Text:   "Ping",
			}))

			// FIXME: at _minimum_ trace should persist
			outgoingSpan := test.clientRec.Started()[0]
			incomingSpan := test.srvRec.Started()[0]
			require.Equal(t, outgoingSpan.SpanContext().TraceID(), incomingSpan.SpanContext().TraceID())

			log.Println(res, err)
		})
	}
}

// FIXME:  tidy up - lifted straight from connect tests
type pingServer struct {
	pingv1connect.UnimplementedPingServiceHandler

	//checkMetadata bool
}

func (p pingServer) Ping(ctx context.Context, request *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	//if err := expectClientHeader(p.checkMetadata, request); err != nil {
	//	return nil, err
	//}

	if request.Peer().Addr == "" {
		return nil, connect.NewError(connect.CodeInternal, errors.New("no peer address"))
	}

	if request.Peer().Protocol == "" {
		return nil, connect.NewError(connect.CodeInternal, errors.New("no peer protocol"))
	}

	response := connect.NewResponse(
		&v1.PingResponse{
			Number: request.Msg.Number,
			Text:   request.Msg.Text,
		},
	)

	//response.Header().Set(handlerHeader, headerValue)
	//response.Trailer().Set(handlerTrailer, trailerValue)
	return response, nil
}

func (p pingServer) Fail(ctx context.Context, request *connect.Request[v1.FailRequest]) (*connect.Response[v1.FailResponse], error) {
	//if err := expectClientHeader(p.checkMetadata, request); err != nil {
	//	return nil, err
	//}

	if request.Peer().Addr == "" {
		return nil, connect.NewError(connect.CodeInternal, errors.New("no peer address"))
	}

	if request.Peer().Protocol == "" {
		return nil, connect.NewError(connect.CodeInternal, errors.New("no peer protocol"))
	}

	err := connect.NewError(connect.Code(request.Msg.Code), errors.New(errorMessage))
	//err.Meta().Set(handlerHeader, headerValue)
	//err.Meta().Set(handlerTrailer, trailerValue)
	return nil, err
}

func (p pingServer) Sum(ctx context.Context, stream *connect.ClientStream[v1.SumRequest]) (*connect.Response[v1.SumResponse], error) {
	//if p.checkMetadata {
	//	if err := expectMetadata(stream.RequestHeader(), "header", clientHeader, headerValue); err != nil {
	//		return nil, err
	//	}
	//}

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

	response := connect.NewResponse(&v1.SumResponse{Sum: sum})
	//response.Header().Set(handlerHeader, headerValue)
	//response.Trailer().Set(handlerTrailer, trailerValue)
	return response, nil
}

func (p pingServer) CountUp(ctx context.Context, request *connect.Request[v1.CountUpRequest], stream *connect.ServerStream[v1.CountUpResponse]) error {
	//if err := expectClientHeader(p.checkMetadata, request); err != nil {
	//	return err
	//}

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

	//stream.ResponseHeader().Set(handlerHeader, headerValue)
	//stream.ResponseTrailer().Set(handlerTrailer, trailerValue)
	for i := int64(1); i <= request.Msg.Number; i++ {
		if err := stream.Send(&v1.CountUpResponse{Number: i}); err != nil {
			return err
		}
	}

	return nil
}

func (p pingServer) CumSum(ctx context.Context, stream *connect.BidiStream[v1.CumSumRequest, v1.CumSumResponse]) error {
	var sum int64
	//if p.checkMetadata {
	//	if err := expectMetadata(stream.RequestHeader(), "header", clientHeader, headerValue); err != nil {
	//		return err
	//	}
	//}

	if stream.Peer().Addr == "" {
		return connect.NewError(connect.CodeInternal, errors.New("no peer address"))
	}

	if stream.Peer().Protocol == "" {
		return connect.NewError(connect.CodeInternal, errors.New("no peer address"))
	}

	//stream.ResponseHeader().Set(handlerHeader, headerValue)
	//stream.ResponseTrailer().Set(handlerTrailer, trailerValue)
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
