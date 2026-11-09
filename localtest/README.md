# Local test stack

Runs an OpenTelemetry collector, Jaeger and Prometheus with Docker, plus a small gin service that uses this library, so you can check `kybernetwork.span.started` end to end.

The collector listens on the same agent ports as Kyber clusters (traces gRPC `4317`, metrics gRPC `4315`). Its `spanmetrics` connector counts only spans that reach the collector, i.e. sampled spans, which gives a baseline to compare against.

## Run

From the repo root:

```bash
docker compose -f localtest/docker-compose.yml up -d
bash localtest/run-demo.sh        # demo on :18080, sample rate 0.5, span metrics on
bash localtest/load.sh 500        # in another terminal
```

`run-demo.sh` reads `OTEL_TRACE_SAMPLE_RATE` and `OTEL_SPAN_METRICS_ENABLED` from the environment if they're set, for example `OTEL_SPAN_METRICS_ENABLED=false bash localtest/run-demo.sh` to check the flag-off behaviour.

## Check

| What | Where |
|---|---|
| Collector metrics | http://localhost:8889/metrics |
| Prometheus | http://localhost:9091 |
| Jaeger | http://localhost:16689 |

Compare in Prometheus:

- `kybernetwork_span_started_total`: every span started, so it should match the number of requests sent.
- `traces_span_metrics_calls_total`: sampled spans only, about `OTEL_TRACE_SAMPLE_RATE` × requests. It can lag behind for up to a minute.

## Stop

Stop the demo with Ctrl+C, then:

```bash
docker compose -f localtest/docker-compose.yml down
```
