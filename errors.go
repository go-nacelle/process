package process

import (
	"errors"
	"fmt"
)

// ErrUnexpectedReturn occurs when a process returns from Run before the process
// runner enters shutdown. This error does not occur if the process is marked with
// early exit.
var ErrUnexpectedReturn = errors.New("unexpected return from process")

// ErrStartupTimeout occurs when the health items associated with a process do not
// report healthy within the configured timeout.
var ErrStartupTimeout = errors.New("process did not become healthy within timeout")

// ErrShutdownTimeout occurs when a process does not return from Run after a call
// to Stop and a context cancellation within the configured timeout.
var ErrShutdownTimeout = errors.New("process refusing to shut down; abandoning goroutine")

// ErrHealthCheckCanceled occurs when a process's initial health check is cancelled
// due to another process exiting in a non-healthy way.
var ErrHealthCheckCanceled = errors.New("health check canceled")

// ErrHealthComponentAlreadyRegistered occurs when a health component is registered
// with the name of previously registered health component.
var ErrHealthComponentAlreadyRegistered = errors.New("health component already registered")

type opError struct {
	source   error
	metaName string
	opName   string
	message  string
}

func (e opError) Error() string {
	suffix := ""
	if e.source != nil {
		suffix = fmt.Sprintf(" (%s)", e.source)
	}

	return fmt.Sprintf("%s: %s %s%s", e.metaName, e.opName, e.message, suffix)
}

func (e opError) Unwrap() error {
	return e.source
}
