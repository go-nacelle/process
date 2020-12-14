package process

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/derision-test/glock"

	"github.com/go-nacelle/config"
	"github.com/go-nacelle/log"
	"github.com/go-nacelle/service"
)

// Runner wraps a process container. Given a loaded configuration object,
// it can run the registered initializers and processes and wait for them
// to exit (cleanly or via shutdown request).
type Runner interface {
	// Run takes a loaded configuration object, then starts and monitors
	// the registered items in the process container. This method returns
	// a channel of errors. Each error from an initializer or a process will
	// be sent on this channel (nil errors are ignored). This channel will
	// close once all processes have exited (or, alternatively, when the
	// shutdown timeout has elapsed).
	Run(ctx context.Context, config config.Config) <-chan error

	// Shutdown will begin a graceful exit of all processes. This method
	// will block until the runner has exited (the channel from the Run
	// method has closed) or the given duration has elapsed. In the later
	// case a non-nil error is returned.
	Shutdown(time.Duration) error
}

type runner struct {
	processes           ProcessContainer
	services            service.ServiceContainer
	health              Health
	watcher             *processWatcher
	errChan             chan errMeta
	outChan             chan error
	wg                  *sync.WaitGroup
	logger              log.Logger
	clock               glock.Clock
	startupTimeout      time.Duration
	shutdownTimeout     time.Duration
	healthCheckInterval time.Duration
}

var _ Runner = &runner{}

type namedInjectable interface {
	Name() string
	LogFields() log.LogFields
	Wrapped() interface{}
}

type namedInitializer interface {
	Initializer
	Name() string
	LogFields() log.LogFields
	InitTimeout() time.Duration
}

type namedFinalizer interface {
	Initializer
	Name() string
	LogFields() log.LogFields
	FinalizeTimeout() time.Duration
	Wrapped() interface{}
}

// NewRunner creates a process runner from the given process and service
// containers.
func NewRunner(
	processes ProcessContainer,
	services service.ServiceContainer,
	health Health,
	runnerConfigs ...RunnerConfigFunc,
) Runner {
	errChan := make(chan errMeta)
	outChan := make(chan error, 1)

	r := &runner{
		processes:           processes,
		services:            services,
		health:              health,
		errChan:             errChan,
		outChan:             outChan,
		wg:                  &sync.WaitGroup{},
		logger:              log.NewNilLogger(),
		clock:               glock.NewRealClock(),
		healthCheckInterval: time.Second,
	}

	for _, f := range runnerConfigs {
		f(r)
	}

	// Create a watcher around the meta error channel (written to by
	// the runner) and the output channel (read by the boot process).
	// Pass our own logger and clock instances and the requested
	// shutdown timeout.

	r.watcher = newWatcher(
		errChan,
		outChan,
		withWatcherLogger(r.logger),
		withWatcherClock(r.clock),
		withWatcherShutdownTimeout(r.shutdownTimeout),
	)

	return r
}

func (r *runner) Run(ctx context.Context, config config.Config) <-chan error {
	// Start watching things before running anything. This ensures that
	// we start listening for shutdown requests and intercepted signals
	// as soon as anything starts being initialized.

	r.watcher.watch()

	// Run the initializers in sequence. If there were no errors, begin
	// initializing and running processes in priority/registration order.

	_ = r.runInitializers(ctx, config) && r.runProcesses(ctx, config)
	return r.outChan
}

func (r *runner) Shutdown(timeout time.Duration) error {
	r.watcher.halt()

	select {
	case <-r.clock.After(timeout):
		return fmt.Errorf("process runner did not shutdown within timeout")
	case <-r.watcher.done:
		return nil
	}
}

//
// Running and Watching

func (r *runner) runInitializers(ctx context.Context, config config.Config) bool {
	r.logger.Info("Running initializers")

	for i, initializer := range r.processes.GetInitializers() {
		if err := r.inject(initializer); err != nil {
			_ = r.unwindInitializers(ctx, i)
			r.errChan <- errMeta{err: err, source: initializer}
			close(r.errChan)
			return false
		}

		if err := r.initWithTimeout(ctx, initializer, config); err != nil {
			_ = r.unwindInitializers(ctx, i)
			// Parallel initializers may return multiple errors, so
			// we return all of them here. This check if asymmetric
			// as there is no equivalent for processes.
			for _, err := range coerceToSet(err, initializer) {
				r.errChan <- err
			}

			close(r.errChan)
			return false
		}
	}

	return true
}

