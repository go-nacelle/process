package process

import (
	"context"
	"fmt"
	"syscall"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/go-nacelle/config"
)

func TestWatcherNoErrors(t *testing.T) {
	errChan := make(chan errMeta)
	outChan := make(chan error)
	watcher := newWatcher(errChan, outChan)
	watcher.watch()

	// Nil errors do not go on out chan
	errChan <- errMeta{nil, makeNamedInitializer("a"), false}
	errChan <- errMeta{nil, makeNamedInitializer("b"), false}
	errChan <- errMeta{nil, makeNamedInitializer("c"), false}
	consistently(t, errorChanDoesNotReceive(outChan))

	// Closing err chan should shutdown watcher
	close(errChan)

	// Ensure we unblock
	eventually(t, errorChanClosed(outChan))
}

func TestWatcherFatalErrorBeginsShutdown(t *testing.T) {
	errChan := make(chan errMeta)
	outChan := make(chan error)
	watcher := newWatcher(errChan, outChan)
	watcher.watch()

	errChan <- errMeta{nil, makeNamedInitializer("a"), true}
	errChan <- errMeta{nil, makeNamedInitializer("b"), true}
	consistently(t, errorChanDoesNotReceive(outChan))
	consistently(t, structChanDoesNotReceive(watcher.shutdownSignal))

	errChan <- errMeta{fmt.Errorf("oops"), makeNamedInitializer("c"), true}
	eventually(t, errorChanReceivesUnordered(outChan, "oops"))
	eventually(t, structChanClosed(watcher.shutdownSignal))
	consistently(t, errorChanDoesNotReceive(outChan))

	// Additional errors
	errChan <- errMeta{nil, makeNamedInitializer("a"), true}
	errChan <- errMeta{nil, makeNamedInitializer("b"), true}
	consistently(t, errorChanDoesNotReceive(outChan))

	// And the same behavior above applies
	close(errChan)
	eventually(t, errorChanClosed(outChan))
}

func TestWatcherNilErrorBeginsShutdown(t *testing.T) {
	errChan := make(chan errMeta)
	outChan := make(chan error)
	watcher := newWatcher(errChan, outChan)
	watcher.watch()

	errChan <- errMeta{nil, makeNamedInitializer("a"), false}
	eventually(t, structChanClosed(watcher.shutdownSignal))
	consistently(t, errorChanDoesNotReceive(outChan))

	// Cleanup
	close(errChan)
	eventually(t, errorChanClosed(outChan))
}

func TestWatcherSignals(t *testing.T) {
	errChan := make(chan errMeta)
	defer close(errChan)

	outChan := make(chan error)
	watcher := newWatcher(errChan, outChan)
	watcher.watch()

	// Try to ensure watcher is waiting on signal
	<-time.After(time.Millisecond * 100)

	// First signal
	syscall.Kill(syscall.Getpid(), shutdownSignals[0])
	eventually(t, structChanClosed(watcher.shutdownSignal))

	// Second signal
	consistentlyNot(t, structChanClosed(watcher.abortSignal))
	syscall.Kill(syscall.Getpid(), shutdownSignals[0])
	eventually(t, structChanClosed(watcher.abortSignal))
}

func TestWatcherExternalHaltRequestBeginsShutdown(t *testing.T) {
	errChan := make(chan errMeta)
	outChan := make(chan error)
	watcher := newWatcher(errChan, outChan)

	watcher.watch()
	watcher.halt()
	eventually(t, structChanClosed(watcher.shutdownSignal))

	// Cleanup
	close(errChan)
	eventually(t, errorChanClosed(outChan))
}

func TestWatcherShutdownTimeout(t *testing.T) {
	errChan := make(chan errMeta)
	defer close(errChan)

	outChan := make(chan error)
	clock := glock.NewMockClock()
	watcher := newWatcher(
		errChan,
		outChan,
		withWatcherClock(clock),
		withWatcherShutdownTimeout(time.Second*10),
	)

	watcher.watch()
	watcher.halt()

	consistentlyNot(t, errorChanClosed(outChan))
	clock.BlockingAdvance(time.Second * 10)
	eventually(t, errorChanClosed(outChan))
}

//
//

func makeNamedInitializer(name string) namedInitializer {
	initializer := InitializerFunc(func(ctx context.Context, config config.Config) error {
		return nil
	})

	meta := newInitializerMeta(initializer)
	meta.name = name
	return meta
}
