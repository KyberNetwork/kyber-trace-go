#!/usr/bin/env bash
# Starts the demo service against the local collector.
# Usage: localtest/run-demo.sh   (from the repo root)
export OTEL_ENABLED=true
export OTEL_AGENT_HOST=127.0.0.1
export OTEL_INSECURE=true
export OTEL_SERVICE_NAME=span-counter-demo
export OTEL_SERVICE_VERSION=0.1.0
export OTEL_TRACE_AGENT_GRPC_PORT=4317
export OTEL_METRIC_AGENT_GRPC_PORT=4315
export OTEL_TRACE_SAMPLE_RATE=${OTEL_TRACE_SAMPLE_RATE:-0.5}
export OTEL_SPAN_METRICS_ENABLED=${OTEL_SPAN_METRICS_ENABLED:-true}
export OTEL_METRIC_EXPORT_INTERVAL=5000 # ms; default is 60s
exec go run ./localtest/demo
