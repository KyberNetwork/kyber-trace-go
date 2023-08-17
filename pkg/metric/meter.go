package metric

import (
	"context"
	"errors"

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
	if m, ok := otel.GetMeterProvider().(*metricsdk.MeterProvider); ok {
		return m.ForceFlush(ctx)
	} else {
		return errors.New("no meter provider was initialized")
	}
}
