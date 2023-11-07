package tracer

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"google.golang.org/grpc"

	"github.com/KyberNetwork/kyber-trace-go/pkg/constant"
	"github.com/KyberNetwork/kyber-trace-go/pkg/util/env"
)

var provider *trace.TracerProvider
var lock sync.Mutex

func newGRPCExporter(ctx context.Context, agentHost string, isInsecure bool) (*otlptrace.Exporter, error) {
	addr := net.JoinHostPort(agentHost, env.StringFromEnv(
		constant.EnvKeyOtelTraceAgentGRPCPort, constant.OtelDefaultTraceAgentGRPCPort))
	clientOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(addr),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	}
	if isInsecure {
		clientOpts = append(clientOpts, otlptracegrpc.WithInsecure())
	}
	exporter, err := otlptrace.New(
		ctx, otlptracegrpc.NewClient(clientOpts...),
	)
	if err != nil {
		return nil, err
	}
	return exporter, nil
}

func newHTTPExporter(ctx context.Context, agentHost string, isInsecure bool) (*otlptrace.Exporter, error) {
	addr := net.JoinHostPort(agentHost, env.StringFromEnv(
		constant.EnvKeyOtelTraceAgentHTTPPort, constant.OtelDefaultTraceAgentHTTPPort))
	clientOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(addr),
	}
	if isInsecure {
		clientOpts = append(clientOpts, otlptracehttp.WithInsecure())
	}
	exporter, err := otlptrace.New(
		ctx, otlptracehttp.NewClient(clientOpts...),
	)
	if err != nil {
		return nil, err
	}
	return exporter, nil
}

func newOTLPExporter() (*otlptrace.Exporter, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	agentHost := env.StringFromEnv(constant.EnvKeyOtelAgentHost, "")
	isInsecure := env.BoolFromEnv(constant.EnvKeyOtelInsecure)
	protocol := env.StringFromEnv(constant.EnvKeyOtelProtocol, constant.OtelProtocolGRPC)

	// gRPC
	if protocol == constant.OtelProtocolGRPC {
		return newGRPCExporter(ctx, agentHost, isInsecure)
	}

	// HTTP
	return newHTTPExporter(ctx, agentHost, isInsecure)
}

func newResources() *resource.Resource {
	// Ensure default SDK resources and the required service name are set.
	// ref: https://opentelemetry.io/docs/instrumentation/go/resources/
	// default resources
	resources := resource.Default()

	// adding extra resources
	extraResources, err := resource.New(context.Background(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceName(env.StringFromEnv(constant.EnvKeyOtelServiceName, constant.OtelDefaultServiceName)),
			semconv.ServiceVersion(env.StringFromEnv(constant.EnvKeyOtelServiceVersion, constant.OtelDefaultServiceVersion)),
		))
	if err != nil {
		return resources
	}

	// merge default and extra resources
	resources, err = resource.Merge(resources, extraResources)
	if err != nil {
		return resources
	}

	return resources
}

func InitProvider() {
	lock.Lock()
	defer lock.Unlock()

	if provider != nil {
		return
	}

	exporter, err := newOTLPExporter()
	if err != nil {
		fmt.Printf("kyber-trace-go: failed to init tracer provider, %s\n", err)
		return
	}

	// Register the trace exporter with a TracerProvider, using a batch span processor to aggregate spans before export.
	bsp := trace.NewBatchSpanProcessor(exporter)

	sampleRate := env.FloatFromEnv(constant.EnvKeyOtelTraceSampleRate, constant.OtelDefaultSampleRate)

	// init tracer provider
	provider = trace.NewTracerProvider(
		trace.WithSampler(trace.TraceIDRatioBased(sampleRate)),
		trace.WithResource(newResources()),
		trace.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
}

func Provider() *trace.TracerProvider {
	return provider
}
