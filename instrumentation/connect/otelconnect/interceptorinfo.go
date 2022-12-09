package otelconnect

import "github.com/bufbuild/connect-go"

// InterceptorInfo is the union of some arguments to four types of
// interceptors.
type InterceptorInfo struct {
	// Method is method name registered to UnaryClient and StreamClient
	Method string
	// Type is the RPC's stream type, whether unary, server/client side or bidi
	Type connect.StreamType
}
