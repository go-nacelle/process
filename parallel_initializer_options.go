package process

import (
	"github.com/derision-test/glock"
	"github.com/go-nacelle/log"
	"github.com/go-nacelle/service"
)

// ParallelInitializerConfigFunc is a function used to configure an instance
// of a ParallelInitializer.
type ParallelInitializerConfigFunc func(*ParallelInitializer)

// WithParallelInitializerLogger sets the logger used by the runner.
func WithParallelInitializerLogger(logger log.Logger) ParallelInitializerConfigFunc {
	return func(pi *ParallelInitializer) { pi.Logger = logger }
}

// WithParallelInitializerContainer sets the service container used by the runner.
func WithParallelInitializerContainer(container service.ServiceContainer) ParallelInitializerConfigFunc {
	return func(pi *ParallelInitializer) { pi.Services = container }
}

// WithParallelInitializerClock sets the clock used by the runner.
func WithParallelInitializerClock(clock glock.Clock) ParallelInitializerConfigFunc {
	return func(pi *ParallelInitializer) { pi.clock = clock }
}
