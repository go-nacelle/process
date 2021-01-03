package process

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/go-nacelle/service"
)

func TestRunnerRunOrder(t *testing.T) {
	services := service.NewServiceContainer()
	processes := NewProcessContainer()
	health := NewHealth()
	runner := NewRunner(processes, services, health)
	init := make(chan string)
	finalize := make(chan string)
	start := make(chan string)
	stop := make(chan string)

	i1 := newTaggedFinalizer(init, finalize, "a")
	i2 := newTaggedInitializer(init, "b")
	i3 := newTaggedFinalizer(init, finalize, "c")
	p1 := newTaggedProcess(init, start, stop, "d")
	p2 := newTaggedProcess(init, start, stop, "e")
	p3 := newTaggedProcess(init, start, stop, "f")
	p4 := newTaggedProcessFinalizer(*newTaggedProcess(init, start, stop, "g"), finalize)
	p5 := newTaggedProcess(init, start, stop, "h")

	// Register things
	processes.RegisterInitializer(i1)
	processes.RegisterInitializer(i2)
	processes.RegisterInitializer(i3)
	processes.RegisterProcess(p1)
	processes.RegisterProcess(p2, WithProcessPriority(5))
	processes.RegisterProcess(p3, WithProcessPriority(5))
	processes.RegisterProcess(p4, WithProcessPriority(3))
	processes.RegisterProcess(p5)

	errChan := make(chan error)
	shutdownChan := make(chan error)

	go func() {
		defer close(errChan)

		for err := range runner.Run(context.Background()) {
			errChan <- err
		}
	}()

	// Initializers
	eventually(t, stringChanReceivesOrdered(init, "a", "b", "c"))

	// Priority index 0
	eventually(t, stringChanReceivesUnordered(init, "d", "h"))

	// May start in either order
	eventually(t, stringChanReceivesUnordered(start, "d", "h"))

	// Priority index 1
	eventually(t, stringChanReceivesUnordered(init, "g"))
	eventually(t, stringChanReceivesUnordered(start, "g"))

	// Priority index 2
	eventually(t, stringChanReceivesUnordered(init, "e", "f"))

	// May start in either order
	eventually(t, stringChanReceivesUnordered(start, "e", "f"))

	go func() {
		defer close(shutdownChan)
		shutdownChan <- runner.Shutdown(time.Minute)
	}()

	// May stop in any order
	eventually(t, stringChanReceivesUnordered(stop, "d", "e", "f", "g", "h"))

	// Finalizers
	eventually(t, stringChanReceivesUnordered(finalize, "g", "a", "c"))

	// Ensure unblocked
	eventually(t, errorChanClosed(shutdownChan))
	eventually(t, errorChanClosed(errChan))
}

func TestRunnerEarlyExit(t *testing.T) {
	services := service.NewServiceContainer()
	processes := NewProcessContainer()
	health := NewHealth()
	runner := NewRunner(processes, services, health)
	init := make(chan string)
	start := make(chan string)
	stop := make(chan string)
	errChan := make(chan error)

	p1 := newTaggedProcess(init, start, stop, "a")
	p2 := newTaggedProcess(init, start, stop, "b")

	// Register things
	processes.RegisterProcess(p1)
	processes.RegisterProcess(p2)

	go func() {
		defer close(errChan)

		for err := range runner.Run(context.Background()) {
			errChan <- err
		}
	}()

	eventually(t, stringChanReceivesUnordered(init, "a", "b"))
	eventually(t, stringChanReceivesN(start, 2))

	go p2.Stop(context.Background())

	// Stopping one process should shutdown the rest
	eventually(t, stringChanReceivesUnordered(stop, "b"))
	eventually(t, stringChanReceivesUnordered(stop, "a", "b"))
	eventually(t, errorChanClosed(errChan))
}

func TestRunnerSilentExit(t *testing.T) {
	services := service.NewServiceContainer()
	processes := NewProcessContainer()
	health := NewHealth()
	runner := NewRunner(processes, services, health)
	init := make(chan string)
	start := make(chan string)
	stop := make(chan string)

	p1 := newTaggedProcess(init, start, stop, "a")
	p2 := newTaggedProcess(init, start, stop, "b")

	// Register things
	processes.RegisterProcess(p1)
	processes.RegisterProcess(p2, WithSilentExit())

	go runner.Run(context.Background())

	eventually(t, stringChanReceivesUnordered(init, "a", "b"))
	eventually(t, stringChanReceivesN(start, 2))

	go p2.Stop(context.Background())

	eventually(t, stringChanReceivesUnordered(stop, "b"))
}

