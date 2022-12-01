package otelconnect

import (
	"context"
	"github.com/bufbuild/connect-go"
)

func ExampleUnaryInterceptorFunc() {
	client := pingv1connect.NewPingServiceClient(
		examplePingServer.Client(),
		examplePingServer.URL(),
		connect.WithInterceptors(loggingInterceptor),
	)
	if _, err := client.Ping(context.Background(), connect.NewRequest(&pingv1.PingRequest{Number: 42})); err != nil {
		logger.Println("error:", err)
		return
	}

	// Output:
	// calling: /connect.ping.v1.PingService/Ping
	// request: number:42
	// response: number:42
}
