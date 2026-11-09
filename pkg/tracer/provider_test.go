package tracer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/KyberNetwork/kyber-trace-go/pkg/constant"
)

func TestIsOBIMode(t *testing.T) {
	t.Setenv(constant.EnvKeyOtelTraceMode, "")
	assert.False(t, IsOBIMode())

	t.Setenv(constant.EnvKeyOtelTraceMode, "sdk")
	assert.False(t, IsOBIMode())

	t.Setenv(constant.EnvKeyOtelTraceMode, "obi")
	assert.True(t, IsOBIMode())

	t.Setenv(constant.EnvKeyOtelTraceMode, " OBI ")
	assert.True(t, IsOBIMode())
}

func TestInitProviderOBIMode(t *testing.T) {
	t.Setenv(constant.EnvKeyOtelTraceMode, "obi")

	InitProvider()

	assert.Nil(t, Provider())
	// the global TracerProvider must stay the delegating one so OBI can activate the Go Auto SDK
	_, isSDK := otel.GetTracerProvider().(*sdktrace.TracerProvider)
	assert.False(t, isSDK)
	assert.NoError(t, Flush(context.Background()))
	assert.NoError(t, Shutdown(context.Background()))
}
