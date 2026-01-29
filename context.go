package core

import (
	"context"
	"os"
	"os/signal"
)

// BuildContext constructs a root application context annotated with identity
// ([AppName]), version ([AppVersion]) and pipeline I/O ([Pipeline]), and
// automatically wired to OS interrupt signals. The returned cancel function
// MUST be invoked by the caller to release signal resources.
//
// Cancellation Sources:
//   - Incoming SIGINT / SIGKILL ([os.Interrupt], [os.Kill]) trigger context
//     cancellation for graceful shutdown.
//   - Manual invocation of the returned cancel function.
//
// The supplied [Pipeline] is stored for later retrieval via [PipelinesFromContext].
// Prefer passing explicit version / name values; fallback defaults remain
// available through helper extraction funcs.
func BuildContext(
	name AppName,
	version AppVersion,
	pipeline Pipeline,
) (
	ctx context.Context,
	cancel context.CancelFunc,
) {
	ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	ctx = WithAppName(ctx, name)
	ctx = WithVersion(ctx, version)
	ctx = WithPipelines(ctx, pipeline)

	return ctx, cancel
}
