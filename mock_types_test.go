// Code generated by go-mockgen 0.1.0; DO NOT EDIT.

package process

import (
	"context"
	"sync"
)

// MockMaximumProcess is a mock implementation of the maximumProcess
// interface (from the package github.com/go-nacelle/process) used for unit
// testing.
type MockMaximumProcess struct {
	// FinalizeFunc is an instance of a mock function object controlling the
	// behavior of the method Finalize.
	FinalizeFunc *MaximumProcessFinalizeFunc
	// InitFunc is an instance of a mock function object controlling the
	// behavior of the method Init.
	InitFunc *MaximumProcessInitFunc
	// StartFunc is an instance of a mock function object controlling the
	// behavior of the method Start.
	StartFunc *MaximumProcessStartFunc
	// StopFunc is an instance of a mock function object controlling the
	// behavior of the method Stop.
	StopFunc *MaximumProcessStopFunc
}

// NewMockMaximumProcess creates a new mock of the maximumProcess interface.
// All methods return zero values for all results, unless overwritten.
func NewMockMaximumProcess() *MockMaximumProcess {
	return &MockMaximumProcess{
		FinalizeFunc: &MaximumProcessFinalizeFunc{
			defaultHook: func(context.Context) error {
				return nil
			},
		},
		InitFunc: &MaximumProcessInitFunc{
			defaultHook: func(context.Context) error {
				return nil
			},
		},
		StartFunc: &MaximumProcessStartFunc{
			defaultHook: func(context.Context) error {
				return nil
			},
		},
		StopFunc: &MaximumProcessStopFunc{
			defaultHook: func(context.Context) error {
				return nil
			},
		},
	}
}

// surrogateMockMaximumProcess is a copy of the maximumProcess interface
// (from the package github.com/go-nacelle/process). It is redefined here as
// it is unexported in the source packge.
type surrogateMockMaximumProcess interface {
	Finalize(context.Context) error
	Init(context.Context) error
	Start(context.Context) error
	Stop(context.Context) error
}

// NewMockMaximumProcessFrom creates a new mock of the MockMaximumProcess
// interface. All methods delegate to the given implementation, unless
// overwritten.
func NewMockMaximumProcessFrom(i surrogateMockMaximumProcess) *MockMaximumProcess {
	return &MockMaximumProcess{
		FinalizeFunc: &MaximumProcessFinalizeFunc{
			defaultHook: i.Finalize,
		},
		InitFunc: &MaximumProcessInitFunc{
			defaultHook: i.Init,
		},
		StartFunc: &MaximumProcessStartFunc{
			defaultHook: i.Start,
		},
		StopFunc: &MaximumProcessStopFunc{
			defaultHook: i.Stop,
		},
	}
}

// MaximumProcessFinalizeFunc describes the behavior when the Finalize
// method of the parent MockMaximumProcess instance is invoked.
type MaximumProcessFinalizeFunc struct {
	defaultHook func(context.Context) error
	hooks       []func(context.Context) error
	history     []MaximumProcessFinalizeFuncCall
	mutex       sync.Mutex
}

// Finalize delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockMaximumProcess) Finalize(v0 context.Context) error {
	r0 := m.FinalizeFunc.nextHook()(v0)
	m.FinalizeFunc.appendCall(MaximumProcessFinalizeFuncCall{v0, r0})
	return r0
}

// SetDefaultHook sets function that is called when the Finalize method of
// the parent MockMaximumProcess instance is invoked and the hook queue is
// empty.
func (f *MaximumProcessFinalizeFunc) SetDefaultHook(hook func(context.Context) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Finalize method of the parent MockMaximumProcess instance invokes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *MaximumProcessFinalizeFunc) PushHook(hook func(context.Context) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *MaximumProcessFinalizeFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *MaximumProcessFinalizeFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context) error {
		return r0
	})
}

func (f *MaximumProcessFinalizeFunc) nextHook() func(context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *MaximumProcessFinalizeFunc) appendCall(r0 MaximumProcessFinalizeFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of MaximumProcessFinalizeFuncCall objects
// describing the invocations of this function.
func (f *MaximumProcessFinalizeFunc) History() []MaximumProcessFinalizeFuncCall {
	f.mutex.Lock()
	history := make([]MaximumProcessFinalizeFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// MaximumProcessFinalizeFuncCall is an object that describes an invocation
// of method Finalize on an instance of MockMaximumProcess.
type MaximumProcessFinalizeFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c MaximumProcessFinalizeFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c MaximumProcessFinalizeFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// MaximumProcessInitFunc describes the behavior when the Init method of the
// parent MockMaximumProcess instance is invoked.
type MaximumProcessInitFunc struct {
	defaultHook func(context.Context) error
	hooks       []func(context.Context) error
	history     []MaximumProcessInitFuncCall
	mutex       sync.Mutex
}

// Init delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockMaximumProcess) Init(v0 context.Context) error {
	r0 := m.InitFunc.nextHook()(v0)
	m.InitFunc.appendCall(MaximumProcessInitFuncCall{v0, r0})
	return r0
}

// SetDefaultHook sets function that is called when the Init method of the
// parent MockMaximumProcess instance is invoked and the hook queue is
// empty.
func (f *MaximumProcessInitFunc) SetDefaultHook(hook func(context.Context) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Init method of the parent MockMaximumProcess instance invokes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *MaximumProcessInitFunc) PushHook(hook func(context.Context) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *MaximumProcessInitFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *MaximumProcessInitFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context) error {
		return r0
	})
}

