package process

import "context"

// Injecter tags a struct as being an expected target of dependency injection.
type Injecter interface {
	// Inject populates the fields of the wrapped value in the given meta value.
	Inject(ctx context.Context, meta *Meta) error
}

// Initializer wraps behavior that happens on application startup.
type Initializer interface {
	// Init is the hook invoked on application startup.
	Init(ctx context.Context) error
}

// Process wraps behavior that happens continually through the application
// lifecycle.
type Process interface {
	// Init is the hook invoked on application startup. The process hook is
	// expected to be long-running. Returning early, prior to the given context
	// being canceled, is generally an unexpected event.
	Start(ctx context.Context) error
}

// Stoppable wraps a process with way to signal graceful exit.
type Stoppable interface {
	// Stop is the hookinvoked on a running process immediately prior to the
	// root context being canceled. This method may exit immediately and is
	// not expected to synchronize on Start returning.
	Stop(ctx context.Context) error
}

// Finalizer wraps behavior that happens directly before application exit.
type Finalizer interface {
	// Finalize is the hook invoked directly before application exit.
	Finalize(ctx context.Context) error
}

// InjecterFunc is a function conforming to the Injecter interface.
type InjecterFunc func(ctx context.Context, meta *Meta) error

// InitializerFunc is a function conforming to the Initializer interface.
type InitializerFunc func(ctx context.Context) error

// ProcessFunc is a function conforming to the Process interface.
type ProcessFunc func(ctx context.Context) error

// FinalizerFunc is a function conforming to the Finalizer interface.
type FinalizerFunc func(ctx context.Context) error

func (f InjecterFunc) Inject(ctx context.Context, meta *Meta) error { return f(ctx, meta) }
func (f InitializerFunc) Init(ctx context.Context) error            { return f(ctx) }
func (f ProcessFunc) Start(ctx context.Context) error               { return f(ctx) }
func (f FinalizerFunc) Finalize(ctx context.Context) error          { return f(ctx) }