func (r *runner) runFinalizers(ctx context.Context, beforeIndex int) bool {
	r.logger.Info("Running finalizers")

	success := true
	for i := beforeIndex - 1; i >= 0; i-- {
		for _, process := range r.processes.GetProcessesAtPriorityIndex(i) {
			if err := r.finalizeWithTimeout(ctx, process); err != nil {
				r.errChan <- errMeta{err: err, source: process}
				success = false
			}
		}
	}

	return r.unwindInitializers(ctx, r.processes.NumInitializers()) && success
}

func (r *runner) unwindInitializers(ctx context.Context, beforeIndex int) bool {
	success := true
	initializers := r.processes.GetInitializers()

	for i := beforeIndex - 1; i >= 0; i-- {
		if err := r.finalizeWithTimeout(ctx, initializers[i]); err != nil {
			// Parallel initializers may return multiple errors, so
			// we return all of them here. This check if asymmetric
			// as there is no equivalent for processes.
			for _, err := range coerceToSet(err, initializers[i]) {
				r.errChan <- err
			}

			success = false
		}
	}

	return success
}

func (r *runner) runProcesses(ctx context.Context, config config.Config) bool {
	r.logger.Info("Running processes")

	if !r.injectProcesses() {
		return false
	}

	// For each priority index, attempt to initialize the processes
	// in sequence. Then, start all processes in a goroutine. If there
	// is any synchronous error occurs (either due to an Init call
	// returning a non-nil error, or the watcher has begun shutdown),
	// stop booting up processes adn simply wait for them to spin down.

	success := true
	index := 0
	for ; index < r.processes.NumPriorities(); index++ {
		if !r.initProcessesAtPriorityIndex(ctx, config, index) {
			success = false
			break
		}

		if !r.startProcessesAtPriorityIndex(ctx, index) {
			success = false
			break
		}
	}

	// Wait for all booted processes to exit any Start/Stop methods,
	// then run all the initializers with finalize methods in their
	// reverse startup order. After all possible writes to the error
	// channel have occurred, close it to signal to the watcher to
	// do its own cleanup (and close the output channel).

	go func() {
		r.wg.Wait()
		_ = r.runFinalizers(ctx, index)
		close(r.errChan)
	}()

	if !success {
		return false
	}

	r.logger.Info("All processes have started")
	return true
}

//
// Injection

func (r *runner) injectProcesses() bool {
	for i := 0; i < r.processes.NumPriorities(); i++ {
		for _, process := range r.processes.GetProcessesAtPriorityIndex(i) {
			if err := r.inject(process); err != nil {
				r.errChan <- errMeta{err: err, source: process}
				close(r.errChan)
				return false
			}
		}
	}

	return true
}

func (r *runner) inject(injectable namedInjectable) error {
	r.logger.WithFields(injectable.LogFields()).Info("Injecting services into %s", injectable.Name())

	if err := inject(injectable, r.services, r.logger); err != nil {
		return fmt.Errorf(
			"failed to inject services into %s (%s)",
			injectable.Name(),
			err.Error(),
		)
	}

	return nil
}

// inject will inject the given injectable with services. The service container
// is first modified via overlay so that the logger is tagged with the service
// name and any additional logging fields registered to the service.
func inject(injectable namedInjectable, services service.ServiceContainer, logger log.Logger) error {
	// Tag the logger with any registered log fields
	logger = logger.WithFields(injectable.LogFields())

	// Create an overlay service map replacing `logger` and `services` keys
	serviceMap := map[string]interface{}{"logger": logger}
	overlayServices := service.Overlay(services, serviceMap)
	serviceMap["services"] = overlayServices

	// Inject the services
	return overlayServices.Inject(injectable.Wrapped())
}

//
// Initialization

func (r *runner) initProcessesAtPriorityIndex(ctx context.Context, config config.Config, index int) bool {
	r.logger.Info("Initializing processes at priority index %d", index)

	for _, process := range r.processes.GetProcessesAtPriorityIndex(index) {
		if err := r.initWithTimeout(ctx, process, config); err != nil {
			r.errChan <- errMeta{err: err, source: process}
			return false
		}
	}

	return true
}

