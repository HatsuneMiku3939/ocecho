# ocecho
> OpenCensus instrumentation for Echo framework

[![License](https://img.shields.io/badge/License-APACHE-blue.svg?style=flat-square)](https://github.com/HatsuneMiku3939/ocecho/blob/master/LICENSE)


The Echo middlleware provide OpenCensus instrumentations.  It provide tracing and metrics features as same as `ochttp`.
Heavily inspired `ochttp` official plugin. Many parts of `ocecho` has been copied from `ochttp`.

Thanks for authors of `ochttp`.


## Installation

Requires Go 1.12 or later.

```sh
go get github.com/HatsuneMiku3939/ocecho
```

## Usage example

```go
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
```

You can found whole example in ``examples``.


## Release History

* 0.3.2
    * Add default Propagation option
* 0.3.1
    * fix nill reference
* 0.3.0
    * `Broke v0.2.0 compatibility`
    * Add Skipper option
    * Update echo.v4
* 0.2.0
    * Add IsPublicEndpoint option
    * Fix Span name as same as `ochttp`
* 0.1.1
    * Fix RequestSize, ResponseSize metric bug
* 0.1.0
    * Initial release

## TODO

* [ ] Add unittest

## Meta

Distributed under the Apache license. See ``LICENSE`` for more information.

[https://github.com/HatsuneMiku3939/](https://github.com/HatsuneMiku3939/)

## Contributing

1. Fork it (<https://github.com/HatsuneMiku3939/ocecho/fork>)
2. Create your feature branch (`git checkout -b feature/fooBar`)
3. Commit your changes (`git commit -am 'Add some fooBar'`)
4. Push to the branch (`git push origin feature/fooBar`)
5. Create a new Pull Request
