package process

import (
	"sync"
	"time"

	"github.com/go-nacelle/log"
)

type (
	// ProcessMeta wraps a process with some package private
	// fields.
	ProcessMeta struct {
		Process
		name            string
		loggingFields   log.LogFields
		priority        int
		silentExit      bool
		once            *sync.Once
		stopped         chan struct{}
		initTimeout     time.Duration
		startTimeout    time.Duration
		shutdownTimeout time.Duration
	}
)

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

// LoggingFields returns logging fields registered to this process.
func (m *ProcessMeta) LoggingFields() log.LogFields {
	return m.loggingFields
}

// InitTimeout returns the maximum timeout allowed for a call to
// the Init function. A zero value indicates no timeout.
func (m *ProcessMeta) InitTimeout() time.Duration {
	return m.initTimeout
}

// Stop wraps the underlying process's Stop method with a Once
// value in order to guarantee that the Stop method will not
// take effect multiple times.
func (m *ProcessMeta) Stop() (err error) {
	m.once.Do(func() {
		close(m.stopped)
		err = m.Process.Stop()
	})

	return
}

// Wrapped returns the underlying process.
func (m *ProcessMeta) Wrapped() interface{} {
	return m.Process
}
