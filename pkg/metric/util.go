package metric

import (
	"context"
	"errors"
)

// Flush flushes all pending telemetry.
//
// This method honors the deadline or cancellation of ctx. An appropriate
// error will be returned in these situations. There is no guaranteed that all
// telemetry be flushed or all resources have been released in these
// situations.
//
// This method is safe to call concurrently.
func Flush(ctx context.Context) error {
	if provider != nil {
		return provider.ForceFlush(ctx)
	} else {
		return errors.New("no meter provider was initialized")
	}
}

// Shutdown shuts down the MeterProvider flushing all pending telemetry and
// releasing any held computational resources.
//
// This call is idempotent. The first call will perform all flush and
// releasing operations. Subsequent calls will perform no action and will
// return an error stating this.
//
// Measurements made by instruments from meters this MeterProvider created
// will not be exported after Shutdown is called.
//
// This method honors the deadline or cancellation of ctx. An appropriate
// error will be returned in these situations. There is no guaranteed that all
// telemetry be flushed or all resources have been released in these
// situations.
//
// This method is safe to call concurrently.
func Shutdown(ctx context.Context) error {
	if provider != nil {
		return provider.Shutdown(ctx)
	} else {
		return errors.New("no meter provider was initialized")
	}
}
