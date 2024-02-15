package constant

const (
	EnvKeyOtelEnabled                             = "OTEL_ENABLED"
	EnvKeyOtelAgentHost                           = "OTEL_AGENT_HOST"
	EnvKeyOtelMetricAgentGRPCPort                 = "OTEL_METRIC_AGENT_GRPC_PORT"
	EnvKeyOtelMetricAgentHTTPPort                 = "OTEL_METRIC_AGENT_HTTP_PORT"
	EnvKeyOtelTraceAgentGRPCPort                  = "OTEL_TRACE_AGENT_GRPC_PORT"
	EnvKeyOtelTraceAgentHTTPPort                  = "OTEL_TRACE_AGENT_HTTP_PORT"
	EnvKeyOtelInsecure                            = "OTEL_INSECURE"
	EnvKeyOtelProtocol                            = "OTEL_PROTOCOL"
	EnvKeyOtelServiceName                         = "OTEL_SERVICE_NAME"
	EnvKeyOtelServiceVersion                      = "OTEL_SERVICE_VERSION"
	EnvKeyOtelTraceSampleRate                     = "OTEL_TRACE_SAMPLE_RATE"
	EnvKeyOtelEnabledExponentialHistogramMetrics  = "OTEL_ENABLED_EXPONENTIAL_HISTOGRAM_METRICS"
	EnvKeyOtelExponentialHistogramMetricsMaxScale = "OTEL_EXPONENTIAL_HISTOGRAM_METRICS_MAX_SCALE"
	EnvKeyOtelExponentialHistogramMetricsMaxSize  = "OTEL_EXPONENTIAL_HISTOGRAM_METRICS_MAX_SIZE"
)
