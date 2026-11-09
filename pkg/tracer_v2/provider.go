package tracerv2

// InitProvider is a no-op retained for interface compatibility.
// Tracing has been replaced by metric-only aggregation;
// metric instruments are initialized lazily in StartSpanFromContext.
func InitProvider() {}