func TestRunnerShutdownTimeout(t *testing.T) {
	services := service.NewServiceContainer()
	processes := NewProcessContainer()
	health := NewHealth()
	clock := glock.NewMockClock()
	runner := NewRunner(processes, services, health, WithClock(clock))
	sync := make(chan struct{})
	process := newBlockingProcess(sync)
	errChan := make(chan error)
	shutdownChan := make(chan error)

	// Register things
	processes.RegisterProcess(process)

	go func() {
		defer close(errChan)

		for err := range runner.Run(context.Background()) {
			errChan <- err
		}
	}()

	eventually(t, structChanClosed(sync))

	go func() {
		defer close(shutdownChan)
		shutdownChan <- runner.Shutdown(time.Minute)
	}()

	clock.BlockingAdvance(time.Minute)
	eventually(t, errorChanReceivesUnordered(shutdownChan, "process runner did not shutdown within timeout"))
}

func TestRunnerProcessStartTimeout(t *testing.T) {
	services := service.NewServiceContainer()
	processes := NewProcessContainer()
	health := NewHealth()
	clock := glock.NewMockClock()
	init := make(chan string)
	start := make(chan string)
	stop := make(chan string)
	errChan := make(chan error)
	runner := NewRunner(
		processes,
		services,
		health,
		WithClock(clock),
		WithStartTimeout(time.Minute),
	)

	// Stop the process from going healthy
	health.AddReason("oops1")
	health.AddReason("oops2")

	p1 := newTaggedProcess(init, start, stop, "a")
	p2 := newTaggedProcess(init, start, stop, "b")

	processes.RegisterProcess(
		p1,
		WithProcessName("a"),
		WithProcessStartTimeout(time.Second*30),
	)

	processes.RegisterProcess(
		p2,
		WithProcessName("b"),
		WithProcessStartTimeout(time.Second*45),
	)

	go func() {
		defer close(errChan)

		for err := range runner.Run(context.Background()) {
			errChan <- err
		}
	}()

	// Don't block startup
	eventually(t, stringChanReceivesN(init, 2))
	eventually(t, stringChanReceivesN(start, 2))

	// Ensure timeout is respected
	consistently(t, errorChanDoesNotReceive(errChan))
	clock.Advance(time.Second * 30)

	// Watcher should shut down
	eventually(t, stringChanReceivesN(stop, 2))

	// Check error message
	eventually(t, errorChanReceivesUnordered(errChan, "processes at priority index 0 did not become healthy within timeout - outstanding reasons: oops1, oops2"))
	eventually(t, errorChanClosed(errChan))
}

func TestRunnerProcessShutdownTimeout(t *testing.T) {
	services := service.NewServiceContainer()
	processes := NewProcessContainer()
	health := NewHealth()
	clock := glock.NewMockClock()
	runner := NewRunner(processes, services, health, WithClock(clock))
	init := make(chan string)
	start := make(chan string)
	stop := make(chan string)

	processes.RegisterProcess(
		newTaggedProcess(init, start, stop, "a"),
		WithProcessName("a"),
		WithProcessShutdownTimeout(time.Second*10),
	)

	errChan := make(chan error)
	shutdownChan := make(chan error)

	go func() {
		defer close(errChan)

		for err := range runner.Run(context.Background()) {
			errChan <- err
		}
	}()

	go func() {
		// Stupid flaky goroutine scheduling
		<-time.After(time.Millisecond * 100)

		defer close(shutdownChan)
		shutdownChan <- runner.Shutdown(time.Minute)
	}()

	eventually(t, stringChanReceivesN(init, 1))
	eventually(t, stringChanReceivesN(stop, 1))

	// Blocked on process start method
	consistently(t, errorChanDoesNotReceive(shutdownChan))
	clock.Advance(time.Second * 5)
	consistently(t, errorChanDoesNotReceive(shutdownChan))
	clock.Advance(time.Second * 5)

	// Unblock after timeout
	eventually(t, errorChanClosed(shutdownChan))
	eventually(t, errorChanReceivesUnordered(errChan, "a did not shutdown within timeout"))
	eventually(t, errorChanClosed(errChan))
}

