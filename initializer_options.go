package process

import (
	"time"

	"github.com/go-nacelle/log"
)

// InitializerConfigFunc is a function used to append additional
// metadata to an initializer during registration.
type InitializerConfigFunc func(*InitializerMeta)

// WithInitializerName assigns a name to an initializer, visible in logs.
func WithInitializerName(name string) InitializerConfigFunc {
	return func(meta *InitializerMeta) { meta.name = name }
}

// WithInitializerLogFields sets additional fields sent with every log message
// from this initializer.
func WithInitializerLogFields(fields log.LogFields) InitializerConfigFunc {
	return func(meta *InitializerMeta) { meta.logFields = fields }
}

// WithInitializerTimeout sets the time limit for the initializer.
func WithInitializerTimeout(timeout time.Duration) InitializerConfigFunc {
	return func(meta *InitializerMeta) { meta.initTimeout = timeout }
}

// WithFinalizerTimeout sets the time limit for the finalizer.
func WithFinalizerTimeout(timeout time.Duration) InitializerConfigFunc {
	return func(meta *InitializerMeta) { meta.finalizeTimeout = timeout }
}
