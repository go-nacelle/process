package process

import (
	"context"

	"github.com/go-nacelle/config"
)

// Initializer is an interface that is called once on app
// startup.
type Initializer interface {
	// Init reads the given configuration and prepares
	// something for use by a process. This can be loading
	// files from disk, connecting to a remote service,
	// initializing shared data structures, and inserting
	// a service into a shared service container.
	Init(ctx context.Context, config config.Config) error
}

// Finalizer is an optional extension of an Initializer that
// supports finalization. This is useful for initializers
// that need to tear down a background process before the
// process exits, but needs to be started early in the boot
// process (such as flushing logs or metrics).
type Finalizer interface {
	// Finalize is called after the application has stopped
	// all running processes.
	Finalize() error // TODO - add context?
}

// InitializerFunc is a non-struct version of an initializer.
type InitializerFunc func(ctx context.Context, config config.Config) error

// Init calls the underlying function.
func (f InitializerFunc) Init(ctx context.Context, config config.Config) error {
	return f(ctx, config)
}
