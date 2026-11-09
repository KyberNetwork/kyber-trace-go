package tracer

import (
	"context"
	"fmt"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/KyberNetwork/kyber-trace-go/pkg/constant"
)

// startSpans starts n spans named name under a fresh TracerProvider using a
// counting sampler around next, and returns the collected counter values keyed
// by "span.name|span.kind" plus how many spans were sampled.
func startSpans(t *testing.T, next sdktrace.Sampler, maxNames int, names []string, kind trace.SpanKind) (map[string]int64, int) {
	t.Helper()

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	sampler, err := newCountingSampler(next, mp.Meter("test"), maxNames)
	if err != nil {
		t.Fatal(err)
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sampler))
	tr := tp.Tracer("test")

	sampled := 0
	for _, name := range names {
		_, span := tr.Start(context.Background(), name, trace.WithSpanKind(kind))
		if span.SpanContext().IsSampled() {
			sampled++
		}
		span.End()
	}

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatal(err)
	}

	got := map[string]int64{}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != constant.OtelSpanStartedMetricName {
				continue
			}
			sum, ok := m.Data.(metricdata.Sum[int64])
			if !ok || !sum.IsMonotonic {
				t.Fatalf("%s is %T, want a monotonic int64 sum", m.Name, m.Data)
			}
			for _, dp := range sum.DataPoints {
				name, _ := dp.Attributes.Value(attribute.Key("span.name"))
				k, _ := dp.Attributes.Value(attribute.Key("span.kind"))
				got[name.AsString()+"|"+k.AsString()] += dp.Value
			}
		}
	}
	return got, sampled
}

func repeat(name string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = name
	}
	return out
}

func TestCountingSamplerIgnoresSamplingRate(t *testing.T) {
	const n = 200
	for _, tc := range []struct {
		name        string
		sampler     sdktrace.Sampler
		wantSampled int
	}{
		{"never", sdktrace.NeverSample(), 0},
		{"ratio 0", sdktrace.TraceIDRatioBased(0), 0},
		{"ratio 0.5", sdktrace.TraceIDRatioBased(0.5), -1},
		{"ratio 1", sdktrace.TraceIDRatioBased(1), n},
		{"parent based", sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.5)), -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, sampled := startSpans(t, tc.sampler, 10, repeat("op", n), trace.SpanKindServer)

			if want := map[string]int64{"op|server": n}; fmt.Sprint(got) != fmt.Sprint(want) {
				t.Errorf("counter = %v, want %v", got, want)
			}
			if tc.wantSampled >= 0 && sampled != tc.wantSampled {
				t.Errorf("sampled %d spans, want %d: the wrapped sampler's decision must be kept", sampled, tc.wantSampled)
			}
		})
	}
}

func TestCountingSamplerKeepsDecision(t *testing.T) {
	// The same trace IDs must get the same decision with and without the wrapper.
	inner := sdktrace.TraceIDRatioBased(0.5)
	reader := sdkmetric.NewManualReader()
	counting, err := newCountingSampler(inner, sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)).Meter("test"), 0)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 256; i++ {
		p := sdktrace.SamplingParameters{
			ParentContext: context.Background(),
			TraceID:       trace.TraceID{byte(i), 0, 0, 0, 0, 0, 0, 0, byte(i), 0xff, 0, 0, 0, 0, 0, 1},
			Name:          "op",
		}
		if a, b := inner.ShouldSample(p).Decision, counting.ShouldSample(p).Decision; a != b {
			t.Fatalf("trace %d: decision %v, want %v", i, b, a)
		}
	}
}

func TestCountingSamplerSpanNameLimit(t *testing.T) {
	names := []string{"a", "b", "c", "d", "a", "e"}
	got, _ := startSpans(t, sdktrace.AlwaysSample(), 2, names, trace.SpanKindInternal)

	want := map[string]int64{
		"a|internal": 2,
		"b|internal": 1,
		constant.OtelSpanMetricsOverflowSpanName + "|internal": 3, // c, d, e
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("counter = %v, want %v", got, want)
	}
}

func TestCountingSamplerNoSpanNameLimit(t *testing.T) {
	names := []string{"a", "b", "c"}
	got, _ := startSpans(t, sdktrace.AlwaysSample(), 0, names, trace.SpanKindClient)

	want := map[string]int64{"a|client": 1, "b|client": 1, "c|client": 1}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("counter = %v, want %v", got, want)
	}
}

func TestWithSpanMetricsIsOptIn(t *testing.T) {
	base := sdktrace.TraceIDRatioBased(0.5)

	t.Setenv(constant.EnvKeyOtelSpanMetricsEnabled, "")
	if got := withSpanMetrics(base); got != base {
		t.Errorf("flag unset: got %s, want the sampler unchanged", got.Description())
	}

	t.Setenv(constant.EnvKeyOtelSpanMetricsEnabled, "false")
	if got := withSpanMetrics(base); got != base {
		t.Errorf("flag false: got %s, want the sampler unchanged", got.Description())
	}

	t.Setenv(constant.EnvKeyOtelSpanMetricsEnabled, "true")
	got := withSpanMetrics(base)
	if _, ok := got.(*countingSampler); !ok {
		t.Fatalf("flag true: got %T, want *countingSampler", got)
	}
	if want := "CountingSampler{" + base.Description() + "}"; got.Description() != want {
		t.Errorf("Description() = %q, want %q", got.Description(), want)
	}
}

func TestCountOnlyProvider(t *testing.T) {
	prevTP, prevMP, prevProp := otel.GetTracerProvider(), otel.GetMeterProvider(), otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
		otel.SetMeterProvider(prevMP)
		otel.SetTextMapPropagator(prevProp)
	})

	reader := sdkmetric.NewManualReader()
	otel.SetMeterProvider(sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)))

	t.Setenv(constant.EnvKeyOtelSpanMetricsEnabled, "")
	initCountOnlyProvider()
	if otel.GetTracerProvider() != prevTP {
		t.Fatal("flag unset: the global TracerProvider must not change")
	}

	t.Setenv(constant.EnvKeyOtelSpanMetricsEnabled, "true")
	initCountOnlyProvider()
	if Provider() != nil {
		t.Error("Provider() must stay nil so Flush/Shutdown and retries behave as before")
	}
	if fields := otel.GetTextMapPropagator().Fields(); len(fields) != len(prevProp.Fields()) {
		t.Errorf("propagator changed to one with fields %v", fields)
	}

	for i := 0; i < 3; i++ {
		_, span := Tracer().Start(context.Background(), "op")
		if span.IsRecording() || span.SpanContext().IsSampled() {
			t.Error("count-only provider must not record or sample spans")
		}
		span.End()
	}

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatal(err)
	}
	var total int64
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == constant.OtelSpanStartedMetricName {
				for _, dp := range m.Data.(metricdata.Sum[int64]).DataPoints {
					total += dp.Value
				}
			}
		}
	}
	if total != 3 {
		t.Errorf("counted %d spans, want 3", total)
	}
}

func TestCountingSamplerDefaultsSpanKindToInternal(t *testing.T) {
	got, _ := startSpans(t, sdktrace.AlwaysSample(), 0, []string{"op"}, trace.SpanKindUnspecified)

	if want := map[string]int64{"op|internal": 1}; fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("counter = %v, want %v", got, want)
	}
}
