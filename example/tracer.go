package example

import (
	"context"
	"time"

	"github.com/KyberNetwork/kyber-trace-go/pkg/tracer"
	_ "github.com/KyberNetwork/kyber-trace-go/tools" // this is important
	"go.opentelemetry.io/otel/attribute"
)

func Tracing() {
	ctx := context.Background()

	parentSpanCtx, parentSpan := tracer.Tracer().Start(ctx, "parent span")
	parentSpan.SetAttributes(attribute.String("parent_attr", "parent_attr_value"))
	time.Sleep(time.Second)

	_, childSpan := tracer.Tracer().Start(parentSpanCtx, "child span")
	childSpan.SetAttributes(attribute.String("parent_attr", "parent_attr_value"))
	time.Sleep(2 * time.Second)

	childSpan.End()
	parentSpan.End()
	time.Sleep(5 * time.Second) // wait to ensure parentSpan was pushed
}
