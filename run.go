package process

import (
	"context"
	"sync"
)

// State tracks the current state of application execution.
type State struct {
	machine      *machine
	errors       <-chan error
	shutdownOnce sync.Once
}

// Run builds a machine to invoke the initializers and processes registered to the given
// container. This method returns a state value that can be used to signal the application
// to begin shutdown, and to block until the active processes have exited.
func Run(ctx context.Context, container *Container, configs ...MachineConfigFunc) *State {
	machineBuilder := newMachineBuilder(configs...)
	runFunc := machineBuilder.buildRun(container)
	shutdownFunc := machineBuilder.buildShutdown(container)

	errors := make(chan error)
	machine := newMachine(runFunc, shutdownFunc, errors)
	machine.run(ctx)

	return &State{machine: machine, errors: errors}
}

// Wait blocks until all processes exit cleanly or until an error occurs during execution
// of the machine built via Run. If an error occurs, all other running processes are
// signalled to exit. This method unblocks once all processes have exited. This method
// returns a boolean flag indicating a clean exit.
func (s *State) Wait(ctx context.Context) bool {
	ok := true
	for range s.errors {
		ok = false
		s.Shutdown(ctx)
	}

	return ok
}

// Shutdown signals all running processes to exit.
func (s *State) Shutdown(ctx context.Context) {
	s.shutdownOnce.Do(func() {
		s.machine.shutdown(ctx)
	})
}
