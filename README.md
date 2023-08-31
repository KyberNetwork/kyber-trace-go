# Kyber Trace GO Lib

## Overview
This is a library that makes it easier to integrate with Kyber's self-hosted open tracing

## Add this lib to your project
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
In order to push spans and traces to Kyber's self-hosted agent, you have to initialize a TracerProvider. When you add the following statement into your code: `import _ "github.com/KyberNetwork/kyber-trace-go/pkg/tracer"`, `kyber-trace-go` will initialize a new TraceProvider, set it to global TracerProvider of `otel` package. Whenever you want to get the Tracer to start (or get) a span, just use the function `tracer.Tracer()` (refer the example at https://github.com/KyberNetwork/kyber-trace-go/blob/main/example/tracer.go for more details).
### Init global `MeterProvider`
In order to push customized metrics to Kyber's self-hosted agent, you have to initialize a MeterProvider. When you add the following statement into your code: `import _ "github.com/KyberNetwork/kyber-trace-go/pkg/metric"`, `kyber-trace-go` will initialize a new MeterProvider, set it to global MeterProvider of `otel` package. Whenever you want to get the Meter to push customized metrics, just use the function `metric.Meter()` (refer the example at https://github.com/KyberNetwork/kyber-trace-go/blob/main/example/meter.go for more details).
### Configurations
When initializing the global TracerProvider and global MeterProvider, `kyber-trace-go` loads configurations from the following environment variables: 
  - OTEL_ENABLED: `kyber-trace-go` only initializes global `TracerProvider` and global `MeterProvider` if  `OTEL_ENABLED = true`
  - OTEL_AGENT_HOST: The host of the agent where traces, spans, customized metrics will be sent to. If you are using helm chart with dependency `base-service` version `0.5.15` or later, this environment variable will be injected via helm chart, and you don't need to set it in your `values.yaml` in kyber-application. Otherwise, you have to set it yourself.
  - OTEL_METRIC_AGENT_GRPC_PORT: The gRPC port of the agent where customized metrics will be sent to. If you are using helm chart with dependency `base-service` version `0.5.15` or later, this environment variable will be injected via helm chart, and you don't need to set it in your `values.yaml` in kyber-application. Otherwise, you have to set it yourself.
  - OTEL_METRIC_AGENT_HTTP_PORT: The HTTP port of the agent where customized metrics will be sent to. If you are using helm chart with dependency `base-service` version `0.5.15` or later, this environment variable will be injected via helm chart, and you don't need to set it in your `values.yaml` in kyber-application. Otherwise, you have to set it yourself.
  - OTEL_TRACE_AGENT_GRPC_PORT: The gRPC port of the agent where traces, spans will be sent to. If you are using helm chart with dependency `base-service` version `0.5.15` or later, this environment variable will be injected via helm chart, and you don't need to set it in your `values.yaml` in kyber-application. Otherwise, you have to set it yourself.
  - OTEL_TRACE_AGENT_HTTP_PORT: The HTTP port of the agent where traces, spans will be sent to. If you are using helm chart with dependency `base-service` version `0.5.15` or later, this environment variable will be injected via helm chart, and you don't need to set it in your `values.yaml` in kyber-application. Otherwise, you have to set it yourself.
  - OTEL_INSECURE: Disables client transport security for HTTP/gRPC connection when connecting to agent. If you are using helm chart with dependency `base-service` version `0.5.15` or later, this environment variable will be set to `true` automatically.
  - OTEL_PROTOCOL: Specify which protocol will be used to connect to the agent. Enum: `grpc`, `http`. The default value is `grpc`. Only add this environment variable to your `value.yaml` in kyber-application when you want to use `http`.
  - OTEL_SERVICE_NAME: Name of your service which can be used in your query in grafana jaeger. If you are using helm chart with dependency `base-service` version `0.5.15` or later, this environment variable will be injected via helm chart, and you don't need to set it in your `values.yaml` in `kyber-application`. Otherwise, you have to set it yourself.
  - OTEL_SERVICE_VERSION: The current version of your service. If you are using helm chart with dependency `base-service` version `0.5.15` or later, this environment variable will be injected via helm chart, and you don't need to set it in your `values.yaml` in `kyber-application`. Otherwise, you have to set it to your current image tag.
  - OTEL_TRACE_SAMPLE_RATE: The default value is `0.5`. If you want your all traces and spans will be recorded, set `OTEL_TRACE_SAMPLE_RATE = 1`

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
