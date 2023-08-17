package tracer

import (
	"github.com/KyberNetwork/kyber-trace-go/pkg/constant"
	"github.com/KyberNetwork/kyber-trace-go/pkg/util/env"
)

func init() {
	if env.BoolFromEnv(constant.EnvKeyOTLPEnable) {
		initProvider()
	}
}
