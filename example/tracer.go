package example

import (
	"context"
	"time"

	"github.com/KyberNetwork/kyber-trace-go/pkg/tracer"
	_ "github.com/KyberNetwork/kyber-trace-go/tools" // this is important
	"go.opentelemetry.io/otel/attribute"
)

func Tracing() {
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

	ctx := context.Background()

	parentSpanCtx, parentSpan := tracer.Tracer().Start(ctx, "parent span")
	parentSpan.SetAttributes(attribute.String("parent_attr", "parent_attr_value"))
	time.Sleep(time.Second)

	_, childSpan := tracer.Tracer().Start(parentSpanCtx, "child span")
	childSpan.SetAttributes(attribute.String("parent_attr", "parent_attr_value"))
	time.Sleep(2 * time.Second)

	childSpan.End()
	parentSpan.End()
	time.Sleep(10 * time.Second) // wait to ensure parentSpan was pushed
}
