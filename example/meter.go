package example

import (
	"context"
	"time"

	"github.com/KyberNetwork/kyber-trace-go/pkg/metric"
)

func PushMetric() {
	counter, err := metric.Meter().Int64Counter("example_count")
	if err != nil {
		panic(err)
	}
	counter.Add(context.Background(), 1)
	time.Sleep(10 * time.Second)
}
