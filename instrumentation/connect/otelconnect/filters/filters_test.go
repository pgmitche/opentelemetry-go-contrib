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

package filters // import "go.opentelemetry.io/contrib/instrumentation/connect/otelconnect/filters"

import (
	"github.com/bufbuild/connect-go"
	"go.opentelemetry.io/contrib/instrumentation/connect/otelconnect"
	"testing"
)

type testCase struct {
	name string
	i    *otelconnect.InterceptorInfo
	f    otelconnect.Filter
	want bool
}

func TestMethodName(t *testing.T) {
	const dummyFullMethodName = "/example.HelloService/Hello"
	tcs := []testCase{
		{
			name: "unary client interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    MethodName("Hello"),
			want: true,
		},
		{
			name: "stream client interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeClient},
			f:    MethodName("Hello"),
			want: true,
		},
		{
			name: "unary server interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    MethodName("Hello"),
			want: true,
		},
		{
			name: "stream server interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeServer},
			f:    MethodName("Hello"),
			want: true,
		},
		{
			name: "unary client interceptor fail",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    MethodName("Goodbye"),
			want: false,
		},
	}

	for _, tc := range tcs {
		out := tc.f(tc.i)
		if tc.want != out {
			t.Errorf("test case '%v' failed, wanted %v but obtained %v", tc.name, tc.want, out)
		}
	}
}

func TestMethodPrefix(t *testing.T) {
	const dummyFullMethodName = "/example.HelloService/FoobarHello"
	tcs := []testCase{
		{
			name: "unary client interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    MethodPrefix("Foobar"),
			want: true,
		},
		{
			name: "stream client interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeClient},
			f:    MethodPrefix("Foobar"),
			want: true,
		},
		{
			name: "unary server interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    MethodPrefix("Foobar"),
			want: true,
		},
		{
			name: "stream server interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeServer},
			f:    MethodPrefix("Foobar"),
			want: true,
		},
		{
			name: "unary client interceptor fail",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    MethodPrefix("Barfoo"),
			want: false,
		},
	}
	for _, tc := range tcs {
		out := tc.f(tc.i)
		if tc.want != out {
			t.Errorf("test case '%v' failed, wanted %v but obtained %v", tc.name, tc.want, out)
		}
	}
}

func TestFullMethodName(t *testing.T) {
	const dummyFullMethodName = "/example.HelloService/Hello"
	tcs := []testCase{
		{
			name: "unary client interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    FullMethodName(dummyFullMethodName),
			want: true,
		},
		{
			name: "stream client interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeClient},
			f:    FullMethodName(dummyFullMethodName),
			want: true,
		},
		{
			name: "unary server interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    FullMethodName(dummyFullMethodName),
			want: true,
		},
		{
			name: "stream server interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeServer},
			f:    FullMethodName(dummyFullMethodName),
			want: true,
		},
		{
			name: "unary client interceptor fail",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    FullMethodName("/example.HelloService/Goodbye"),
			want: false,
		},
	}

	for _, tc := range tcs {
		out := tc.f(tc.i)
		if tc.want != out {
			t.Errorf("test case '%v' failed, wanted %v but obtained %v", tc.name, tc.want, out)
		}
	}
}

func TestServiceName(t *testing.T) {
	const dummyFullMethodName = "/example.HelloService/Hello"

	tcs := []testCase{
		{
			name: "unary client interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    ServiceName("example.HelloService"),
			want: true,
		},
		{
			name: "stream client interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeClient},
			f:    ServiceName("example.HelloService"),
			want: true,
		},
		{
			name: "unary server interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    ServiceName("example.HelloService"),
			want: true,
		},
		{
			name: "stream server interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeServer},
			f:    ServiceName("example.HelloService"),
			want: true,
		},
		{
			name: "unary client interceptor fail",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    ServiceName("opentelemetry.HelloService"),
			want: false,
		},
	}

	for _, tc := range tcs {
		out := tc.f(tc.i)
		if tc.want != out {
			t.Errorf("test case '%v' failed, wanted %v but obtained %v", tc.name, tc.want, out)
		}
	}
}

func TestServicePrefix(t *testing.T) {
	const dummyFullMethodName = "/example.HelloService/FoobarHello"
	tcs := []testCase{
		{
			name: "unary client interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    ServicePrefix("example"),
			want: true,
		},
		{
			name: "stream client interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeClient},
			f:    ServicePrefix("example"),
			want: true,
		},
		{
			name: "unary server interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    ServicePrefix("example"),
			want: true,
		},
		{
			name: "stream server interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeServer},
			f:    ServicePrefix("example"),
			want: true,
		},
		{
			name: "unary client interceptor fail",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    ServicePrefix("opentelemetry"),
			want: false,
		},
	}
	for _, tc := range tcs {
		out := tc.f(tc.i)
		if tc.want != out {
			t.Errorf("test case '%v' failed, wanted %v but obtained %v", tc.name, tc.want, out)
		}
	}
}

