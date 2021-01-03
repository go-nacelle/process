package process

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/derision-test/glock"
	"github.com/go-nacelle/log"
	"github.com/go-nacelle/service"
)

// ParallelInitializer is a container for initializers that are initialized in
// parallel. This is useful when groups of initializers are independent and may
// contain some longer-running process (such as dialing a remote service).
type ParallelInitializer struct {
	Logger       log.Logger               `service:"logger"`
	Services     service.ServiceContainer `service:"services"`
	clock        glock.Clock
	initializers []*InitializerMeta
}

var _ Initializer = &ParallelInitializer{}
var _ Configurable = &ParallelInitializer{}

// NewParallelInitializer creates a new parallel initializer.
func NewParallelInitializer(initializerConfigs ...ParallelInitializerConfigFunc) *ParallelInitializer {
	pi := &ParallelInitializer{
		initializers: []*InitializerMeta{},
	}

	for _, f := range initializerConfigs {
		f(pi)
	}

	return pi
}

// RegisterInitializer adds an initializer to the initializer set
// with the given configuration.
func (i *ParallelInitializer) RegisterInitializer(
	initializer Initializer,
	initializerConfigs ...InitializerConfigFunc,
) {
	meta := newInitializerMeta(initializer)

	for _, f := range initializerConfigs {
		f(meta)
	}

	i.initializers = append(i.initializers, meta)
}

// RegisterConfiguration runs RegisterConfiguration on all registered initializers.
func (pi *ParallelInitializer) RegisterConfiguration(config ConfigurationRegistry) {
	for _, initializer := range pi.initializers {
		if configurable, ok := initializer.Wrapped().(Configurable); ok {
			configurable.RegisterConfiguration(config)
		}
	}
}

// Init runs Init on all registered initializers concurrently.
func (pi *ParallelInitializer) Init(ctx context.Context) error {
	for _, initializer := range pi.initializers {
		if err := pi.inject(initializer); err != nil {
			return errMetaSet{
				errMeta{err: err, source: initializer},
			}
		}
	}

	errMetas := errMetaSet{}
	initErrors := pi.initializeAll(ctx)

	for i, err := range initErrors {
		if err != nil {
			errMetas = append(errMetas, errMeta{err: err, source: pi.initializers[i]})
		}
	}

	if len(errMetas) > 0 {
		for i, err := range pi.finalizeAll(ctx, initErrors) {
			if err != nil {
				errMetas = append(errMetas, errMeta{err: err, source: pi.initializers[i]})
			}
		}

		return errMetas
	}

	return nil
}

// Finalize runs Finalize on all registered initializers concurrently.
func (pi *ParallelInitializer) Finalize(ctx context.Context) error {
	errMetas := errMetaSet{}
	for i, err := range pi.finalizeAll(ctx, make([]error, len(pi.initializers))) {
		if err != nil {
			errMetas = append(errMetas, errMeta{err: err, source: pi.initializers[i]})
		}
	}

	if len(errMetas) > 0 {
		return errMetas
	}

	return nil
}

func (pi *ParallelInitializer) inject(initializer namedInjectable) error {
	pi.Logger.WithFields(initializer.LogFields()).Info("Injecting services into %s", initializer.Name())

	if err := inject(initializer, pi.Services, pi.Logger); err != nil {
		return fmt.Errorf(
			"failed to inject services into %s (%s)",
			initializer.Name(),
			err.Error(),
		)
	}

	return nil
}

func (pi *ParallelInitializer) initializeAll(ctx context.Context) []error {
	errors := make([]error, len(pi.initializers))
	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}

	for i := range pi.initializers {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			if err := pi.initWithTimeout(ctx, pi.initializers[i]); err != nil {
				mutex.Lock()
				errors[i] = err
				mutex.Unlock()
			}
		}(i)
	}

	wg.Wait()
	return errors
}

func (pi *ParallelInitializer) finalizeAll(ctx context.Context, initErrors []error) []error {
	errors := make([]error, len(pi.initializers))
	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}

	for i := range pi.initializers {
		if initErrors[i] != nil {
			continue
		}

		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			if err := pi.finalizeWithTimeout(ctx, pi.initializers[i]); err != nil {
				mutex.Lock()
				errors[i] = err
				mutex.Unlock()
			}
		}(i)
	}

	wg.Wait()
	return errors
}

func (pi *ParallelInitializer) initWithTimeout(ctx context.Context, initializer namedInitializer) error {
	errChan := makeErrChan(func() error {
		return pi.init(ctx, initializer)
	})

	select {
	case err := <-errChan:
		return err

	case <-pi.makeTimeoutChan(initializer.InitTimeout()):
		return fmt.Errorf("%s did not initialize within timeout", initializer.Name())
	}
}

func (pi *ParallelInitializer) init(ctx context.Context, initializer namedInitializer) error {
	pi.Logger.WithFields(initializer.LogFields()).Info("Initializing %s", initializer.Name())

	if err := initializer.Init(ctx); err != nil {
		return fmt.Errorf(
			"failed to initialize %s (%s)",
			initializer.Name(),
			err.Error(),
		)
	}

	pi.Logger.WithFields(initializer.LogFields()).Info("Initialized %s", initializer.Name())
	return nil
}

func (pi *ParallelInitializer) finalizeWithTimeout(ctx context.Context, initializer namedFinalizer) error {
	errChan := makeErrChan(func() error {
		return pi.finalize(ctx, initializer)
	})

	select {
	case err := <-errChan:
		return err

	case <-pi.makeTimeoutChan(initializer.FinalizeTimeout()):
		return fmt.Errorf("%s did not finalize within timeout", initializer.Name())
	}
}

func (pi *ParallelInitializer) finalize(ctx context.Context, initializer namedFinalizer) error {
	finalizer, ok := initializer.Wrapped().(Finalizer)
	if !ok {
		return nil
	}

	pi.Logger.WithFields(initializer.LogFields()).Info("Finalizing %s", initializer.Name())

	if err := finalizer.Finalize(ctx); err != nil {
		return fmt.Errorf(
			"%s returned error from finalize (%s)",
			initializer.Name(),
			err.Error(),
		)
	}

	pi.Logger.WithFields(initializer.LogFields()).Info("Finalized %s", initializer.Name())
	return nil
}

func (pi *ParallelInitializer) makeTimeoutChan(timeout time.Duration) <-chan time.Time {
	return makeTimeoutChan(pi.clock, timeout)
}
