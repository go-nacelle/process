package process

import (
	"context"
	"time"

	"github.com/go-nacelle/log"
)

// InitializerMeta wraps an initializer with some package
// private fields.
type InitializerMeta struct {
	Initializer
	contextFilter   func(ctx context.Context) context.Context
	name            string
	logFields       log.LogFields
	priority        int
	initTimeout     time.Duration
	finalizeTimeout time.Duration
}

func newInitializerMeta(initializer Initializer) *InitializerMeta {
	return &InitializerMeta{
		Initializer: initializer,
	}
}

func (m *InitializerMeta) FilterContext(ctx context.Context) context.Context {
	if m.contextFilter == nil {
		return ctx
	}

	return m.contextFilter(ctx)
}

// Name returns the name of the initializer.
func (m *InitializerMeta) Name() string {
	if m.name == "" {
		return "<unnamed>"
	}

	return m.name
}

// LogFields returns logging fields registered to this initializer.
func (m *InitializerMeta) LogFields() log.LogFields {
	return m.logFields
}

// InitTimeout returns the maximum timeout allowed for a call to
// the Init function. A zero value indicates no timeout.
func (m *InitializerMeta) InitTimeout() time.Duration {
	return m.initTimeout
}

// FinalizeTimeout returns the maximum timeout allowed for a call to
// the Finalize function. A zero value indicates no timeout.
func (m *InitializerMeta) FinalizeTimeout() time.Duration {
	return m.finalizeTimeout
}

// Wrapped returns the underlying initializer.
func (m *InitializerMeta) Wrapped() interface{} {
	return m.Initializer
}
