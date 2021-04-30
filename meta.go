package process

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/derision-test/glock"
)

// Meta is a wrapper around a process value. This wrapper ensures that the configured
// receiver's methods are only called once and not called from an invalid state
// (e.g. Run called before Init or after a failed Init).
type Meta struct {
	wrapped     interface{}
	options     *metaOptions
	logger      Logger
	mu          sync.Mutex
	initialized bool
	running     bool
	stopping    bool
	stopped     chan struct{}
}

var defaultClock = glock.NewRealClock()

func newMeta(wrapped interface{}, configs ...MetaConfigFunc) *Meta {
	options := &metaOptions{
		health:        NewHealth(),
		contextFilter: func(ctx context.Context) context.Context { return ctx },
		logger:        NilLogger,
		initClock:     defaultClock,
		startupClock:  defaultClock,
		stopClock:     defaultClock,
		shutdownClock: defaultClock,
		finalizeClock: defaultClock,
	}

	for _, f := range configs {
		f(options)
	}

	return &Meta{
		wrapped: wrapped,
		options: options,
		logger:  options.logger.WithFields(options.metadata),
		stopped: make(chan struct{}),
	}
}

// Wrapped returns the underlying receiver.
func (m *Meta) Wrapped() interface{} {
	return m.wrapped
}

// Name returns the process's configured name or `<unnamed>` if one was not supplied.
func (m *Meta) Name() string {
	if m.options.name == "" {
		return "<unnamed>"
	}

	return m.options.name
}

// Metadata returns the process's configured metadata.
func (m *Meta) Metadata() map[string]interface{} {
	return m.options.metadata
}

// Init invokes the wrapped value's Init method.
//
// A timeout error will be returned if the invocation does not unblock within the configured
// init timeout.
func (m *Meta) Init(ctx context.Context) (err error) {
	defer func() {
		if err == nil {
			m.mu.Lock()
			m.initialized = true
			m.mu.Unlock()
		}
	}()

	if initializer, ok := m.wrapped.(Initializer); ok {
		return m.makeRunWithTimeout(ctx, "init", initializer.Init, m.options.initClock, m.options.initTimeout)
	}

	return nil
}

// Run invokes the wrapped value's Run method.
//
// A timeout error will be returned if the associated health instance does not report
// healthy within the configured startup timeout, or if Run method does not unblock
// after the meta instance's Stop method is called within the configured shutdown
// timeout.
//
// A canned error will also be returned if the underlying Run method unblocks with a
// nil value without Stop being called  and the meta was not configured with the silent
// exit flag.
//
// This method will no-op if the meta instance was not initialized.
func (m *Meta) Run(ctx context.Context) error {
	if runner, ok := m.wrapped.(Runner); ok && m.shouldRun() {
		defer func() {
			m.mu.Lock()
			defer m.mu.Unlock()
			m.running = false
		}()

		return m.run(ctx, runner)
	}

	return nil
}

func (m *Meta) shouldRun() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized || m.stopping {
		return false
	}

	m.running = true
	return true
}

func (m *Meta) run(ctx context.Context, runner Runner) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	result := runAsync(ctx, func(ctx context.Context) error {
		return m.makeRunWithTimeout(ctx, "run", runner.Run, nil, 0)
	})

	healthStatusChannel, err := m.watchHealthStatus()
	if err != nil {
		return err
	}

	select {
	case v := <-healthStatusChannel:
		if !v {
			return ErrStartupTimeout
		}

		select {
		case err := <-result:
			return m.handleResult(ctx, err)

		case <-m.stopped:
			cancel()
		}

	case err := <-result:
		return m.handleResult(ctx, err)

	case <-m.stopped:
		cancel()
	}

	select {
	case err := <-result:
		return ignoreContextError(ctx, err)

	case <-afterZeroUnbounded(m.options.shutdownClock, m.options.shutdownTimeout):
		return ErrShutdownTimeout
	}
}

// watchHealthStatus returns a channel that will receive the value true when the process
// becomes healthy, or false after the startup timeout has elapsed. If there are no health
// keys registered to this process then a nil channel is returned. Note that reading from
// a nil channel blocks forever.
func (m *Meta) watchHealthStatus() (chan bool, error) {
	components, err := m.options.health.GetAll(m.options.healthKeys...)
	if err != nil {
		return nil, err
	}

	return m.watchHealthComponentStatus(components), nil
}