func (r *runner) initWithTimeout(ctx context.Context, initializer namedInitializer, config config.Config) error {
	// Run the initializer in a goroutine. We don't want to block
	// on this in case we want to abandon reading from this channel
	// (timeout or shutdown). This is only true for initializer
	// methods (will not be true for process Start methods).

	errChan := makeErrChan(func() error {
		return r.init(ctx, initializer, config)
	})

	// Construct a timeout chan for the init (if timeout is set to
	// zero, this chan is nil and will never yield a value).

	initTimeoutChan := r.makeTimeoutChan(initializer.InitTimeout())

	// Now, wait for one of three results:
	//   - Init completed, return its value
	//   - Initialization took too long, return an error
	//   - Watcher is shutting down, ignore the return value

	select {
	case err := <-errChan:
		return err

	case <-initTimeoutChan:
		return fmt.Errorf("%s did not initialize within timeout", initializer.Name())

	case <-r.watcher.shutdownSignal:
		return fmt.Errorf("aborting initialization of %s", initializer.Name())
	}
}

func (r *runner) init(ctx context.Context, initializer namedInitializer, config config.Config) error {
	r.logger.WithFields(initializer.LogFields()).Info("Initializing %s", initializer.Name())

	if err := initializer.Init(ctx, config); err != nil {
		if _, ok := err.(errMetaSet); ok {
			// Pass error sets up unchanged
			return err
		}

		return fmt.Errorf(
			"failed to initialize %s (%s)",
			initializer.Name(),
			err.Error(),
		)
	}

	r.logger.WithFields(initializer.LogFields()).Info("Initialized %s", initializer.Name())
	return nil
}

//
// Finalization

func (r *runner) finalizeWithTimeout(ctx context.Context, initializer namedFinalizer) error {
	// Similar to initWithTimeout, run the finalizer in a goroutine
	// and either return the error result or return an error value
	// if the finalizer took too long.

	errChan := makeErrChan(func() error {
		return r.finalize(ctx, initializer)
	})

	finalizeTimeoutChan := r.makeTimeoutChan(initializer.FinalizeTimeout())

	select {
	case err := <-errChan:
		return err

	case <-finalizeTimeoutChan:
		return fmt.Errorf("%s did not finalize within timeout", initializer.Name())
	}
}

func (r *runner) finalize(ctx context.Context, initializer namedFinalizer) error {
	// Finalizer is an optional interface on Initializer. Skip
	// if this initializer doesn't conform.
	finalizer, ok := initializer.Wrapped().(Finalizer)
	if !ok {
		return nil
	}

	r.logger.WithFields(initializer.LogFields()).Info("Finalizing %s", initializer.Name())

	if err := finalizer.Finalize(ctx); err != nil {
		if _, ok := err.(errMetaSet); ok {
			// Pass error sets up unchanged
			return err
		}

		return fmt.Errorf(
			"%s returned error from finalize (%s)",
			initializer.Name(),
			err.Error(),
		)
	}

	r.logger.WithFields(initializer.LogFields()).Info("Finalized %s", initializer.Name())
	return nil
}

//
// Process Starting

func (r *runner) startProcessesAtPriorityIndex(ctx context.Context, index int) bool {
	r.logger.Info("Starting processes at priority index %d", index)

	// For each process group, we create a goroutine that will shutdown
	// all processes once the watcher begins shutting down. We add one to
	// the wait group to "bridge the gap" between the exit of the start
	// methods and the call to a stop method -- this situation is likely
	// rare, but would cause a panic.

	r.wg.Add(1)

	go func() {
		defer r.wg.Done()
		<-r.watcher.shutdownSignal
		r.stopProcessesAtPriorityIndex(ctx, index)
	}()

	// Create an abandon channel that closes to signal the routine invoking
	// the Start method of a process to ignore its return value -- we really
	// should not allow a timed-out start method to block the entire process.
	abandonSignal := make(chan struct{})

	// Actually start each process. Each call to startProcess blocks until
	// the process exits, so we perform each one in a goroutine guarded by
	// the runner's wait group.

	for _, process := range r.processes.GetProcessesAtPriorityIndex(index) {
		r.wg.Add(1)

		go func(p *ProcessMeta) {
			defer r.wg.Done()
			r.startProcess(ctx, p, abandonSignal)
		}(process)
	}

	// If initializers did not set a health description then we can assume that
	// the start method will immediately begin useful work and we don't need to
	// monitor it for an ok signal.
	if len(r.getHealthDescriptions()) == 0 {
		r.logger.Info("All processes at priority index %d have reported healthy", index)
		return true
	}

	// Otherwise, we'll keep re-checking the health descriptions until the list
	// goes empty, or the startup timeout elapses. The timeout is calculated from
	// the minimum timeout values of the runner and each process at this priority
	// index. If no such values are set, then the startup timeout channel is nil
	// and will never yield a value.
	startupTimeoutChan := r.makeTimeoutChan(r.startupTimeoutForPriorityIndex(index))

	for {
		select {
		case <-r.clock.After(r.healthCheckInterval):
			if descriptions := r.getHealthDescriptions(); len(descriptions) != 0 {
				r.logger.Warning("Process is not yet healthy - outstanding reasons: %s", strings.Join(descriptions, ", "))
				continue
			}

			r.logger.Info("All processes at priority index %d have reported healthy", index)
			return true

		case <-startupTimeoutChan:
			if descriptions := r.getHealthDescriptions(); len(descriptions) != 0 {
				r.errChan <- errMeta{err: fmt.Errorf(
					"processes at priority index %d did not become healthy within timeout - outstanding reasons: %s",
					index,
					strings.Join(descriptions, ", "),
				)}

				close(abandonSignal)
				return false
			}

			r.logger.Info("All processes at priority index %d have reported healthy", index)
			return true
		}
	}
}