func TestAny(t *testing.T) {
	const dummyFullMethodName = "/example.HelloService/FoobarHello"
	tcs := []testCase{
		{
			name: "unary client interceptor true && true",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    Any(MethodName("FoobarHello"), MethodPrefix("Foobar")),
			want: true,
		},
		{
			name: "unary client interceptor false && true",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    Any(MethodName("Hello"), MethodPrefix("Foobar")),
			want: true,
		},
		{
			name: "unary client interceptor false && false",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    Any(MethodName("Goodbye"), MethodPrefix("Barfoo")),
			want: false,
		},
	}
	for _, tc := range tcs {
		out := tc.f(tc.i)
		if tc.want != out {
			t.Errorf("test case '%v' failed, wanted %v but obtained %v", tc.name, tc.want, out)
		}
	}
}

func TestAll(t *testing.T) {
	const dummyFullMethodName = "/example.HelloService/FoobarHello"
	tcs := []testCase{
		{
			name: "unary client interceptor true && true",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    All(MethodName("FoobarHello"), MethodPrefix("Foobar")),
			want: true,
		},
		{
			name: "unary client interceptor true && false",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    All(MethodName("FoobarHello"), MethodPrefix("Barfoo")),
			want: false,
		},
		{
			name: "unary client interceptor false && false",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    All(MethodName("FoobarGoodbye"), MethodPrefix("Barfoo")),
			want: false,
		},
	}
	for _, tc := range tcs {
		out := tc.f(tc.i)
		if tc.want != out {
			t.Errorf("test case '%v' failed, wanted %v but obtained %v", tc.name, tc.want, out)
		}
	}
}

func TestNone(t *testing.T) {
	const dummyFullMethodName = "/example.HelloService/FoobarHello"
	tcs := []testCase{
		{
			name: "unary client interceptor true && true",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    None(MethodName("FoobarHello"), MethodPrefix("Foobar")),
			want: false,
		},
		{
			name: "unary client interceptor true && false",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    None(MethodName("FoobarHello"), MethodPrefix("Barfoo")),
			want: false,
		},
		{
			name: "unary client interceptor false && false",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    None(MethodName("FoobarGoodbye"), MethodPrefix("Barfoo")),
			want: true,
		},
	}
	for _, tc := range tcs {
		out := tc.f(tc.i)
		if tc.want != out {
			t.Errorf("test case '%v' failed, wanted %v but obtained %v", tc.name, tc.want, out)
		}
	}
}

func TestNot(t *testing.T) {
	const dummyFullMethodName = "/example.HelloService/FoobarHello"
	tcs := []testCase{
		{
			name: "methodname not",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    Not(MethodName("FoobarHello")),
			want: false,
		},
		{
			name: "method prefix not",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    Not(MethodPrefix("FoobarHello")),
			want: false,
		},
		{
			name: "unary client interceptor not all(true && true)",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethodName, Type: connect.StreamTypeUnary},
			f:    Not(All(MethodName("FoobarHello"), MethodPrefix("Foobar"))),
			want: false,
		},
	}

	for _, tc := range tcs {
		out := tc.f(tc.i)
		if tc.want != out {
			t.Errorf("test case '%v' failed, wanted %v but obtained %v", tc.name, tc.want, out)
		}
	}
}

func TestHealthCheck(t *testing.T) {
	const (
		healthCheck     = "/grpc.health.v1.Health/Check"
		dummyFullMethod = "/example.HelloService/FoobarHello"
	)
	tcs := []testCase{
		{
			name: "unary client interceptor healthcheck",
			i:    &otelconnect.InterceptorInfo{Method: healthCheck, Type: connect.StreamTypeUnary},
			f:    HealthCheck(),
			want: true,
		},
		{
			name: "stream client interceptor healthcheck",
			i:    &otelconnect.InterceptorInfo{Method: healthCheck, Type: connect.StreamTypeClient},
			f:    HealthCheck(),
			want: true,
		},
		{
			name: "unary server interceptor healthcheck",
			i:    &otelconnect.InterceptorInfo{Method: healthCheck, Type: connect.StreamTypeUnary},
			f:    HealthCheck(),
			want: true,
		},
		{
			name: "stream server interceptor healthcheck",
			i:    &otelconnect.InterceptorInfo{Method: healthCheck, Type: connect.StreamTypeServer},
			f:    HealthCheck(),
			want: true,
		},
		{
			name: "unary client interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethod, Type: connect.StreamTypeUnary},
			f:    HealthCheck(),
			want: false,
		},
		{
			name: "stream client interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethod, Type: connect.StreamTypeClient},
			f:    HealthCheck(),
			want: false,
		},
		{
			name: "unary server interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethod, Type: connect.StreamTypeUnary},
			f:    HealthCheck(),
			want: false,
		},
		{
			name: "stream server interceptor",
			i:    &otelconnect.InterceptorInfo{Method: dummyFullMethod, Type: connect.StreamTypeServer},
			f:    HealthCheck(),
			want: false,
		},
	}

	for _, tc := range tcs {
		out := tc.f(tc.i)
		if tc.want != out {
			t.Errorf("test case '%v' failed, wanted %v but obtained %v", tc.name, tc.want, out)
		}
	}
}
