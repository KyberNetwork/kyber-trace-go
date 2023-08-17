package example

import (
	"context"

	"github.com/KyberNetwork/kyber-trace-go/pkg/metric"
	_ "github.com/KyberNetwork/kyber-trace-go/tools" // this is important
)

func PushMetric() {
	counter, err := metric.Meter().Int64Counter("example_count_edge_4")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	counter.Add(context.Background(), 1)
	err = metric.Flush(ctx)
	if err != nil {
		panic(err)
	}
}
