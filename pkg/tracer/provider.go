package tracer

import (
	"context"
	"fmt"
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

func newGRPCExporter(ctx context.Context, providerServerUrl string, isInsecure bool) (*otlptrace.Exporter, error) {
	clientOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(providerServerUrl),
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

func newHTTPExporter(ctx context.Context, providerServerUrl string, isInsecure bool) (*otlptrace.Exporter, error) {
	clientOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(providerServerUrl),
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
	providerServerUrl := env.StringFromEnv(constant.EnvKeyOTLPCollectorUrl, "")
	isInsecure := env.BoolFromEnv(constant.EnvKeyOTLPInsecure)
	protocol := env.StringFromEnv(constant.EnvKeyOTLPProtocol, constant.OTLPProtocolGRPC)

	// gRPC
	if protocol == constant.OTLPProtocolGRPC {
		return newGRPCExporter(ctx, providerServerUrl, isInsecure)
	}

	// HTTP
	return newHTTPExporter(ctx, providerServerUrl, isInsecure)
}

func newResources() *resource.Resource {
	// Ensure default SDK resources and the required service name are set.
	// ref: https://opentelemetry.io/docs/instrumentation/go/resources/
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(env.StringFromEnv(constant.EnvKeyOTLPServiceName, constant.OTLPDefaultServiceName)),
		semconv.ServiceVersion(env.StringFromEnv(constant.EnvKeyOTLPServiceVersion, constant.OTLPDefaultServiceVersion)),
	)
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

	sampleRate := env.FloatFromEnv(constant.EnvKeyOTLPTraceSampleRate, constant.OTLPDefaultSampleRate)

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
