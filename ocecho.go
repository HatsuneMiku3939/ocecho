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
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

// TraceOptions
type TraceOptions struct {
	IsPublicEndpoint bool
	Propagation      propagation.HTTPFormat
	StartOptions     trace.StartOptions
}

// OpenCensusMiddleware OpenCensus trace, stats middleware
func OpenCensusMiddleware(opts TraceOptions) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		m := &ocechoHandler{
			IsPublicEndpoint: opts.IsPublicEndpoint,
			Propagation:      opts.Propagation,
			StartOptions:     opts.StartOptions,
		}

		return func(c echo.Context) error {
			var tags addedTags

			c, traceEnd := m.startTrace(c)
			defer traceEnd()
			c, statsEnd := m.startStats(c)
			defer statsEnd(c, &tags)

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			return err
		}
	}
}

type ocechoHandler struct {
	IsPublicEndpoint bool
	Propagation      propagation.HTTPFormat
	StartOptions     trace.StartOptions
}

func (h *ocechoHandler) startTrace(c echo.Context) (echo.Context, func()) {
	r := c.Request()

	name := c.Path()
	ctx := r.Context()

	var span *trace.Span
	sc, ok := h.Propagation.SpanContextFromRequest(r)
	if ok && !h.IsPublicEndpoint {
		ctx, span = trace.StartSpanWithRemoteParent(ctx, name, sc,
			trace.WithSampler(h.StartOptions.Sampler),
			trace.WithSpanKind(trace.SpanKindServer))
	} else {
		ctx, span = trace.StartSpan(ctx, name,
			trace.WithSampler(h.StartOptions.Sampler),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		if ok {
			span.AddLink(trace.Link{
				TraceID:    sc.TraceID,
				SpanID:     sc.SpanID,
				Type:       trace.LinkTypeParent,
				Attributes: nil,
			})
		}
	}

	span.AddAttributes(requestAttrs(c.Path(), r)...)
	if r.Body == nil {
		// TODO: Handle cases where ContentLength is not set.
	} else if r.ContentLength > 0 {
		span.AddMessageReceiveEvent(0, r.ContentLength, -1)
	}

	r = r.WithContext(ctx)
	c.SetRequest(r)
	return c, span.End
}

func (h *ocechoHandler) startStats(c echo.Context) (echo.Context, func(echo.Context, *addedTags)) {
	r := c.Request()

	ctx, _ := tag.New(r.Context(),
		tag.Upsert(Host, r.Host),
		tag.Upsert(Path, r.URL.Path),
		tag.Upsert(Method, r.Method))
	track := &statsTracker{
		start: time.Now(),
		ctx:   ctx,
	}

	if r.Body == nil {
		track.reqSize = -1
	} else if r.ContentLength > 0 {
		track.reqSize = r.ContentLength
	}

	r = r.WithContext(ctx)
	c.SetRequest(r)
	stats.Record(ctx, ServerRequestCount.M(1))
	return c, track.end
}

type statsTracker struct {
	ctx     context.Context
	reqSize int64
	start   time.Time
}

func (t *statsTracker) end(c echo.Context, tags *addedTags) {
	statusCode := c.Response().Status
	if statusCode == 0 {
		statusCode = 200
	}
	statusLine := http.StatusText(statusCode)

	span := trace.FromContext(t.ctx)
	span.SetStatus(TraceStatus(statusCode, statusLine))
	span.AddAttributes(trace.Int64Attribute(StatusCodeAttribute, int64(statusCode)))

	m := []stats.Measurement{
		ServerLatency.M(float64(time.Since(t.start)) / float64(time.Millisecond)),
		ServerResponseBytes.M(c.Response().Size),
	}
	if t.reqSize >= 0 {
		m = append(m, ServerRequestBytes.M(t.reqSize))
	}
	allTags := make([]tag.Mutator, len(tags.t)+1)
	allTags[0] = tag.Upsert(StatusCode, strconv.Itoa(statusCode))
	copy(allTags[1:], tags.t)
	stats.RecordWithTags(t.ctx, allTags, m...)
}
