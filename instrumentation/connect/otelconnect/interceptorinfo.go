package otelconnect

// InterceptorInfo is the union of some arguments to four types of
// interceptors.
type InterceptorInfo struct {
	// Method is method name registered to UnaryClient and StreamClient
	Method string
	// TODO: is(nt)Stream, isServer, isClient
}
