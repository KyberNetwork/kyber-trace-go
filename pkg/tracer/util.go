package tracer

import (
	"context"
	"errors"
)

// Shutdown shuts down TracerProvider. All registered span processors are shut down
// in the order they were registered and any held computational resources are released.
// After Shutdown is called, all methods are no-ops.
// In OBI mode, spans are owned by the external eBPF agent, so this is a no-op.
func Shutdown(ctx context.Context) error {
	if provider != nil {
		return provider.Shutdown(ctx)
	} else if IsOBIMode() {
		return nil
	} else {
		return errors.New("no tracer provider was initialized")
	}
}

// Flush immediately exports all spans that have not yet been exported for
// all the registered span processors.
// In OBI mode, spans are owned by the external eBPF agent, so this is a no-op.
func Flush(ctx context.Context) error {
	if provider != nil {
		return provider.ForceFlush(ctx)
	} else if IsOBIMode() {
		return nil
	} else {
		return errors.New("no tracer provider was initialized")
	}
}
