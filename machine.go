package process

import (
	"context"
	"sync"
)

type machine struct {
	runFunc      streamErrorFunc
	shutdownFunc streamErrorFunc
	wg           sync.WaitGroup
	ch           chan<- error
	closeOnce    sync.Once
}

// newMachine creates a new machine instance with the given run and shutdown functions. Errors
// occurring from either function will be sent on the given channel. Ownership of the channel
// is transferred and will be closed after the run function (and shutdown function, if invoked)
// have returned.
func newMachine(runFunc streamErrorFunc, shutdownFunc streamErrorFunc, ch chan<- error) *machine {
	return &machine{
		runFunc:      runFunc,
		shutdownFunc: shutdownFunc,
		ch:           ch,
	}
}

// run calls the configured runFunc in a separate goroutine. If the function returns an error,
// it will be sent on the configured channel.
func (m *machine) run(ctx context.Context) {
	m.runAsync(ctx, m.runFunc)
}

// shutdown calls the configured shutdownFunc in a separate goroutine. If the function returns
// an error, it will be sent on the configured channel.
func (m *machine) shutdown(ctx context.Context) {
	m.runAsync(ctx, m.shutdownFunc)
}

func (m *machine) runAsync(ctx context.Context, fn streamErrorFunc) {
	m.wg.Add(1)

	go func() {
		m.closeOnce.Do(func() {
			m.wg.Wait()
			close(m.ch)
		})
	}()

	go func() {
		defer m.wg.Done()

		for err := range fn(ctx) {
			m.ch <- err
		}
	}()
}
