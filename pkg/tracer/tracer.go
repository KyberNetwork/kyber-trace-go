package tracer

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/KyberNetwork/kyber-trace-go/pkg/constant"
	"github.com/KyberNetwork/kyber-trace-go/pkg/util/env"
)

func Tracer() trace.Tracer {
	return otel.Tracer(env.StringFromEnv(constant.EnvKeyOtelServiceName, constant.OtelDefaultServiceName))
}