func TestRunnerInitializerInjectionError(t *testing.T) {
	services := service.NewServiceContainer()
	processes := NewProcessContainer()
	health := NewHealth()
	runner := NewRunner(processes, services, health)
	init := make(chan string)
	start := make(chan string)
	stop := make(chan string)
	errChan := make(chan error)

	i1 := newTaggedInitializer(init, "a")
	i2 := newInitializerWithService()
	i3 := newTaggedInitializer(init, "c")
	p1 := newTaggedProcess(init, start, stop, "d")

	// Register things
	processes.RegisterInitializer(i1)
	processes.RegisterInitializer(i2, WithInitializerName("b"))
	processes.RegisterInitializer(i3)
	processes.RegisterProcess(p1)

	go func() {
		defer close(errChan)

		for err := range runner.Run(context.Background()) {
			errChan <- err
		}
	}()

	// Ensure error is encountered
	eventually(t, stringChanReceivesOrdered(init, "a"))
	consistently(t, stringChanDoesNotReceive(init))
	eventually(t, errorChanReceivesUnordered(errChan, "failed to inject services into b"))

	// Nothing else called
	consistently(t, stringChanDoesNotReceive(init))
	consistently(t, stringChanDoesNotReceive(start))
	eventually(t, errorChanClosed(errChan))
}

func TestRunnerProcessInjectionError(t *testing.T) {
	services := service.NewServiceContainer()
	processes := NewProcessContainer()
	health := NewHealth()
	runner := NewRunner(processes, services, health)
	init := make(chan string)
	start := make(chan string)
	stop := make(chan string)
	errChan := make(chan error)

	i1 := newTaggedInitializer(init, "a")
	i2 := newTaggedInitializer(init, "b")
	i3 := newTaggedInitializer(init, "c")
	p1 := newTaggedProcess(init, start, stop, "d")
	p2 := newTaggedProcess(init, start, stop, "e")
	p3 := newProcessWithService()
	p4 := newTaggedProcess(init, start, stop, "g")
	p5 := newTaggedProcess(init, start, stop, "h")

	// Register things
	processes.RegisterInitializer(i1)
	processes.RegisterInitializer(i2)
	processes.RegisterInitializer(i3)
	processes.RegisterProcess(p1)
	processes.RegisterProcess(p2, WithProcessPriority(2))
	processes.RegisterProcess(p3, WithProcessPriority(2), WithProcessName("f"))
	processes.RegisterProcess(p4, WithProcessPriority(2))
	processes.RegisterProcess(p5, WithProcessPriority(3))

	go func() {
		defer close(errChan)

		for err := range runner.Run(context.Background()) {
			errChan <- err
		}
	}()

	// Initializers
	eventually(t, stringChanReceivesOrdered(init, "a", "b", "c"))

	// All processes are injected before any are initialized
	consistently(t, stringChanDoesNotReceive(init))

	eventually(t, errorChanReceivesUnordered(errChan, "failed to inject services into f"))

	// Nothing else called
	consistently(t, stringChanDoesNotReceive(init))
	consistently(t, stringChanDoesNotReceive(start))
	consistently(t, stringChanDoesNotReceive(stop))
	eventually(t, errorChanClosed(errChan))
}

func TestRunnerInitializerInitTimeout(t *testing.T) {
	services := service.NewServiceContainer()
	processes := NewProcessContainer()
	health := NewHealth()
	clock := glock.NewMockClock()
	runner := NewRunner(processes, services, health, WithClock(clock))
	init := make(chan string)
	errChan := make(chan error)

	i1 := newTaggedInitializer(init, "a")
	i2 := newTaggedInitializer(init, "b")

	// Register things
	processes.RegisterInitializer(i1, WithInitializerName("a"))
	processes.RegisterInitializer(i2, WithInitializerName("b"), WithInitializerTimeout(time.Minute), WithInitializerPriority(1))

	go func() {
		defer close(errChan)

		for err := range runner.Run(context.Background()) {
			errChan <- err
		}
	}()

	// Don't read second value - this blocks i2.Init
	eventually(t, stringChanReceivesOrdered(init, "a"))

	// Ensure error / unblocked
	clock.BlockingAdvance(time.Minute)
	eventually(t, errorChanReceivesUnordered(errChan, "b did not initialize within timeout"))
	eventually(t, errorChanClosed(errChan))
}

