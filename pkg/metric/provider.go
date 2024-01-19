package metric

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc"

	"github.com/KyberNetwork/kyber-trace-go/pkg/constant"
	"github.com/KyberNetwork/kyber-trace-go/pkg/util/env"
)

var provider *metric.MeterProvider
var lock sync.Mutex

func newGRPCExporter(ctx context.Context, agentHost string, isInsecure bool) (metric.Exporter, error) {
	addr := net.JoinHostPort(agentHost, env.StringFromEnv(
		constant.EnvKeyOtelMetricAgentGRPCPort, constant.OtelDefaultMetricAgentGRPCPort))
	clientOpts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(addr),
		otlpmetricgrpc.WithDialOption(grpc.WithBlock()),
	}
	if isInsecure {
		clientOpts = append(clientOpts, otlpmetricgrpc.WithInsecure())
	}
	exporter, err := otlpmetricgrpc.New(ctx, clientOpts...)
	if err != nil {
		return nil, err
	}
	return exporter, nil
}

func newHTTPExporter(ctx context.Context, agentHost string, isInsecure bool) (metric.Exporter, error) {
	addr := net.JoinHostPort(agentHost, env.StringFromEnv(
		constant.EnvKeyOtelMetricAgentHTTPPort, constant.OtelDefaultMetricAgentHTTPPort))
	clientOpts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(addr),
	}
	if isInsecure {
		clientOpts = append(clientOpts, otlpmetrichttp.WithInsecure())
	}
	exporter, err := otlpmetrichttp.New(ctx, clientOpts...)
	if err != nil {
		return nil, err
	}
	return exporter, nil
}

func newOTLPExporter() (metric.Exporter, error) {
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

	metricsView := make([]metric.View, 0)

	exporter, err := newOTLPExporter()
	if err != nil {
		fmt.Printf("kyber-trace-go: failed to init metric provider, %s\n", err)
		return
	}

	if constant.EnvKeyOtelEnabledExponentialHistogramMetrics == "true" {
		exponentialHistogramView := metric.NewView(
			metric.Instrument{
				Kind: metric.InstrumentKindHistogram,
			}, metric.Stream{
				Aggregation: metric.AggregationBase2ExponentialHistogram{
					MaxSize:  30,
					MaxScale: 3,
				},
			})
		metricsView = append(metricsView, exponentialHistogramView)
	}

	provider = metric.NewMeterProvider(
		metric.WithResource(newResources()),
		metric.WithReader(metric.NewPeriodicReader(exporter)),
		metric.WithView(metricsView...),
	)

	otel.SetMeterProvider(provider)
}

func Provider() *metric.MeterProvider {
	return provider
}
