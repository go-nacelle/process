package process

import (
	"time"

	"github.com/derision-test/glock"
)

// RunnerConfigFunc is a function used to configure an instance
// of a ProcessRunner.
type RunnerConfigFunc func(*runner)

// WithInjectHook sets the inject hook used by the runner.
func WithInjectHook(injectHook InjectHook) RunnerConfigFunc {
	return func(r *runner) { r.injectHook = injectHook }
}

// WihHealth sets the health container used by the runner.
func WithHealth(health Health) RunnerConfigFunc {
	return func(r *runner) { r.health = health }
}

// WithLogger sets the logger used by the runner.
func WithLogger(logger Logger) RunnerConfigFunc {
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
