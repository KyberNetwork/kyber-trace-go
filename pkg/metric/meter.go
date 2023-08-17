package metric

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"

	"github.com/KyberNetwork/kyber-trace-go/pkg/constant"
	"github.com/KyberNetwork/kyber-trace-go/pkg/util/env"
)

func Meter() metric.Meter {
	return otel.GetMeterProvider().Meter(env.StringFromEnv(constant.EnvKeyOTLPServiceName, constant.OTLPDefaultServiceName))
}

func Flush(ctx context.Context) error {
	return otel.GetMeterProvider().(*metricsdk.MeterProvider).ForceFlush(ctx)
}
