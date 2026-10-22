package tracerv2

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/KyberNetwork/kyber-trace-go/pkg/constant"
	kybermetric "github.com/KyberNetwork/kyber-trace-go/pkg/metric"
	"github.com/KyberNetwork/kyber-trace-go/pkg/util/env"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	instrumentsOnce sync.Once
	callsCounter    metric.Int64Counter
	durationHist    metric.Float64Histogram
	recorderCh      chan *recorderEvent
	recorderOnce    sync.Once
)

type recorderEvent struct {
	elapsed float64
	attrs   metric.MeasurementOption
}

func initInstruments() bool {
	kybermetric.InitProvider()
	if otel.GetMeterProvider() == nil {
		return false
	}
	m := kybermetric.Meter()

	var err error
	callsCounter, err = m.Int64Counter(
		"kybernetwork.span.metrics.calls.total",
		metric.WithUnit("{call}"),
		metric.WithDescription("Number of span calls"),
	)
	if err != nil {
		log.Printf("tracer: failed to create calls counter: %s", err)

		return false
	}

	durationHist, err = m.Float64Histogram(
		"kybernetwork.span.metrics.duration.milliseconds",
		metric.WithUnit("ms"),
		metric.WithDescription("Span duration in milliseconds"),
	)
	if err != nil {
		log.Printf("tracer: failed to create duration histogram: %s", err)

		return false
	}

	return true
}

func ensureRecorder() {
	recorderOnce.Do(func() {
		recorderCh = make(chan *recorderEvent, 1024)
		go func() {
			for ev := range recorderCh {
				instrumentsOnce.Do(func() {
					if !initInstruments() {
						instrumentsOnce = sync.Once{}
					}
				})

				if callsCounter == nil || durationHist == nil {
					continue
				}

				callsCounter.Add(context.Background(), 1, ev.attrs)
				durationHist.Record(context.Background(), ev.elapsed, ev.attrs)
			}
		}()
	})
}

type Span struct {
	operationName string
	startTime     time.Time
	tags          []attribute.KeyValue
}

func (_self *Span) SetTag(name string, value string) {
	_self.tags = append(_self.tags, attribute.String(name, value))
}

func (_self *Span) End() {
	ensureRecorder()

	elapsed := float64(time.Since(_self.startTime).Milliseconds())
	serviceVersion := env.StringFromEnv(constant.EnvKeyOtelServiceVersion, constant.OtelDefaultServiceVersion)
	kvs := make([]attribute.KeyValue, 0, 2+len(_self.tags))
	kvs = append(kvs, attribute.String("service_version", serviceVersion))
	kvs = append(kvs, attribute.String("span_name", _self.operationName))
	kvs = append(kvs, _self.tags...)

	select {
	case recorderCh <- &recorderEvent{
		elapsed: elapsed,
		attrs:   metric.WithAttributes(kvs...),
	}:
	default:
	}
}

func StartSpanFromContext(ctx context.Context, operationName string) (*Span, context.Context) {
	return &Span{
		operationName: operationName,
		startTime:     time.Now(),
	}, ctx
}
