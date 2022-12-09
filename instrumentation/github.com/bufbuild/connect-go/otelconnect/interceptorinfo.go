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

package otelconnect // import "go.opentelemetry.io/contrib/instrumentation/connect/otelconnect"

import "github.com/bufbuild/connect-go"

// InterceptorInfo is the union of some arguments to four types of
// interceptors.
type InterceptorInfo struct {
	// Method is method name registered to UnaryClient and StreamClient
	Method string
	// Type is the RPC's stream type, whether unary, server/client side or bidi
	Type connect.StreamType
}
