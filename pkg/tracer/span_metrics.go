package tracer

import (
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/KyberNetwork/kyber-trace-go/pkg/constant"
)

const meterScopeName = "github.com/KyberNetwork/kyber-trace-go/pkg/tracer"

// countingSampler wraps another Sampler and counts every span start in the
// kybernetwork.span.started counter.
//
// The SDK calls ShouldSample for every span before it decides whether to keep
// it, so the count is exact and independent of the sampling rate. The wrapped
// sampler's decision is returned unchanged.
type countingSampler struct {
	next    sdktrace.Sampler
	counter metric.Int64Counter
	names   *spanNameLimiter
}

func newCountingSampler(next sdktrace.Sampler, meter metric.Meter, maxSpanNames int) (*countingSampler, error) {
	counter, err := meter.Int64Counter(
		constant.OtelSpanStartedMetricName,
		metric.WithDescription("Number of spans started, counted before sampling."),
		metric.WithUnit("{span}"),
	)
	if err != nil {
		return nil, err
	}
	return &countingSampler{
		next:    next,
		counter: counter,
		names:   newSpanNameLimiter(maxSpanNames),
	}, nil
}

func (s *countingSampler) ShouldSample(p sdktrace.SamplingParameters) sdktrace.SamplingResult {
	res := s.next.ShouldSample(p)
	s.counter.Add(p.ParentContext, 1, metric.WithAttributes(
		attribute.String("span.name", s.names.get(p.Name)),
		// Match the kind the SDK records: an unset kind becomes internal.
		attribute.String("span.kind", trace.ValidateSpanKind(p.Kind).String()),
	))
	return res
}

func (s *countingSampler) Description() string {
	return "CountingSampler{" + s.next.Description() + "}"
}

// spanNameLimiter caps how many distinct span names are reported. Once the cap
// is reached, new names are reported as OtelSpanMetricsOverflowSpanName so a
// span name containing IDs cannot create unbounded metric series.
type spanNameLimiter struct {
	max  int
	mu   sync.RWMutex
	seen map[string]struct{}
}

func newSpanNameLimiter(max int) *spanNameLimiter {
	return &spanNameLimiter{max: max, seen: make(map[string]struct{})}
}

func (l *spanNameLimiter) get(name string) string {
	if l.max <= 0 {
		return name
	}

	l.mu.RLock()
	_, ok := l.seen[name]
	full := len(l.seen) >= l.max
	l.mu.RUnlock()
	if ok {
		return name
	}
	if full {
		return constant.OtelSpanMetricsOverflowSpanName
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.seen[name]; ok {
		return name
	}
	if len(l.seen) >= l.max {
		return constant.OtelSpanMetricsOverflowSpanName
	}
	l.seen[name] = struct{}{}
	return name
}
