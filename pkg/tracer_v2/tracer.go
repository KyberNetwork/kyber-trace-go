package tracerv2

import (
	"context"
	"log"
	"sync"
	"time"

	kybermetric "github.com/KyberNetwork/kyber-trace-go/pkg/metric"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

var (
	instrumentsOnce sync.Once
	callsCounter    metric.Int64Counter
	durationHist    metric.Float64Histogram
	recordCh        chan *recordEvent
	recordOnce      sync.Once
	resourceAttrs   []attribute.KeyValue
	resourceOnce    sync.Once
)

type recordEvent struct {
	elapsed float64
	attrs   metric.MeasurementOption
}

func ensureInstruments() {
	instrumentsOnce.Do(func() {
		m := kybermetric.Meter()

		var err error
		callsCounter, err = m.Int64Counter(
			"kybernetwork.span.metrics.calls.total",
			metric.WithUnit("{call}"),
			metric.WithDescription("Number of span calls"),
		)
		if err != nil {
			log.Printf("tracer: failed to create calls counter: %s", err)
		}

		durationHist, err = m.Float64Histogram(
			"kybernetwork.span.metrics.duration.milliseconds",
			metric.WithUnit("ms"),
			metric.WithDescription("Span duration in milliseconds"),
		)
		if err != nil {
			log.Printf("tracer: failed to create duration histogram: %s", err)
		}
	})
}

func ensureRecordLoop() {
	recordOnce.Do(func() {
		recordCh = make(chan *recordEvent, 1024)
		go func() {
			for ev := range recordCh {
				if callsCounter != nil {
					callsCounter.Add(context.Background(), 1, ev.attrs)
				}

				if durationHist != nil {
					durationHist.Record(context.Background(), ev.elapsed, ev.attrs)
				}
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
	ensureInstruments()
	ensureRecordLoop()

	resourceOnce.Do(func() {
		resourceAttrs = resource.Default().Attributes()
		log.Printf("tracer: resource attributes: %v", resourceAttrs)
	})

	elapsed := float64(time.Since(_self.startTime).Milliseconds())
	kvs := make([]attribute.KeyValue, 0, 1+len(resourceAttrs)+len(_self.tags))
	kvs = append(kvs, attribute.String("span_name", _self.operationName))
	kvs = append(kvs, resourceAttrs...)
	kvs = append(kvs, _self.tags...)

	select {
	case recordCh <- &recordEvent{
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
