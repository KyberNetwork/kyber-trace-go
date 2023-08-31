package example

import (
	"context"

	"github.com/KyberNetwork/kyber-trace-go/pkg/metric"
	_ "github.com/KyberNetwork/kyber-trace-go/tools" // this is important
)

func PushMetric() {
	// Please port forward collector from develop environment to local then set environment variables before run this example:
	// kubectl -n observability port-forward daemonset/opentelemetry-collector-agent 4317:4317 4315:4315
	// export OTEL_ENABLED=true
	// export OTEL_AGENT_HOST=127.0.0.1
	// export OTEL_SERVICE_NAME=your_service_name
	// export OTEL_SERVICE_VERSION=0.1.0
	// export OTEL_TRACE_SAMPLE_RATE=1
	// export OTEL_TRACE_AGENT_GRPC_PORT=4317
	// export OTEL_METRIC_AGENT_GRPC_PORT=4315
	// export OTEL_INSECURE=true

	// When you deploy your service using helm chart with base-service from version 0.5.16, the following variables will be injected directly via helm chart:
	// OTEL_AGENT_HOST, OTEL_SERVICE_NAME, OTEL_SERVICE_VERSION, OTEL_TRACE_AGENT_GRPC_PORT, OTEL_METRIC_AGENT_GRPC_PORT, OTEL_INSECURE
	// You just need to set OTEL_ENABLED.

	counter, err := metric.Meter().Int64Counter("example_count_metric")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	counter.Add(context.Background(), 1)
	err = metric.Flush(ctx)
	if err != nil {
		panic(err)
	}
}
