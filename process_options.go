package process

import (
	"context"
	"time"

	"github.com/go-nacelle/log"
)

// ProcessConfigFunc is a function used to append additional metadata
// to an process during registration.
type ProcessConfigFunc func(*ProcessMeta)

// WithProcessContextFilter sets the context filter for the process.
func WithProcessContextFilter(f func(ctx context.Context) context.Context) ProcessConfigFunc {
	return func(meta *ProcessMeta) { meta.contextFilter = f }
}

// WithProcessName assigns a name to an process, visible in logs.
func WithProcessName(name string) ProcessConfigFunc {
	return func(meta *ProcessMeta) { meta.name = name }
}

// WithProcessLogFields sets additional fields sent with every log message
// from this process.
func WithProcessLogFields(fields log.LogFields) ProcessConfigFunc {
	return func(meta *ProcessMeta) { meta.logFields = fields }
}

// WithPriority assigns a priority to a process. A process with a lower-valued
// priority is initialized and started before a process with a higher-valued
// priority. Two processes with the same priority are started concurrently.
func WithPriority(priority int) ProcessConfigFunc {
	return func(meta *ProcessMeta) { meta.priority = priority }
}

// WithSilentExit allows a process to exit without causing the program to halt.
// The default is the opposite, where the completion of any registered process
// (even successful) causes a graceful shutdown of the other processes.
func WithSilentExit() ProcessConfigFunc {
	return func(meta *ProcessMeta) { meta.silentExit = true }
}

// WithProcessInitTimeout sets the time limit for the process's Init method.
func WithProcessInitTimeout(timeout time.Duration) ProcessConfigFunc {
	return func(meta *ProcessMeta) { meta.initTimeout = timeout }
}

// WithProcessStartTimeout sets the time limit for the process to become healthy.
func WithProcessStartTimeout(timeout time.Duration) ProcessConfigFunc {
	return func(meta *ProcessMeta) { meta.startTimeout = timeout }
}

// WithProcessShutdownTimeout sets the time limit for the process's Start method
// to yield after the Stop method has been called.
func WithProcessShutdownTimeout(timeout time.Duration) ProcessConfigFunc {
	return func(meta *ProcessMeta) { meta.shutdownTimeout = timeout }
}

// WithProcessFinalizerTimeout sets the time limit for the finalizer.
func WithProcessFinalizerTimeout(timeout time.Duration) ProcessConfigFunc {
	return func(meta *ProcessMeta) { meta.finalizeTimeout = timeout }
}
