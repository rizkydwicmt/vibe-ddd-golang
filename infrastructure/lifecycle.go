package infrastructure

import (
	"context"
	"time"
)

// infrastructureStartupCheckTimeout bounds each dependency's OnStart readiness
// check so a wedged dependency fails the boot rather than hanging it.
const infrastructureStartupCheckTimeout = 15 * time.Second

// boundedWaitTimeout returns the smaller of fallback and the context's remaining
// deadline, so a check never waits past its parent's budget.
func boundedWaitTimeout(ctx context.Context, fallback time.Duration) time.Duration {
	if deadline, ok := ctx.Deadline(); ok {
		if remaining := time.Until(deadline); remaining > 0 && remaining < fallback {
			return remaining
		}
	}
	return fallback
}
