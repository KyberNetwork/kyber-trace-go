package example

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"

	"github.com/KyberNetwork/kyber-trace-go/pkg/tracer"
	_ "github.com/KyberNetwork/kyber-trace-go/tools" // this is important
)

func OBI() {
	// Run this example together with OpenTelemetry eBPF Instrumentation (OBI), see deploy/obi/README.md:
	// export OTEL_ENABLED=true
	// export OTEL_TRACE_MODE=obi
	// export OTEL_SERVICE_NAME=your_service_name

	// In OBI mode, no middleware is needed: OBI captures the incoming HTTP request span from outside the process
	// and propagates trace context into the Go process. Manual spans created with tracer.Tracer() are picked up by
	// OBI via the Go Auto SDK and attached as children of the request span.
	server := gin.New()
	server.GET("/api/v1/greeting", func(c *gin.Context) {
		_, span := tracer.Tracer().Start(c.Request.Context(), "build greeting")
		span.SetAttributes(attribute.String("greeting.lang", "en"))
		time.Sleep(10 * time.Millisecond)
		span.End()

		c.AbortWithStatusJSON(http.StatusOK, "Hello World")
	})
	if err := server.Run(); err != nil {
		panic(err)
	}
}
