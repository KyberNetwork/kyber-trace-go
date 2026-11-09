// Demo service for testing kybernetwork.span.started locally.
// Run it with localtest/run-demo.sh.
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/KyberNetwork/kyber-trace-go/pkg/tracer"
	_ "github.com/KyberNetwork/kyber-trace-go/tools"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	// Library span: one server span per request, named by route template.
	r.Use(otelgin.Middleware("span-counter-demo"))

	r.GET("/api/v1/greeting", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})

	r.GET("/api/v1/users/:id", func(c *gin.Context) {
		// Hand-made child span, as a service would add around a repository call.
		_, span := tracer.Tracer().Start(c.Request.Context(), "UserRepo.FindByID")
		time.Sleep(2 * time.Millisecond)
		span.End()
		c.JSON(http.StatusOK, gin.H{"id": c.Param("id")})
	})

	log.Println("demo listening on :18080")
	log.Fatal(r.Run(":18080"))
}
