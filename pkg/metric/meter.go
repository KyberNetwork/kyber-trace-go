package metric

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"github.com/KyberNetwork/kyber-trace-go/pkg/constant"
	"github.com/KyberNetwork/kyber-trace-go/pkg/util/env"
)

func Meter() metric.Meter {
	return otel.GetMeterProvider().Meter(env.StringFromEnv(constant.EnvKeyOTLPServiceName, constant.OTLPDefaultServiceName))
}
