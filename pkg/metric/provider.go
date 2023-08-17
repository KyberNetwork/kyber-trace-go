package metric

import (
	"context"
	"fmt"
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

func newGRPCExporter(ctx context.Context, providerServerUrl string, isInsecure bool) (metric.Exporter, error) {
	clientOpts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(providerServerUrl),
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

func newHTTPExporter(ctx context.Context, providerServerUrl string, isInsecure bool) (metric.Exporter, error) {
	clientOpts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(providerServerUrl),
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
		fmt.Printf("kyber-trace-go: failed to init metric provider, %s\n", err)
		return
	}

	provider = metric.NewMeterProvider(
		metric.WithResource(newResources()),
		metric.WithReader(metric.NewPeriodicReader(exporter)),
	)

	otel.SetMeterProvider(provider)
}

func Provider() *metric.MeterProvider {
	return provider
}
