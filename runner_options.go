package process

import (
	"time"

	"github.com/derision-test/glock"
	"github.com/go-nacelle/log"
)

// RunnerConfigFunc is a function used to configure an instance
// of a ProcessRunner.
type RunnerConfigFunc func(*runner)

// WithLogger sets the logger used by the runner.
func WithLogger(logger log.Logger) RunnerConfigFunc {
	return func(r *runner) { r.logger = logger }
}

// WithClock sets the clock used by the runner.
func WithClock(clock glock.Clock) RunnerConfigFunc {
	return func(r *runner) { r.clock = clock }
}

// WithStartTimeout sets the time it will wait for a process to become
// healthy after startup.
func WithStartTimeout(timeout time.Duration) RunnerConfigFunc {
	return func(r *runner) { r.startupTimeout = timeout }
}

// WithHealthCheckInterval sets the frequency between checks of process
// health during startup.
func WithHealthCheckInterval(interval time.Duration) RunnerConfigFunc {
	return func(r *runner) { r.healthCheckInterval = interval }
}

// WithShutdownTimeout sets the maximum time it will wait for a process to
// exit during a graceful shutdown.
func WithShutdownTimeout(timeout time.Duration) RunnerConfigFunc {
	return func(r *runner) { r.shutdownTimeout = timeout }
}