func TestRunnerFinalizerFinalizeTimeout(t *testing.T) {
	services := service.NewServiceContainer()
	processes := NewProcessContainer()
	health := NewHealth()
	clock := glock.NewMockClock()
	runner := NewRunner(processes, services, health, WithClock(clock))
	init := make(chan string)
	finalize := make(chan string)
	start := make(chan string)
	stop := make(chan string)
	errChan := make(chan error)

	i1 := newTaggedFinalizer(init, finalize, "a")
	i2 := newTaggedFinalizer(init, finalize, "b")
	p1 := newTaggedProcess(init, start, stop, "c")

	// Register things
	processes.RegisterInitializer(i1, WithInitializerName("a"), WithFinalizerTimeout(time.Minute))
	processes.RegisterInitializer(i2, WithInitializerName("b"))
	processes.RegisterProcess(p1, WithProcessName("c"))

	go func() {
		defer close(errChan)

		for err := range runner.Run(context.Background()) {
			errChan <- err
		}
	}()

	eventually(t, stringChanReceivesOrdered(init, "a", "b", "c"))
	eventually(t, stringChanReceivesUnordered(start, "c"))

	// Shutdown
	go runner.Shutdown(0)
	eventually(t, stringChanReceivesN(stop, 1))

	// Finalize first initializer
	eventually(t, stringChanReceivesUnordered(finalize, "b"))
	consistently(t, errorChanDoesNotReceive(errChan))

	// Timeout second finalizer
	clock.BlockingAdvance(time.Minute)
	eventually(t, errorChanReceivesUnordered(errChan, "a did not finalize within timeout"))
	eventually(t, errorChanClosed(errChan))
}

func TestRunnerFinalizerError(t *testing.T) {
	services := service.NewServiceContainer()
	processes := NewProcessContainer()
	health := NewHealth()
	runner := NewRunner(processes, services, health)
	init := make(chan string)
	finalize := make(chan string)
	errChan := make(chan error)

	i1 := newTaggedFinalizer(init, finalize, "a")
	i2 := newTaggedFinalizer(init, finalize, "b")
	i3 := newTaggedFinalizer(init, finalize, "c")

	// Register things
	processes.RegisterInitializer(i1, WithInitializerName("a"))
	processes.RegisterInitializer(i2, WithInitializerName("b"))
	processes.RegisterInitializer(i3, WithInitializerName("c"))

	i1.finalizeErr = fmt.Errorf("oops x")
	i2.finalizeErr = fmt.Errorf("oops y")
	i3.finalizeErr = fmt.Errorf("oops z")

	go func() {
		defer close(errChan)

		for err := range runner.Run(context.Background()) {
			errChan <- err
		}
	}()

	eventually(t, stringChanReceivesN(init, 3))
	eventually(t, stringChanReceivesN(finalize, 3))

	// Stop should emit errors but continue running
	// the remaining finalizers.

	eventually(t, errorChanReceivesUnordered(
		errChan,
		"c returned error from finalize (oops z)",
		"b returned error from finalize (oops y)",
		"a returned error from finalize (oops x)",
	))
	eventually(t, errorChanClosed(errChan))
}

func TestRunnerProcessInitTimeout(t *testing.T) {
	services := service.NewServiceContainer()
	processes := NewProcessContainer()
	health := NewHealth()
	clock := glock.NewMockClock()
	runner := NewRunner(processes, services, health, WithClock(clock))
	init := make(chan string)
	start := make(chan string)
	stop := make(chan string)
	errChan := make(chan error)

	p1 := newTaggedProcess(init, start, stop, "a")
	p2 := newTaggedProcess(init, start, stop, "b")

	// Register things
	processes.RegisterProcess(p1, WithProcessName("a"))
	processes.RegisterProcess(p2, WithProcessName("b"), WithProcessInitTimeout(time.Minute), WithProcessPriority(1))

	go func() {
		defer close(errChan)

		for err := range runner.Run(context.Background()) {
			errChan <- err
		}
	}()

	// Only read lower priority process
	eventually(t, stringChanReceivesOrdered(init, "a"))
	eventually(t, stringChanReceivesOrdered(start, "a"))

	// Ensure error
	clock.BlockingAdvance(time.Minute)
	eventually(t, errorChanReceivesUnordered(errChan, "b did not initialize within timeout"))

	// Ensure unblocked
	eventually(t, stringChanReceivesOrdered(stop, "a"))
	eventually(t, errorChanClosed(errChan))
}

