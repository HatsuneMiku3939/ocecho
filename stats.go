// Copyright 2017, ocecho Authors
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
//

package ocecho

import (
	"net/http"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
)

var (
	Host           = tag.MustNewKey("http.host")
	StatusCode     = tag.MustNewKey("http.status")
	Path           = tag.MustNewKey("http.path")
	Method         = tag.MustNewKey("http.method")
	KeyServerRoute = tag.MustNewKey("http_server_route")
)

var (
	ServerRequestCount = stats.Int64(
		"opencensus.io/http/server/request_count",
		"Number of HTTP requests started",
		stats.UnitDimensionless)
	ServerRequestBytes = stats.Int64(
		"opencensus.io/http/server/request_bytes",
		"HTTP request body size if set as ContentLength (uncompressed)",
		stats.UnitBytes)
	ServerResponseBytes = stats.Int64(
		"opencensus.io/http/server/response_bytes",
		"HTTP response body size (uncompressed)",
		stats.UnitBytes)
	ServerLatency = stats.Float64(
		"opencensus.io/http/server/latency",
		"End-to-end latency",
		stats.UnitMilliseconds)
)

var (
	ServerRequestCountView = &view.View{
		Name:        "opencensus.io/http/server/request_count",
		Description: "Count of HTTP requests started",
		Measure:     ServerRequestCount,
		Aggregation: view.Count(),
	}

	ServerRequestBytesView = &view.View{
		Name:        "opencensus.io/http/server/request_bytes",
		Description: "Size distribution of HTTP request body",
		Measure:     ServerRequestBytes,
		Aggregation: DefaultSizeDistribution,
	}

	ServerResponseBytesView = &view.View{
		Name:        "opencensus.io/http/server/response_bytes",
		Description: "Size distribution of HTTP response body",
		Measure:     ServerResponseBytes,
		Aggregation: DefaultSizeDistribution,
	}

	ServerLatencyView = &view.View{
		Name:        "opencensus.io/http/server/latency",
		Description: "Latency distribution of HTTP requests",
		Measure:     ServerLatency,
		Aggregation: DefaultLatencyDistribution,
	}

	ServerRequestCountByMethod = &view.View{
		Name:        "opencensus.io/http/server/request_count_by_method",
		Description: "Server request count by HTTP method",
		TagKeys:     []tag.Key{Method},
		Measure:     ServerRequestCount,
		Aggregation: view.Count(),
	}

	ServerResponseCountByStatusCode = &view.View{
		Name:        "opencensus.io/http/server/response_count_by_status_code",
		Description: "Server response count by status code",
		TagKeys:     []tag.Key{StatusCode},
		Measure:     ServerLatency,
		Aggregation: view.Count(),
	}
)

var (
	DefaultSizeDistribution    = view.Distribution(1024, 2048, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216, 67108864, 268435456, 1073741824, 4294967296)
	DefaultLatencyDistribution = view.Distribution(1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 130, 160, 200, 250, 300, 400, 500, 650, 800, 1000, 2000, 5000, 10000, 20000, 50000, 100000)
)

func TraceStatus(httpStatusCode int, statusLine string) trace.Status {
	var code int32
	if httpStatusCode < 200 || httpStatusCode >= 400 {
		code = trace.StatusCodeUnknown
	}
	switch httpStatusCode {
	case 499:
		code = trace.StatusCodeCancelled
	case http.StatusBadRequest:
		code = trace.StatusCodeInvalidArgument
	case http.StatusUnprocessableEntity:
		code = trace.StatusCodeInvalidArgument
	case http.StatusGatewayTimeout:
		code = trace.StatusCodeDeadlineExceeded
	case http.StatusNotFound:
		code = trace.StatusCodeNotFound
	case http.StatusForbidden:
		code = trace.StatusCodePermissionDenied
	case http.StatusUnauthorized: // 401 is actually unauthenticated.
		code = trace.StatusCodeUnauthenticated
	case http.StatusTooManyRequests:
		code = trace.StatusCodeResourceExhausted
	case http.StatusNotImplemented:
		code = trace.StatusCodeUnimplemented
	case http.StatusServiceUnavailable:
		code = trace.StatusCodeUnavailable
	case http.StatusOK:
		code = trace.StatusCodeOK
	}
	return trace.Status{Code: code, Message: codeToStr[code]}
}

var codeToStr = map[int32]string{
	trace.StatusCodeOK:                 `OK`,
	trace.StatusCodeCancelled:          `CANCELLED`,
	trace.StatusCodeUnknown:            `UNKNOWN`,
	trace.StatusCodeInvalidArgument:    `INVALID_ARGUMENT`,
	trace.StatusCodeDeadlineExceeded:   `DEADLINE_EXCEEDED`,
	trace.StatusCodeNotFound:           `NOT_FOUND`,
	trace.StatusCodeAlreadyExists:      `ALREADY_EXISTS`,
	trace.StatusCodePermissionDenied:   `PERMISSION_DENIED`,
	trace.StatusCodeResourceExhausted:  `RESOURCE_EXHAUSTED`,
	trace.StatusCodeFailedPrecondition: `FAILED_PRECONDITION`,
	trace.StatusCodeAborted:            `ABORTED`,
	trace.StatusCodeOutOfRange:         `OUT_OF_RANGE`,
	trace.StatusCodeUnimplemented:      `UNIMPLEMENTED`,
	trace.StatusCodeInternal:           `INTERNAL`,
	trace.StatusCodeUnavailable:        `UNAVAILABLE`,
	trace.StatusCodeDataLoss:           `DATA_LOSS`,
	trace.StatusCodeUnauthenticated:    `UNAUTHENTICATED`,
}

var DefaultServerViews = []*view.View{
	ServerRequestCountView,
	ServerRequestBytesView,
	ServerResponseBytesView,
	ServerLatencyView,
	ServerRequestCountByMethod,
	ServerResponseCountByStatusCode,
}
