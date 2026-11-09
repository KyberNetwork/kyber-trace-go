# Kyber Trace GO Lib

## Overview
This is a library that makes it easier to integrate with Kyber's self-hosted open tracing

## Add this lib to your project
- Require go version 1.27.3 or later
- Step 1:
```
$ export GOPRIVATE=github.com/KyberNetwork/kyber-trace-go
```
- Step 2: Add file `tools/tools.go` with content:
```
package tools

import (
    _ "github.com/KyberNetwork/kyber-trace-go/tools"
)
```
- Step 3:
```
$ go mod tidy
$ go mod vendor
```

## Update to latest version
```
$ go get -u github.com/KyberNetwork/kyber-trace-go
$ go mod vendor
```

## Docker build
Please adjust your `RUN go mod download` statement in your Dockerfile to `RUN GO111MODULE=on GOPRIVATE=github.com/KyberNetwork/kybers-trace-error go mod download`

## How to use

### Init global `TracerProvider`
In order to push spans and traces to Kyber's self-hosted agent, you have to initialize a TracerProvider. When you add the following statements into your code: `import _ "github.com/KyberNetwork/kyber-trace-go/pkg/tracer"`, `kyber-trace-go`, it will initialize a new TraceProvider, set it to global TracerProvider of `otel` package. Whenever you want to get the Tracer to start (or get) a span, just use the function `tracer.Tracer()` (refer the example at https://github.com/KyberNetwork/kyber-trace-go/blob/main/example/tracer.go for more details).
### Init global `MeterProvider`
In order to push customized metrics to Kyber's self-hosted agent, you have to initialize a MeterProvider. When you add the following statement into your code: `import _ "github.com/KyberNetwork/kyber-trace-go/pkg/metric"`, `kyber-trace-go`, it will initialize a new MeterProvider, set it to global MeterProvider of `otel` package. Whenever you want to get the Meter to push customized metrics, just use the function `metric.Meter()` (refer the example at https://github.com/KyberNetwork/kyber-trace-go/blob/main/example/meter.go for more details).
### Configurations
When initializing the global TracerProvider and global MeterProvider, `kyber-trace-go` loads configurations from the following environment variables: 
  - OTEL_ENABLED: `kyber-trace-go` only initializes global `TracerProvider` and global `MeterProvider` if  `OTEL_ENABLED = true`
  - OTEL_AGENT_HOST: The host of the agent where traces, spans, customized metrics will be sent to. If you are using helm chart with dependency `base-service` version `0.5.16` or later, this environment variable will be injected via helm chart, and you don't need to set it in your `values.yaml` in kyber-application. Otherwise, you have to set it yourself.
  - OTEL_METRIC_AGENT_GRPC_PORT: The gRPC port of the agent where customized metrics will be sent to. If you are using helm chart with dependency `base-service` version `0.5.16` or later, this environment variable will be injected via helm chart, and you don't need to set it in your `values.yaml` in kyber-application. Otherwise, you have to set it yourself.
  - OTEL_METRIC_AGENT_HTTP_PORT: The HTTP port of the agent where customized metrics will be sent to. If you are using helm chart with dependency `base-service` version `0.5.16` or later, this environment variable will be injected via helm chart, and you don't need to set it in your `values.yaml` in kyber-application. Otherwise, you have to set it yourself.
  - OTEL_TRACE_AGENT_GRPC_PORT: The gRPC port of the agent where traces, spans will be sent to. If you are using helm chart with dependency `base-service` version `0.5.16` or later, this environment variable will be injected via helm chart, and you don't need to set it in your `values.yaml` in kyber-application. Otherwise, you have to set it yourself.
  - OTEL_TRACE_AGENT_HTTP_PORT: The HTTP port of the agent where traces, spans will be sent to. If you are using helm chart with dependency `base-service` version `0.5.16` or later, this environment variable will be injected via helm chart, and you don't need to set it in your `values.yaml` in kyber-application. Otherwise, you have to set it yourself.
  - OTEL_INSECURE: Disables client transport security for HTTP/gRPC connection when connecting to agent. If you are using helm chart with dependency `base-service` version `0.5.16` or later, this environment variable will be set to `true` automatically.
  - OTEL_PROTOCOL: Specify which protocol will be used to connect to the agent. Enum: `grpc`, `http`. The default value is `grpc`. Only add this environment variable to your `value.yaml` in kyber-application when you want to use `http`.
  - OTEL_SERVICE_NAME: Name of your service which can be used in your query in grafana jaeger. If you are using helm chart with dependency `base-service` version `0.5.16` or later, this environment variable will be injected via helm chart, and you don't need to set it in your `values.yaml` in `kyber-application`. Otherwise, you have to set it yourself.
  - OTEL_SERVICE_VERSION: The current version of your service. If you are using helm chart with dependency `base-service` version `0.5.16` or later, this environment variable will be injected via helm chart, and you don't need to set it in your `values.yaml` in `kyber-application`. Otherwise, you have to set it to your current image tag.
  - OTEL_TRACE_SAMPLE_RATE: The default value is `0.5`. If you want your all traces and spans will be recorded, set `OTEL_TRACE_SAMPLE_RATE = 1`
  - OTEL_SPAN_METRICS_ENABLED: Opt-in. Set `true` to count every span started in the `kybernetwork.span.started` metric (see [Span started counter](#span-started-counter)). Disabled by default.
  - OTEL_SPAN_METRICS_MAX_SPAN_NAMES: The maximum number of distinct span names reported by `kybernetwork.span.started`. The default value is `1000`. Set `0` for no limit.

### Span started counter
Set `OTEL_SPAN_METRICS_ENABLED=true` (together with `OTEL_ENABLED=true`) to get the counter `kybernetwork.span.started` (shown in Prometheus as `kybernetwork_span_started_total`). It counts every span started through the global `TracerProvider`, including spans from `otelgin`, `otelgorm` and other instrumentation libraries, with the labels:
  - `span.name`: the span name, e.g. `GET /users/:id`
  - `span.kind`: `server`, `client`, `internal`, `producer` or `consumer`

The count is taken before the sampling decision, so it is exact and does not depend on `OTEL_TRACE_SAMPLE_RATE`. Sampling decisions are unchanged. Use it for request/call rates by span name, and use traces for latency and errors.

Notes:
  - The metric is exported by the global `MeterProvider`, so your service must also import `github.com/KyberNetwork/kyber-trace-go/pkg/metric` (the `tools` package does both).
  - Span names must be low-cardinality: put IDs in span attributes, not in the name. After `OTEL_SPAN_METRICS_MAX_SPAN_NAMES` distinct names, new names are reported as `span.name="_other"`. If you see `_other`, look for span names that contain IDs.
  - If the trace exporter cannot be created, spans are still counted but not exported. In that case `kyber-trace-go` registers a `TracerProvider` that drops every span, so `trace.SpanFromContext(ctx).SpanContext()` returns trace IDs that don't exist in the tracing backend.

### For gin framework
If your application is using gin framework, you can use [this package](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin) to trace requests to your application automatically. You can see the example at https://github.com/KyberNetwork/kyber-trace-go/blob/main/example/gin.go

### For GORM
OpenTelemetry GORM is designed is easy to use and provides a simple API for instrumenting GORM applications, making it possible for developers to quickly add observability to their applications without having to write a lot of code. To instrument GORM, you need to install the plugin provided by otelgorm:
```
import (
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
if err != nil {
	panic(err)
}

if err := db.Use(otelgorm.NewPlugin()); err != nil {
	panic(err)
}
```
And then use db.WithContext(ctx) to propagate the active span via context:
```
var num int
if err := db.WithContext(ctx).Raw("SELECT 42").Scan(&num).Error; err != nil {
	panic(err)
}
```