func TestRunnerInitializerError(t *testing.T) {
	services := service.NewServiceContainer()
	processes := NewProcessContainer()
	health := NewHealth()
	runner := NewRunner(processes, services, health)
	init := make(chan string)
	finalize := make(chan string)
	start := make(chan string)
	stop := make(chan string)
	errChan := make(chan error)

	i1 := newTaggedFinalizer(init, finalize, "a")
	i2 := newTaggedFinalizer(init, finalize, "b")
	i3 := newTaggedInitializer(init, "c")
	p1 := newTaggedProcess(init, start, stop, "d")

	i2.initErr = fmt.Errorf("oops")

	// Register things
	processes.RegisterInitializer(i1)
	processes.RegisterInitializer(i2, WithInitializerName("b"))
	processes.RegisterInitializer(i3)
	processes.RegisterProcess(p1)

	go func() {
		defer close(errChan)

		for err := range runner.Run(context.Background()) {
			errChan <- err
		}
	}()

	// Check run order
	eventually(t, stringChanReceivesOrdered(init, "a", "b"))
	eventually(t, stringChanReceivesUnordered(finalize, "a", "b"))

	// Ensure error is encountered
	eventually(t, errorChanReceivesUnordered(errChan, "failed to initialize b (oops)"))

	// Nothing else called
	consistently(t, stringChanDoesNotReceive(init))
	consistently(t, stringChanDoesNotReceive(start))
	consistently(t, stringChanDoesNotReceive(finalize))
	eventually(t, errorChanClosed(errChan))
}

func TestRunnerProcessInitError(t *testing.T) {
	services := service.NewServiceContainer()
	processes := NewProcessContainer()
	health := NewHealth()
	runner := NewRunner(processes, services, health)
	init := make(chan string)
	start := make(chan string)
	stop := make(chan string)
	errChan := make(chan error)

	i1 := newTaggedInitializer(init, "a")
	i2 := newTaggedInitializer(init, "b")
	i3 := newTaggedInitializer(init, "c")
	p1 := newTaggedProcess(init, start, stop, "d")
	p2 := newTaggedProcess(init, start, stop, "e")
	p3 := newTaggedProcess(init, start, stop, "f")
	p4 := newTaggedProcess(init, start, stop, "g")
	p5 := newTaggedProcess(init, start, stop, "h")

	// Register things
	processes.RegisterInitializer(i1)
	processes.RegisterInitializer(i2)
	processes.RegisterInitializer(i3)
	processes.RegisterProcess(p1)
	processes.RegisterProcess(p2, WithProcessPriority(2))
	processes.RegisterProcess(p3, WithProcessPriority(2), WithProcessName("f"))
	processes.RegisterProcess(p4, WithProcessPriority(2))
	processes.RegisterProcess(p5, WithProcessPriority(3))

	p3.initErr = fmt.Errorf("oops")

	go func() {
		defer close(errChan)

		for err := range runner.Run(context.Background()) {
			errChan <- err
		}
	}()

	// Initializers

	eventually(t, stringChanReceivesOrdered(init, "a", "b", "c"))

	// Lower-priority process
	eventually(t, stringChanReceivesOrdered(init, "d"))
	eventually(t, stringChanReceivesUnordered(start, "d"))

	// Ensure error is encountered
	eventually(t, stringChanReceivesUnordered(init, "e", "f", "g"))
	eventually(t, errorChanReceivesUnordered(errChan, "failed to initialize f (oops)"))
	consistently(t, stringChanDoesNotReceive(init))

	// Shutdown only things that started
	eventually(t, stringChanReceivesUnordered(stop, "d"))
	consistently(t, stringChanDoesNotReceive(stop))

	// Nothing else called
	consistently(t, stringChanDoesNotReceive(init))
	consistently(t, stringChanDoesNotReceive(stop))
	eventually(t, errorChanClosed(errChan))
}

