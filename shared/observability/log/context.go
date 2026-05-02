package log

import "context"

// FromContext is a free-function alias for Logger.FromContext that mirrors
// the helper signature requested by RDK-003: callers can write
//
//	log.FromContext(ctx).Info(msg, data)
//
// once they have a *Logger, and `log.FromContext(base, ctx)` when the base
// logger is passed in by dependency injection. We expose both styles so
// existing call sites that hold a domain.LogProvider can be migrated
// incrementally without ripping out the DI wiring.
func FromContext(base *Logger, ctx context.Context) *Logger {
	if base == nil {
		return nil
	}
	return base.FromContext(ctx)
}
