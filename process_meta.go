package process

import (
	"context"
	"sync"
	"time"

	"github.com/go-nacelle/log"
)

// ProcessMeta wraps a process with some package private
// fields.
type ProcessMeta struct {
	sync.RWMutex
	Process
	name            string
	logFields       log.LogFields
	priority        int
	silentExit      bool
	once            *sync.Once
	stopped         chan struct{}
	cancelCtx       func()
	initTimeout     time.Duration
	startTimeout    time.Duration
	shutdownTimeout time.Duration
	finalizeTimeout time.Duration
}

func newProcessMeta(process Process) *ProcessMeta {
	return &ProcessMeta{
		Process: process,
		once:    &sync.Once{},
		stopped: make(chan struct{}),
	}
}

// Name returns the name of the process.
func (m *ProcessMeta) Name() string {
	if m.name == "" {
		return "<unnamed>"
	}

	return m.name
}

// LogFields returns logging fields registered to this process.
func (m *ProcessMeta) LogFields() log.LogFields {
	return m.logFields
}

// InitTimeout returns the maximum timeout allowed for a call to
// the Init function. A zero value indicates no timeout.
func (m *ProcessMeta) InitTimeout() time.Duration {
	return m.initTimeout
}

// Stop wraps the underlying process's Stop method with a Once
// value in order to guarantee that the Stop method will not
// take effect multiple times.
func (m *ProcessMeta) Stop(ctx context.Context) (err error) {
	m.once.Do(func() {
		m.RLock()
		close(m.stopped)
		cancelCtx := m.cancelCtx
		m.RUnlock()

		err = m.Process.Stop(ctx)

		if cancelCtx != nil {
			cancelCtx()
		}
	})

	return
}

func (m *ProcessMeta) setCancelCtx(cancelCtx func()) {
	m.Lock()
	defer m.Unlock()

	select {
	case <-m.stopped:
		cancelCtx()
		return
	default:
		m.cancelCtx = cancelCtx
	}
}

// FinalizeTimeout returns the maximum timeout allowed for a call to
// the Finalize function. A zero value indicates no timeout.
func (m *ProcessMeta) FinalizeTimeout() time.Duration {
	return m.finalizeTimeout
}

// Wrapped returns the underlying process.
func (m *ProcessMeta) Wrapped() interface{} {
	return m.Process
}
