# OBI (OpenTelemetry eBPF Instrumentation)

[OBI](https://opentelemetry.io/docs/zero-code/obi/) is the runtime, zero-code approach described in
[Choosing Go auto-instrumentation](https://opentelemetry.io/docs/zero-code/go/choosing-auto-instrumentation/).
It runs as a separate agent and uses eBPF to capture spans and RED metrics at the network boundary
(incoming/outgoing HTTP/S, gRPC, SQL, Redis, Kafka...) without rebuilding or restarting the service.

`kyber-trace-go` supports it through `OTEL_TRACE_MODE=obi`:

| | `OTEL_TRACE_MODE=sdk` (default) | `OTEL_TRACE_MODE=obi` |
| --- | --- | --- |
| Request spans (HTTP/gRPC/SQL) | needs middleware / plugins (`otelgin`, `otelgorm`, ...) | captured by OBI, no code needed |
| Manual spans from `tracer.Tracer()` | exported by the in-process SDK | captured by OBI via the [Go Auto SDK](https://opentelemetry.io/docs/zero-code/go/autosdk/) |
| Sampling | `OTEL_TRACE_SAMPLE_RATE` | OBI config (`otel_traces_export.sampler`) |
| Custom metrics from `metric.Meter()` | exported by the in-process SDK | exported by the in-process SDK (unchanged) |
| Runtime requirements | none | Linux 5.8+ with BTF, privileged agent |

In OBI mode the library does **not** register a global `TracerProvider`: registering one would
prevent OBI from activating the Go Auto SDK. The W3C `traceparent` propagator is still registered.

## Using it in a service

1. Set the environment variables:
   ```
   OTEL_ENABLED=true
   OTEL_TRACE_MODE=obi
   ```
2. Remove (or disable) HTTP/gRPC/DB tracing middleware such as `otelgin`, otherwise request spans are
   produced twice (once by the middleware, once by OBI).
3. Keep creating business-logic spans with `tracer.Tracer()`, always starting them from the request
   context (`c.Request.Context()`) so they are attached to the request span captured by OBI.
4. Opt the pod in to OBI with the label `kyber-trace/obi: "true"` (see `kubernetes/configmap.yaml`).
   OBI takes the service name from the `OTEL_SERVICE_NAME` env var of the process.

See [`example/obi.go`](../../example/obi.go).

## Kubernetes

```
kubectl apply -f deploy/obi/kubernetes/
```

This deploys OBI as a DaemonSet in the `observability` namespace, exporting to the existing
`opentelemetry-collector-agent` over OTLP/HTTP (port 4318). Adjust the endpoint and sampler in
`kubernetes/configmap.yaml` for your cluster.

## Local playground (Docker)

Requires a Linux host or Docker Desktop's Linux VM:

```
docker compose -f deploy/obi/docker/compose.yaml up --build
curl http://localhost:8080/api/v1/greeting
```

The collector prints the received spans. Each request produces one trace:

```
GET /api/v1/greeting   (Server, captured by OBI)
├── in queue           (Internal, captured by OBI)
└── processing         (Internal, captured by OBI)
    └── build greeting (Internal, manual span from tracer.Tracer(), exported by OBI via the Go Auto SDK)
```

Stopping the `obi` container can take ~1 minute while it detaches its eBPF probes (Docker may
report it as a zombie in the meantime); it then exits on its own.

## Caveats

- The Go Auto SDK only activates when the `go.opentelemetry.io/otel` version is supported by the OBI release,
  the module is not `replace`d, and OBI may use `bpf_probe_write_user` (privileged / `CAP_SYS_ADMIN`,
  kernel not in integrity lockdown mode). If any of these fail, OBI still records request spans, but manual
  spans are dropped silently. Upgrade OBI together with the otel version of this library.
- OBI only supports Linux; on other platforms use `OTEL_TRACE_MODE=sdk`.
- On kernels 6.6.128+, 6.12.75+, 6.18.14+ and 6.19+, OBI may log `context propagation is disabled`
  (kernel `FIONREAD` regression). Spans inside the service are still correlated, but OBI will not
  inject `traceparent` into outgoing requests, so cross-service traces break. Seen on Docker Desktop's
  WSL2 kernel 6.18.
- Tested with OBI v1.0.0, `go.opentelemetry.io/otel` v1.47.0 and Go 1.27.3.
