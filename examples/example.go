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

package main

import (
	"log"
	"net/http"
	"time"

	"github.com/HatsuneMiku3939/ocecho"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.opencensus.io/zpages"
)

const (
	metricsLogFile = "/tmp/metrics.log"
	tracesLogFile  = "/tmp/trace.log"
)

func main() {
	// Start z-Pages server.
	go func() {
		mux := http.NewServeMux()
		zpages.Handle(mux, "/debug")
		log.Fatal(http.ListenAndServe("127.0.0.1:8081", mux))
	}()

	// Using log exporter to export metrics but you can choose any supported exporter.
	exporter, err := exporter.NewLogExporter(exporter.Options{
		ReportingInterval: 10 * time.Second,
		MetricsLogFile:    metricsLogFile,
		TracesLogFile:     tracesLogFile,
	})
	if err != nil {
		log.Fatalf("Error creating log exporter: %v", err)
	}
	exporter.Start()
	defer exporter.Stop()
	defer exporter.Close()

	// Always trace for this demo. In a production application, you should
	// configure this to a trace.ProbabilitySampler set at the desired
	// probability.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	// Report stats at every second.
	view.SetReportingPeriod(1 * time.Second)

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// ocecho Middleware
	e.Use(ocecho.OpenCensusMiddleware(
		ocecho.OpenCensusConfig{
			Skipper: middleware.DefaultSkipper,
			TraceOptions: ocecho.TraceOptions{
				IsPublicEndpoint: true,
				Propagation:      &b3.HTTPFormat{},
				StartOptions:     trace.StartOptions{},
			},
		},
	))

	// Register server views
	if err := view.Register(ocecho.DefaultServerViews...); err != nil {
		log.Fatalf("Error creating metric views: %v", err)
	}

	// Route => handler
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!\n")
	})

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
