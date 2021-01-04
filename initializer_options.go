package process

import (
	"context"
	"time"
)

// InitializerConfigFunc is a function used to append additional
// metadata to an initializer during registration.
type InitializerConfigFunc func(*InitializerMeta)

// WithInitializerContextFilter sets the context filter for the initializer.
func WithInitializerContextFilter(f func(ctx context.Context) context.Context) InitializerConfigFunc {
	return func(meta *InitializerMeta) { meta.contextFilter = f }
}

// WithInitializerName assigns a name to an initializer, visible in logs.
func WithInitializerName(name string) InitializerConfigFunc {
	return func(meta *InitializerMeta) { meta.name = name }
}

// WithInitializerLogFields sets additional fields sent with every log message
// from this initializer.
func WithInitializerLogFields(fields LogFields) InitializerConfigFunc {
	return func(meta *InitializerMeta) { meta.logFields = fields }
}

// WithInitializerPriority assigns a priority to an initializer. A initializer with a
// lower valued priority is initialized before an initializer with a higher valued priority.
// Two initializers with the same priority may be called concurrently.
func WithInitializerPriority(priority int) InitializerConfigFunc {
	return func(meta *InitializerMeta) { meta.priority = priority }
}

// WithInitializerTimeout sets the time limit for the initializer.
func WithInitializerTimeout(timeout time.Duration) InitializerConfigFunc {
	return func(meta *InitializerMeta) { meta.initTimeout = timeout }
}

// WithFinalizerTimeout sets the time limit for the finalizer.
func WithFinalizerTimeout(timeout time.Duration) InitializerConfigFunc {
	return func(meta *InitializerMeta) { meta.finalizeTimeout = timeout }
}
