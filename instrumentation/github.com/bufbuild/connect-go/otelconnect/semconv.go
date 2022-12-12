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

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

// Semantic conventions for attribute keys for connect.
const (
	// Name of message transmitted or received.
	RPCNameKey = attribute.Key("name")

	// Type of message transmitted or received.
	RPCEventTypeKey = attribute.Key("type")
)

// Semantic conventions for common RPC attributes.
var (
	// Semantic convention for connect as the remoting system.
	RPCSystemConnect = semconv.RPCSystemKey.String("connect")

	// Semantic convention for a message named message.
	RPCNameMessage = RPCNameKey.String("message")

	// Semantic conventions for RPC message types.
	RPCEventTypeSent           = RPCEventTypeKey.String("SENT")
	RPCEventTypeReceived       = RPCEventTypeKey.String("RECEIVED")
	RPCEventTypeClosedRequest  = RPCEventTypeKey.String("CLOSED-REQ")
	RPCEventTypeClosedResponse = RPCEventTypeKey.String("CLOSED-RES")
)
