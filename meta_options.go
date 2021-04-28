package process

import (
	"context"
	"time"

	"github.com/derision-test/glock"
)

type metaOptions struct {
	health          *Health
	healthKeys      []interface{}
	contextFilter   func(ctx context.Context) context.Context
	name            string
	metadata        map[string]interface{}
	priority        int
	allowEarlyExit  bool
	initTimeout     time.Duration
	startupTimeout  time.Duration
	stopTimeout     time.Duration
	shutdownTimeout time.Duration
	finalizeTimeout time.Duration
	logger          Logger
	initClock       glock.Clock
	startupClock    glock.Clock
	stopClock       glock.Clock
	shutdownClock   glock.Clock
	finalizeClock   glock.Clock
}

type MetaConfigFunc func(meta *metaOptions)

// WithMetaHealth configures a Meta instance to use the given health instance.
func WithMetaHealth(health *Health) MetaConfigFunc {
	return func(meta *metaOptions) { meta.health = health }
}

// WithMetaHealthKey configures a Meta instance to use the given keys to search for
// registered health component belonging to this process.
func WithMetaHealthKey(healthKeys ...interface{}) MetaConfigFunc {
	return func(meta *metaOptions) { meta.healthKeys = append(meta.healthKeys, healthKeys...) }
}

// WithMetaContext configures a Meta instance to use the context returned by the given
// function when invoking the wrapped value's underlying Init, Start, Stop, or Finalize
// methods.
func WithMetaContext(f func(ctx context.Context) context.Context) MetaConfigFunc {
	return func(meta *metaOptions) { meta.contextFilter = f }
}

// WithMetaName tags a Meta instance with the given name.
func WithMetaName(name string) MetaConfigFunc {
	return func(meta *metaOptions) { meta.name = name }
}

// WithMetaPriority tags a Meta instance with the given priority.
func WithMetaPriority(priority int) MetaConfigFunc {
	return func(meta *metaOptions) { meta.priority = priority }
}

// WithMetadata tags a Meta instance with the given metadata.
func WithMetadata(metadata map[string]interface{}) MetaConfigFunc {
	return func(meta *metaOptions) { meta.metadata = metadata }
}

// WithEarlyExit sets the flag that determines if the process is allowed to return from
// the Start method (with a nil error value) before the Stop method is called. The default
// behavior is to treat exiting processes as erroneous.
func WithEarlyExit(allowed bool) MetaConfigFunc {
	return func(meta *metaOptions) { meta.allowEarlyExit = allowed }
}

// WithMetaInitTimeout configures a Meta instance with the given timeout for the
// invocation of the wrapped value's Init method.
func WithMetaInitTimeout(timeout time.Duration) MetaConfigFunc {
	return func(meta *metaOptions) { meta.initTimeout = timeout }
}

// WithMetaStartupTimeout configures a Meta instance with the given timeout for the
// time between the wrapped value's Start method being invoked and the process becoming
// healthy.
func WithMetaStartupTimeout(timeout time.Duration) MetaConfigFunc {
	return func(meta *metaOptions) { meta.startupTimeout = timeout }
}

// WithMetaStopTimeout configures a Meta instance with the given timeout for the
// invocation of the wrapped value's Stop method.
func WithMetaStopTimeout(timeout time.Duration) MetaConfigFunc {
	return func(meta *metaOptions) { meta.stopTimeout = timeout }
}

// WithMetaShutdownTimeout configures a Meta instance with the given timeout for the
// time between the wrapped value's Stop method being invoked and the process's Start
// method unblocking.
func WithMetaShutdownTimeout(timeout time.Duration) MetaConfigFunc {
	return func(meta *metaOptions) { meta.shutdownTimeout = timeout }
}

// WithMetaFinalizeTimeout configures a Meta instance with the given timeout for the
// invocation of the wrapped value's Finalize method.
func WithMetaFinalizeTimeout(timeout time.Duration) MetaConfigFunc {
	return func(meta *metaOptions) { meta.finalizeTimeout = timeout }
}

// WithMetaLogger configures a Meta instance with the given logger instance.
func WithMetaLogger(logger Logger) MetaConfigFunc {
	return func(meta *metaOptions) { meta.logger = logger }
}

func withMetaInitClock(clock glock.Clock) MetaConfigFunc {
	return func(meta *metaOptions) { meta.initClock = clock }
}

func withMetaStartupClock(clock glock.Clock) MetaConfigFunc {
	return func(meta *metaOptions) { meta.startupClock = clock }
}

func withMetaStopClock(clock glock.Clock) MetaConfigFunc {
	return func(meta *metaOptions) { meta.stopClock = clock }
}

func withMetaShutdownClock(clock glock.Clock) MetaConfigFunc {
	return func(meta *metaOptions) { meta.shutdownClock = clock }
}

func withMetaFinalizeClock(clock glock.Clock) MetaConfigFunc {
	return func(meta *metaOptions) { meta.finalizeClock = clock }
}