func TestRunnerProcessStartError(t *testing.T) {
	services := service.NewServiceContainer()
	processes := NewProcessContainer()
	health := NewHealth()
	runner := NewRunner(processes, services, health)
	init := make(chan string)
	start := make(chan string)
	stop := make(chan string)

	i1 := newTaggedInitializer(init, "a")
	i2 := newTaggedInitializer(init, "b")
	i3 := newTaggedInitializer(init, "c")
	p1 := newTaggedProcess(init, start, stop, "d")
	p2 := newTaggedProcess(init, start, stop, "e")
	p3 := newTaggedProcess(init, start, stop, "f")
	p4 := newTaggedProcess(init, start, stop, "g")
	p5 := newTaggedProcess(init, start, stop, "h")

	// Register things
	processes.RegisterInitializer(i1)
	processes.RegisterInitializer(i2)
	processes.RegisterInitializer(i3)
	processes.RegisterProcess(p1)
	processes.RegisterProcess(p2, WithProcessPriority(2))
	processes.RegisterProcess(p3, WithProcessPriority(2), WithProcessName("f"))
	processes.RegisterProcess(p4, WithProcessPriority(2))
	processes.RegisterProcess(p5, WithProcessPriority(3), WithProcessName("h"))

	p3.startErr = fmt.Errorf("oops")

	errChan := make(chan error)

	go func() {
		defer close(errChan)

		for err := range runner.Run(context.Background()) {
			errChan <- err
		}
	}()

	// Initializers
	eventually(t, stringChanReceivesOrdered(init, "a", "b", "c"))

	// Lower-priority process
	eventually(t, stringChanReceivesUnordered(init, "d"))
	eventually(t, stringChanReceivesUnordered(start, "d"))

	// Higher-priority processes
	eventually(t, stringChanReceivesUnordered(init, "e", "f", "g"))
	eventually(t, stringChanReceivesUnordered(start, "e", "f", "g"))

	// Shutdown everything that's started
	eventually(t, stringChanReceivesUnordered(stop, "d", "e", "f", "g"))
	consistently(t, stringChanDoesNotReceive(stop))

	// We get a start error from a goroutine, which means that
	// the next priority may be initializing. Since we're blocked
	// there, we should get a message saying that we're ignoring
	// the value of that process as well as the error from the
	// failing process.

	eventually(t, errorChanReceivesUnordered(
		errChan,
		"aborting initialization of h",
		"f returned a fatal error (oops)",
	))
	eventually(t, errorChanClosed(errChan))
}

func TestRunnerProcessStopError(t *testing.T) {
	services := service.NewServiceContainer()
	processes := NewProcessContainer()
	health := NewHealth()
	runner := NewRunner(processes, services, health)
	init := make(chan string)
	start := make(chan string)
	stop := make(chan string)
	errChan := make(chan error)

	i1 := newTaggedInitializer(init, "a")
	i2 := newTaggedInitializer(init, "b")
	i3 := newTaggedInitializer(init, "c")
	p1 := newTaggedProcess(init, start, stop, "d")
	p2 := newTaggedProcess(init, start, stop, "e")
	p3 := newTaggedProcess(init, start, stop, "f")
	p4 := newTaggedProcess(init, start, stop, "g")
	p5 := newTaggedProcess(init, start, stop, "h")

	// Register things
	processes.RegisterInitializer(i1)
	processes.RegisterInitializer(i2)
	processes.RegisterInitializer(i3)
	processes.RegisterProcess(p1, WithProcessName("d"))
	processes.RegisterProcess(p2, WithProcessName("e"), WithProcessPriority(5))
	processes.RegisterProcess(p3, WithProcessName("f"), WithProcessPriority(5))
	processes.RegisterProcess(p4, WithProcessName("g"), WithProcessPriority(3))
	processes.RegisterProcess(p5, WithProcessName("h"))

	p1.stopErr = fmt.Errorf("oops x")
	p3.stopErr = fmt.Errorf("oops y")
	p5.stopErr = fmt.Errorf("oops z")

	go func() {
		defer close(errChan)

		for err := range runner.Run(context.Background()) {
			errChan <- err
		}
	}()

	eventually(t, stringChanReceivesN(init, 8))
	eventually(t, stringChanReceivesN(start, 5))

	// Shutdown
	go runner.Shutdown(time.Minute)

	eventually(t, stringChanReceivesN(stop, 5))

	// Stop should emit errors but not block the progress
	// of the runner in any significant way (stop is only
	// called on shutdown, so cannot be double-fatal).

	eventually(t, errorChanReceivesUnordered(
		errChan,
		"d returned error from stop (oops x)",
		"f returned error from stop (oops y)",
		"h returned error from stop (oops z)",
	))
	eventually(t, errorChanClosed(errChan))
}