// watchHealthStatus returns a channel that will receive the value true when all of the
// given health components healthy, or false after the startup timeout has elapsed.
func (m *Meta) watchHealthComponentStatus(components []*HealthComponentStatus) chan bool {
	timedOut := make(chan bool, 1)

	go func() {
		defer close(timedOut)

		ch, cancel := m.options.health.Subscribe()
		defer cancel()

		for {
			select {
			case <-afterZeroUnbounded(m.options.startupClock, m.options.startupTimeout):
				return

			case <-ch:
				healthy := true
				for _, component := range components {
					if !component.Healthy() {
						healthy = false
						break
					}
				}

				if healthy {
					timedOut <- true
					return
				}
			}
		}
	}()

	return timedOut
}

// handleResult is invoked after a value is received from the underlying run method.
// This will mark the meta instance as stopping, and determine the appropriate error
// value.
func (m *Meta) handleResult(ctx context.Context, err error) error {
	m.mu.Lock()
	stopping := m.stopping
	m.mu.Unlock()

	if err == nil && !stopping && !m.options.allowEarlyExit {
		err = ErrUnexpectedReturn
	}

	return ignoreContextError(ctx, err)
}

// ignoreContextError returns nil if the given error is equal to the given context's
// underlying error and the given error otherwise. The given error may be wrapped.
func ignoreContextError(ctx context.Context, err error) error {
	for ex := err; ex != nil; ex = errors.Unwrap(ex) {
		if ex == ctx.Err() {
			return nil
		}
	}

	return err
}

// Stop invokes the wrapped value's Stop method.
//
// A timeout error will be returned if the invocation does not unblock within the configured
// stop timeout.
//
// This method will no-op if the meta instance is not currently running.
func (m *Meta) Stop(ctx context.Context) error {
	if !m.shouldRunStop() {
		return nil
	}

	defer close(m.stopped)

	if stopper, ok := m.wrapped.(Stopper); ok {
		return m.makeRunWithTimeout(ctx, "stop", stopper.Stop, m.options.stopClock, m.options.stopTimeout)
	}

	return nil
}

func (m *Meta) shouldRunStop() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return false
	}

	m.stopping = true
	return m.running
}

// Finalize invokes the wrapped value's Finalize method.
//
// A timeout error will be returned if the invocation does not unblock within the configured
// finalize timeout.
//
// This method will no-op if the meta instance was not initialized.
func (m *Meta) Finalize(ctx context.Context) error {
	if finalizer, ok := m.wrapped.(Finalizer); ok && m.shouldRunFinalize() {
		return m.makeRunWithTimeout(ctx, "finalize", finalizer.Finalize, m.options.finalizeClock, m.options.finalizeTimeout)
	}

	return nil
}

func (m *Meta) shouldRunFinalize() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.initialized
}

func (m *Meta) makeRunWithTimeout(ctx context.Context, opName string, fn func(ctx context.Context) error, clock glock.Clock, timeout time.Duration) error {
	m.logger.Info("%s: %s starting", m.Name(), opName)

	ctx, cancel := context.WithCancel(m.options.contextFilter(ctx))
	defer cancel()

	select {
	case err := <-toStreamErrorFunc(fn)(ctx):
		if err != nil {
			return &opError{
				source:   err,
				metaName: m.Name(),
				opName:   opName,
				message:  "failed",
			}
		}

		m.logger.Info("%s: %s finished", m.Name(), opName)
		return nil

	case <-afterZeroUnbounded(clock, timeout):
		return &opError{
			source:   nil,
			metaName: m.Name(),
			opName:   opName,
			message:  "timeout",
		}
	}
}

// afterZeroUnbounded returns a channel that will receive a value after the given
// timeout. If the given timeout is zero, a nil channel will be returned. Note that
// reading from a nil channel blocks forever.
func afterZeroUnbounded(clock glock.Clock, timeout time.Duration) <-chan time.Time {
	if timeout == 0 {
		return nil
	}

	return clock.After(timeout)
}

// withTimeoutZeroUnbounded returns a context that will be cancelled after the given
// timeout, as well as a cleanup function. If the given timeout is zero, the context
// is returned unmodified.
func withTimeoutZeroUnbounded(ctx context.Context, clock glock.Clock, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout == 0 {
		return ctx, func() {}
	}

	return glock.ContextWithTimeout(ctx, clock, timeout)
}