func (f *MaximumProcessInitFunc) nextHook() func(context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *MaximumProcessInitFunc) appendCall(r0 MaximumProcessInitFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of MaximumProcessInitFuncCall objects
// describing the invocations of this function.
func (f *MaximumProcessInitFunc) History() []MaximumProcessInitFuncCall {
	f.mutex.Lock()
	history := make([]MaximumProcessInitFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// MaximumProcessInitFuncCall is an object that describes an invocation of
// method Init on an instance of MockMaximumProcess.
type MaximumProcessInitFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c MaximumProcessInitFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c MaximumProcessInitFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// MaximumProcessStartFunc describes the behavior when the Start method of
// the parent MockMaximumProcess instance is invoked.
type MaximumProcessStartFunc struct {
	defaultHook func(context.Context) error
	hooks       []func(context.Context) error
	history     []MaximumProcessStartFuncCall
	mutex       sync.Mutex
}

// Start delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockMaximumProcess) Start(v0 context.Context) error {
	r0 := m.StartFunc.nextHook()(v0)
	m.StartFunc.appendCall(MaximumProcessStartFuncCall{v0, r0})
	return r0
}

// SetDefaultHook sets function that is called when the Start method of the
// parent MockMaximumProcess instance is invoked and the hook queue is
// empty.
func (f *MaximumProcessStartFunc) SetDefaultHook(hook func(context.Context) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Start method of the parent MockMaximumProcess instance invokes the hook
// at the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *MaximumProcessStartFunc) PushHook(hook func(context.Context) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *MaximumProcessStartFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *MaximumProcessStartFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context) error {
		return r0
	})
}

func (f *MaximumProcessStartFunc) nextHook() func(context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *MaximumProcessStartFunc) appendCall(r0 MaximumProcessStartFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of MaximumProcessStartFuncCall objects
// describing the invocations of this function.
func (f *MaximumProcessStartFunc) History() []MaximumProcessStartFuncCall {
	f.mutex.Lock()
	history := make([]MaximumProcessStartFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// MaximumProcessStartFuncCall is an object that describes an invocation of
// method Start on an instance of MockMaximumProcess.
type MaximumProcessStartFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c MaximumProcessStartFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c MaximumProcessStartFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// MaximumProcessStopFunc describes the behavior when the Stop method of the
// parent MockMaximumProcess instance is invoked.
type MaximumProcessStopFunc struct {
	defaultHook func(context.Context) error
	hooks       []func(context.Context) error
	history     []MaximumProcessStopFuncCall
	mutex       sync.Mutex
}

// Stop delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockMaximumProcess) Stop(v0 context.Context) error {
	r0 := m.StopFunc.nextHook()(v0)
	m.StopFunc.appendCall(MaximumProcessStopFuncCall{v0, r0})
	return r0
}

// SetDefaultHook sets function that is called when the Stop method of the
// parent MockMaximumProcess instance is invoked and the hook queue is
// empty.
func (f *MaximumProcessStopFunc) SetDefaultHook(hook func(context.Context) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Stop method of the parent MockMaximumProcess instance invokes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *MaximumProcessStopFunc) PushHook(hook func(context.Context) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *MaximumProcessStopFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *MaximumProcessStopFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context) error {
		return r0
	})
}

func (f *MaximumProcessStopFunc) nextHook() func(context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *MaximumProcessStopFunc) appendCall(r0 MaximumProcessStopFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of MaximumProcessStopFuncCall objects
// describing the invocations of this function.
func (f *MaximumProcessStopFunc) History() []MaximumProcessStopFuncCall {
	f.mutex.Lock()
	history := make([]MaximumProcessStopFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// MaximumProcessStopFuncCall is an object that describes an invocation of
// method Stop on an instance of MockMaximumProcess.
type MaximumProcessStopFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c MaximumProcessStopFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c MaximumProcessStopFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}