func TestContext(t *testing.T) {
	services := service.NewServiceContainer()
	processes := NewProcessContainer()
	health := NewHealth()
	runner := NewRunner(processes, services, health)

	p1 := newContextProcess()
	p2 := newContextProcess()
	p3 := newContextProcess()

	// Register things
	processes.RegisterProcess(p1, WithProcessPriority(1))
	processes.RegisterProcess(p2, WithProcessPriority(2))
	processes.RegisterProcess(p3, WithProcessPriority(3))

	errChan := make(chan error)
	shutdownChan := make(chan error)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		defer close(errChan)

		for err := range runner.Run(ctx) {
			errChan <- err
		}
	}()

	cancelCtx()

	go func() {
		defer close(shutdownChan)
		shutdownChan <- runner.Shutdown(time.Minute)
	}()

	// Ensure unblocked
	eventually(t, errorChanClosed(shutdownChan))
	eventually(t, errorChanClosed(errChan))
}

//
//

type taggedInitializer struct {
	name    string
	init    chan<- string
	initErr error
}

func newTaggedInitializer(init chan<- string, name string) *taggedInitializer {
	return &taggedInitializer{
		name: name,
		init: init,
	}
}

func (i *taggedInitializer) Init(ctx context.Context) error {
	i.init <- i.name
	return i.initErr
}

//
//

type taggedFinalizer struct {
	taggedInitializer
	finalize    chan<- string
	finalizeErr error
}

func newTaggedFinalizer(init chan<- string, finalize chan<- string, name string) *taggedFinalizer {
	return &taggedFinalizer{
		taggedInitializer: taggedInitializer{
			name: name,
			init: init,
		},
		finalize: finalize,
	}
}

func (i *taggedFinalizer) Finalize(ctx context.Context) error {
	i.finalize <- i.name
	return i.finalizeErr
}

//
//

type taggedProcess struct {
	name  string
	init  chan<- string
	start chan<- string
	stop  chan<- string
	wait  chan struct{}

	initErr  error
	startErr error
	stopErr  error
}

func newTaggedProcess(init, start, stop chan<- string, name string) *taggedProcess {
	return &taggedProcess{
		name:  name,
		init:  init,
		start: start,
		stop:  stop,
		wait:  make(chan struct{}, 1), // Make this safe to close twice w/o blocking
	}
}

func (p *taggedProcess) Init(ctx context.Context) error {
	p.init <- p.name
	return p.initErr
}

func (p *taggedProcess) Start(ctx context.Context) error {
	p.start <- p.name

	if p.startErr != nil {
		return p.startErr
	}

	<-p.wait
	return nil
}

func (p *taggedProcess) Stop(ctx context.Context) error {
	p.stop <- p.name
	p.wait <- struct{}{}
	return p.stopErr
}

type taggedProcessFinalizer struct {
	taggedProcess
	finalize    chan<- string
	finalizeErr error
}

func newTaggedProcessFinalizer(taggedProcess taggedProcess, finalize chan<- string) *taggedProcessFinalizer {
	return &taggedProcessFinalizer{
		taggedProcess: taggedProcess,
		finalize:      finalize,
	}
}

func (i *taggedProcessFinalizer) Finalize(ctx context.Context) error {
	i.finalize <- i.name
	return i.finalizeErr
}

type blockingProcess struct {
	sync chan struct{}
	wait chan struct{}
}

func newBlockingProcess(sync chan struct{}) *blockingProcess {
	return &blockingProcess{
		sync: sync,
	}
}

func (p *blockingProcess) Init(ctx context.Context) error  { return nil }
func (p *blockingProcess) Start(ctx context.Context) error { close(p.sync); <-p.wait; return nil }
func (p *blockingProcess) Stop(ctx context.Context) error  { return nil }

//
//

type initializerWithService struct {
	X struct{} `service:"notset"`
}

type processWithService struct {
	X struct{} `service:"notset"`
}

func newInitializerWithService() *initializerWithService { return &initializerWithService{} }
func newProcessWithService() *processWithService         { return &processWithService{} }

func (i *initializerWithService) Init(ctx context.Context) error { return nil }
func (p *processWithService) Init(ctx context.Context) error     { return nil }
func (p *processWithService) Start(ctx context.Context) error    { return nil }
func (p *processWithService) Stop(ctx context.Context) error     { return nil }

//
//

type contextProcess struct {
}

func newContextProcess() *contextProcess {
	return &contextProcess{}
}

func (p *contextProcess) Init(ctx context.Context) error  { return nil }
func (p *contextProcess) Start(ctx context.Context) error { <-ctx.Done(); return nil }
func (p *contextProcess) Stop(ctx context.Context) error  { return nil }
