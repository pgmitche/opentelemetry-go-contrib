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
	"path"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/connect/otelconnect"
)

type rpcPath struct {
	service string
	method  string
}

// splitFullMethod splits the method name and returns an rpcPath object
// that has divided service and method names.
func splitFullMethod(i *otelconnect.InterceptorInfo) rpcPath {
	s, m := path.Split(i.Method)
	if s != "" {
		s = path.Clean(s)
		s = strings.TrimLeft(s, "/")
	}

	return rpcPath{
		service: s,
		method:  m,
	}
}

// Any takes a list of Filters and returns a Filter that
// returns true if any Filter in the list returns true.
func Any(fs ...otelconnect.Filter) otelconnect.Filter {
	return func(i *otelconnect.InterceptorInfo) bool {
		for _, f := range fs {
			if f(i) {
				return true
			}
		}
		return false
	}
}

// All takes a list of Filters and returns a Filter that
// returns true only if all Filters in the list return true.
func All(fs ...otelconnect.Filter) otelconnect.Filter {
	return func(i *otelconnect.InterceptorInfo) bool {
		for _, f := range fs {
			if !f(i) {
				return false
			}
		}
		return true
	}
}

// None takes a list of Filters and returns a Filter that returns
// true only if none of the Filters in the list return true.
func None(fs ...otelconnect.Filter) otelconnect.Filter {
	return Not(Any(fs...))
}

// Not provides a convenience mechanism for inverting a Filter.
func Not(f otelconnect.Filter) otelconnect.Filter {
	return func(i *otelconnect.InterceptorInfo) bool {
		return !f(i)
	}
}

// MethodName returns a Filter that returns true if the request's
// method name matches the provided string n.
func MethodName(n string) otelconnect.Filter {
	return func(i *otelconnect.InterceptorInfo) bool {
		p := splitFullMethod(i)
		return p.method == n
	}
}

// MethodPrefix returns a Filter that returns true if the request's
// method starts with the provided string pre.
func MethodPrefix(pre string) otelconnect.Filter {
	return func(i *otelconnect.InterceptorInfo) bool {
		p := splitFullMethod(i)
		return strings.HasPrefix(p.method, pre)
	}
}

// FullMethodName returns a Filter that returns true if the request's
// full RPC method string, i.e. /package.service/method, starts with
// the provided string n.
func FullMethodName(n string) otelconnect.Filter {
	return func(i *otelconnect.InterceptorInfo) bool {
		return i.Method == n
	}
}

// ServiceName returns a Filter that returns true if the request's
// service name, i.e. package.service, matches s.
func ServiceName(s string) otelconnect.Filter {
	return func(i *otelconnect.InterceptorInfo) bool {
		p := splitFullMethod(i)
		return p.service == s
	}
}

// ServicePrefix returns a Filter that returns true if the request's
// service name, i.e. package.service, starts with the provided string pre.
func ServicePrefix(pre string) otelconnect.Filter {
	return func(i *otelconnect.InterceptorInfo) bool {
		p := splitFullMethod(i)
		return strings.HasPrefix(p.service, pre)
	}
}

// HealthCheck returns a Filter that returns true if the request's
// service name is health check defined by gRPC Health Checking Protocol.
// https://github.com/grpc/grpc/blob/master/doc/health-checking.md
//
// TODO: adapt for connect's healthcheck
func HealthCheck() otelconnect.Filter {
	return ServicePrefix("grpc.health.v1.Health")
}