func (r *runner) startupTimeoutForPriorityIndex(index int) time.Duration {
	timeout := r.startupTimeout

	for _, process := range r.processes.GetProcessesAtPriorityIndex(index) {
		if process.startTimeout != 0 && (timeout == 0 || process.startTimeout < timeout) {
			timeout = process.startTimeout
		}
	}

	return timeout
}

func (r *runner) getHealthDescriptions() []string {
	descriptions := []string{}
	for _, reason := range r.health.Reasons() {
		descriptions = append(descriptions, fmt.Sprintf("%s", reason.Key))
	}

	return descriptions
}

func (r *runner) startProcess(ctx context.Context, process *ProcessMeta, abandonSignal <-chan struct{}) {
	r.logger.WithFields(process.LogFields()).Info("Starting %s", process.Name())

	// Run the start method in a goroutine. We need to do
	// this as we assume all processes are long-running
	// and need to read from other sources for shutdown
	// and timeout behavior.

	errChan := makeErrChan(func() error {
		return process.Start(ctx)
	})

	// Create a channel for the shutdown timeout. This
	// channel will close only after the timeout duration
	// elapses AFTER the stop method of the process is
	// called. If the shutdown timeout is set to zero, this
	// channel will remain nil and will never yield.

	var shutdownTimeout chan (struct{})

	if process.shutdownTimeout > 0 {
		shutdownTimeout = make(chan struct{})

		go func() {
			<-process.stopped
			<-r.clock.After(process.shutdownTimeout)
			close(shutdownTimeout)
		}()
	}

	// Now, wait for the Start method to yield, in which case the error value
	// is passed to the watcher, or for either the abandon channel or timeout
	// channel to signal, in which case we abandon the reading of the return
	// value from the Start method.

	select {
	case <-abandonSignal:
		r.logger.WithFields(process.LogFields()).Error("Abandoning result of %s", process.Name())
		return

	case err := <-errChan:
		if err != nil {
			wrappedErr := fmt.Errorf(
				"%s returned a fatal error (%s)",
				process.Name(),
				err.Error(),
			)

			r.errChan <- errMeta{err: wrappedErr, source: process}
		} else {
			r.errChan <- errMeta{err: nil, source: process, silentExit: process.silentExit}
		}

	case <-shutdownTimeout:
		wrappedErr := fmt.Errorf(
			"%s did not shutdown within timeout",
			process.Name(),
		)

		r.errChan <- errMeta{err: wrappedErr, source: process}
	}
}

//
// Process Stopping

func (r *runner) stopProcessesAtPriorityIndex(ctx context.Context, index int) {
	r.logger.Info("Stopping processes at priority index %d", index)

	// Call stop on all processes at this priority index in parallel. We
	// add one to the wait group for each routine to ensure that we do
	// not close the err channel until all possible error producers have
	// exited.

	for _, process := range r.processes.GetProcessesAtPriorityIndex(index) {
		r.wg.Add(1)

		go func(process *ProcessMeta) {
			defer r.wg.Done()

			if err := r.stop(ctx, process); err != nil {
				r.errChan <- errMeta{err: err, source: process}
			}
		}(process)
	}
}

func (r *runner) stop(ctx context.Context, process *ProcessMeta) error {
	r.logger.WithFields(process.LogFields()).Info("Stopping %s", process.Name())

	if err := process.Stop(ctx); err != nil {
		return fmt.Errorf(
			"%s returned error from stop (%s)",
			process.Name(),
			err.Error(),
		)
	}

	return nil
}

func (r *runner) makeTimeoutChan(timeout time.Duration) <-chan time.Time {
	return makeTimeoutChan(r.clock, timeout)
}
