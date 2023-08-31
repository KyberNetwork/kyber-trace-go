package example

import (
	"net/http"
	"os"

	_ "github.com/KyberNetwork/kyber-trace-go/tools" // this is important
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func GinFramework() {
	// Please port forward collector from develop environment to local then set environment variables before run this example:
	// kubectl -n observability port-forward daemonset/opentelemetry-collector-agent 4317:4317 4315:4315
	// export OTEL_ENABLED=true
	// export OTEL_AGENT_HOST=127.0.0.1
	// export OTEL_SERVICE_NAME=your_service_name
	// export OTEL_SERVICE_VERSION=0.1.0
	// export OTEL_TRACE_SAMPLE_RATE=1
	// export OTEL_TRACE_AGENT_GRPC_PORT=4317
	// export OTEL_METRIC_AGENT_GRPC_PORT=4315
	// export OTEL_INSECURE=true

	// When you deploy your service using helm chart with base-service from version 0.5.16, the following variables will be injected directly via helm chart:
	// OTEL_AGENT_HOST, OTEL_SERVICE_NAME, OTEL_SERVICE_VERSION, OTEL_TRACE_AGENT_GRPC_PORT, OTEL_METRIC_AGENT_GRPC_PORT, OTEL_INSECURE
	// You just need to set OTEL_ENABLED.

	server := newServer()
	err := server.Run()
	if err != nil {
		panic(err)
	}
}

func newServer() *gin.Engine {
	server := gin.New()
	if os.Getenv("OTEL_ENABLED") == "true" {
		server.Use(otelgin.Middleware(
			os.Getenv("OTEL_SERVICE_NAME"),
		))
	}
	rg := server.Group("/api/v1")
	rg.GET("/greeting", func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusOK, "Hello World")
	})
	return server
}
