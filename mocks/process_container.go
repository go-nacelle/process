// Code generated by go-mockgen 0.1.0; DO NOT EDIT.

package mocks

import (
	process "github.com/go-nacelle/process"
	"sync"
)

// MockProcessContainer is a mock implementation of the ProcessContainer
// interface (from the package github.com/go-nacelle/process) used for unit
// testing.
type MockProcessContainer struct {
	// GetInitializersFunc is an instance of a mock function object
	// controlling the behavior of the method GetInitializers.
	GetInitializersFunc *ProcessContainerGetInitializersFunc
	// GetProcessesAtPriorityIndexFunc is an instance of a mock function
	// object controlling the behavior of the method
	// GetProcessesAtPriorityIndex.
	GetProcessesAtPriorityIndexFunc *ProcessContainerGetProcessesAtPriorityIndexFunc
	// NumInitializersFunc is an instance of a mock function object
	// controlling the behavior of the method NumInitializers.
	NumInitializersFunc *ProcessContainerNumInitializersFunc
	// NumPrioritiesFunc is an instance of a mock function object
	// controlling the behavior of the method NumPriorities.
	NumPrioritiesFunc *ProcessContainerNumPrioritiesFunc
	// NumProcessesFunc is an instance of a mock function object controlling
	// the behavior of the method NumProcesses.
	NumProcessesFunc *ProcessContainerNumProcessesFunc
	// RegisterInitializerFunc is an instance of a mock function object
	// controlling the behavior of the method RegisterInitializer.
	RegisterInitializerFunc *ProcessContainerRegisterInitializerFunc
	// RegisterProcessFunc is an instance of a mock function object
	// controlling the behavior of the method RegisterProcess.
	RegisterProcessFunc *ProcessContainerRegisterProcessFunc
}

// NewMockProcessContainer creates a new mock of the ProcessContainer
// interface. All methods return zero values for all results, unless
// overwritten.
func NewMockProcessContainer() *MockProcessContainer {
	return &MockProcessContainer{
		GetInitializersFunc: &ProcessContainerGetInitializersFunc{
			defaultHook: func() []*process.InitializerMeta {
				return nil
			},
		},
		GetProcessesAtPriorityIndexFunc: &ProcessContainerGetProcessesAtPriorityIndexFunc{
			defaultHook: func(int) []*process.ProcessMeta {
				return nil
			},
		},
		NumInitializersFunc: &ProcessContainerNumInitializersFunc{
			defaultHook: func() int {
				return 0
			},
		},
		NumPrioritiesFunc: &ProcessContainerNumPrioritiesFunc{
			defaultHook: func() int {
				return 0
			},
		},
		NumProcessesFunc: &ProcessContainerNumProcessesFunc{
			defaultHook: func() int {
				return 0
			},
		},
		RegisterInitializerFunc: &ProcessContainerRegisterInitializerFunc{
			defaultHook: func(process.Initializer, ...process.InitializerConfigFunc) {
				return
			},
		},
		RegisterProcessFunc: &ProcessContainerRegisterProcessFunc{
			defaultHook: func(process.Process, ...process.ProcessConfigFunc) {
				return
			},
		},
	}
}

// NewMockProcessContainerFrom creates a new mock of the
// MockProcessContainer interface. All methods delegate to the given
// implementation, unless overwritten.
func NewMockProcessContainerFrom(i process.ProcessContainer) *MockProcessContainer {
	return &MockProcessContainer{
		GetInitializersFunc: &ProcessContainerGetInitializersFunc{
			defaultHook: i.GetInitializers,
		},
		GetProcessesAtPriorityIndexFunc: &ProcessContainerGetProcessesAtPriorityIndexFunc{
			defaultHook: i.GetProcessesAtPriorityIndex,
		},
		NumInitializersFunc: &ProcessContainerNumInitializersFunc{
			defaultHook: i.NumInitializers,
		},
		NumPrioritiesFunc: &ProcessContainerNumPrioritiesFunc{
			defaultHook: i.NumPriorities,
		},
		NumProcessesFunc: &ProcessContainerNumProcessesFunc{
			defaultHook: i.NumProcesses,
		},
		RegisterInitializerFunc: &ProcessContainerRegisterInitializerFunc{
			defaultHook: i.RegisterInitializer,
		},
		RegisterProcessFunc: &ProcessContainerRegisterProcessFunc{
			defaultHook: i.RegisterProcess,
		},
	}
}

// ProcessContainerGetInitializersFunc describes the behavior when the
// GetInitializers method of the parent MockProcessContainer instance is
// invoked.
type ProcessContainerGetInitializersFunc struct {
	defaultHook func() []*process.InitializerMeta
	hooks       []func() []*process.InitializerMeta
	history     []ProcessContainerGetInitializersFuncCall
	mutex       sync.Mutex
}

// GetInitializers delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockProcessContainer) GetInitializers() []*process.InitializerMeta {
	r0 := m.GetInitializersFunc.nextHook()()
	m.GetInitializersFunc.appendCall(ProcessContainerGetInitializersFuncCall{r0})
	return r0
}

// SetDefaultHook sets function that is called when the GetInitializers
// method of the parent MockProcessContainer instance is invoked and the
// hook queue is empty.
func (f *ProcessContainerGetInitializersFunc) SetDefaultHook(hook func() []*process.InitializerMeta) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// GetInitializers method of the parent MockProcessContainer instance
// invokes the hook at the front of the queue and discards it. After the
// queue is empty, the default hook function is invoked for any future
// action.
func (f *ProcessContainerGetInitializersFunc) PushHook(hook func() []*process.InitializerMeta) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *ProcessContainerGetInitializersFunc) SetDefaultReturn(r0 []*process.InitializerMeta) {
	f.SetDefaultHook(func() []*process.InitializerMeta {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *ProcessContainerGetInitializersFunc) PushReturn(r0 []*process.InitializerMeta) {
	f.PushHook(func() []*process.InitializerMeta {
		return r0
	})
}

func (f *ProcessContainerGetInitializersFunc) nextHook() func() []*process.InitializerMeta {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ProcessContainerGetInitializersFunc) appendCall(r0 ProcessContainerGetInitializersFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of ProcessContainerGetInitializersFuncCall
// objects describing the invocations of this function.
func (f *ProcessContainerGetInitializersFunc) History() []ProcessContainerGetInitializersFuncCall {
	f.mutex.Lock()
	history := make([]ProcessContainerGetInitializersFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ProcessContainerGetInitializersFuncCall is an object that describes an
// invocation of method GetInitializers on an instance of
// MockProcessContainer.
type ProcessContainerGetInitializersFuncCall struct {
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 []*process.InitializerMeta
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c ProcessContainerGetInitializersFuncCall) Args() []interface{} {
	return []interface{}{}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c ProcessContainerGetInitializersFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// ProcessContainerGetProcessesAtPriorityIndexFunc describes the behavior
// when the GetProcessesAtPriorityIndex method of the parent
// MockProcessContainer instance is invoked.
type ProcessContainerGetProcessesAtPriorityIndexFunc struct {
	defaultHook func(int) []*process.ProcessMeta
	hooks       []func(int) []*process.ProcessMeta
	history     []ProcessContainerGetProcessesAtPriorityIndexFuncCall
	mutex       sync.Mutex
}

// GetProcessesAtPriorityIndex delegates to the next hook function in the
// queue and stores the parameter and result values of this invocation.
func (m *MockProcessContainer) GetProcessesAtPriorityIndex(v0 int) []*process.ProcessMeta {
	r0 := m.GetProcessesAtPriorityIndexFunc.nextHook()(v0)
	m.GetProcessesAtPriorityIndexFunc.appendCall(ProcessContainerGetProcessesAtPriorityIndexFuncCall{v0, r0})
	return r0
}

// SetDefaultHook sets function that is called when the
// GetProcessesAtPriorityIndex method of the parent MockProcessContainer
// instance is invoked and the hook queue is empty.
func (f *ProcessContainerGetProcessesAtPriorityIndexFunc) SetDefaultHook(hook func(int) []*process.ProcessMeta) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// GetProcessesAtPriorityIndex method of the parent MockProcessContainer
// instance invokes the hook at the front of the queue and discards it.
// After the queue is empty, the default hook function is invoked for any
// future action.
func (f *ProcessContainerGetProcessesAtPriorityIndexFunc) PushHook(hook func(int) []*process.ProcessMeta) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *ProcessContainerGetProcessesAtPriorityIndexFunc) SetDefaultReturn(r0 []*process.ProcessMeta) {
	f.SetDefaultHook(func(int) []*process.ProcessMeta {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *ProcessContainerGetProcessesAtPriorityIndexFunc) PushReturn(r0 []*process.ProcessMeta) {
	f.PushHook(func(int) []*process.ProcessMeta {
		return r0
	})
}

func (f *ProcessContainerGetProcessesAtPriorityIndexFunc) nextHook() func(int) []*process.ProcessMeta {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ProcessContainerGetProcessesAtPriorityIndexFunc) appendCall(r0 ProcessContainerGetProcessesAtPriorityIndexFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of
// ProcessContainerGetProcessesAtPriorityIndexFuncCall objects describing
// the invocations of this function.
func (f *ProcessContainerGetProcessesAtPriorityIndexFunc) History() []ProcessContainerGetProcessesAtPriorityIndexFuncCall {
	f.mutex.Lock()
	history := make([]ProcessContainerGetProcessesAtPriorityIndexFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ProcessContainerGetProcessesAtPriorityIndexFuncCall is an object that
// describes an invocation of method GetProcessesAtPriorityIndex on an
// instance of MockProcessContainer.
type ProcessContainerGetProcessesAtPriorityIndexFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 int
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 []*process.ProcessMeta
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c ProcessContainerGetProcessesAtPriorityIndexFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c ProcessContainerGetProcessesAtPriorityIndexFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// ProcessContainerNumInitializersFunc describes the behavior when the
// NumInitializers method of the parent MockProcessContainer instance is
// invoked.
type ProcessContainerNumInitializersFunc struct {
	defaultHook func() int
	hooks       []func() int
	history     []ProcessContainerNumInitializersFuncCall
	mutex       sync.Mutex
}

// NumInitializers delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockProcessContainer) NumInitializers() int {
	r0 := m.NumInitializersFunc.nextHook()()
	m.NumInitializersFunc.appendCall(ProcessContainerNumInitializersFuncCall{r0})
	return r0
}

// SetDefaultHook sets function that is called when the NumInitializers
// method of the parent MockProcessContainer instance is invoked and the
// hook queue is empty.
func (f *ProcessContainerNumInitializersFunc) SetDefaultHook(hook func() int) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// NumInitializers method of the parent MockProcessContainer instance
// invokes the hook at the front of the queue and discards it. After the
// queue is empty, the default hook function is invoked for any future
// action.
func (f *ProcessContainerNumInitializersFunc) PushHook(hook func() int) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *ProcessContainerNumInitializersFunc) SetDefaultReturn(r0 int) {
	f.SetDefaultHook(func() int {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *ProcessContainerNumInitializersFunc) PushReturn(r0 int) {
	f.PushHook(func() int {
		return r0
	})
}

func (f *ProcessContainerNumInitializersFunc) nextHook() func() int {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ProcessContainerNumInitializersFunc) appendCall(r0 ProcessContainerNumInitializersFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of ProcessContainerNumInitializersFuncCall
// objects describing the invocations of this function.
func (f *ProcessContainerNumInitializersFunc) History() []ProcessContainerNumInitializersFuncCall {
	f.mutex.Lock()
	history := make([]ProcessContainerNumInitializersFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ProcessContainerNumInitializersFuncCall is an object that describes an
// invocation of method NumInitializers on an instance of
// MockProcessContainer.
type ProcessContainerNumInitializersFuncCall struct {
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 int
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c ProcessContainerNumInitializersFuncCall) Args() []interface{} {
	return []interface{}{}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c ProcessContainerNumInitializersFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// ProcessContainerNumPrioritiesFunc describes the behavior when the
// NumPriorities method of the parent MockProcessContainer instance is
// invoked.
type ProcessContainerNumPrioritiesFunc struct {
	defaultHook func() int
	hooks       []func() int
	history     []ProcessContainerNumPrioritiesFuncCall
	mutex       sync.Mutex
}

// NumPriorities delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockProcessContainer) NumPriorities() int {
	r0 := m.NumPrioritiesFunc.nextHook()()
	m.NumPrioritiesFunc.appendCall(ProcessContainerNumPrioritiesFuncCall{r0})
	return r0
}

// SetDefaultHook sets function that is called when the NumPriorities method
// of the parent MockProcessContainer instance is invoked and the hook queue
// is empty.
func (f *ProcessContainerNumPrioritiesFunc) SetDefaultHook(hook func() int) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// NumPriorities method of the parent MockProcessContainer instance invokes
// the hook at the front of the queue and discards it. After the queue is
// empty, the default hook function is invoked for any future action.
func (f *ProcessContainerNumPrioritiesFunc) PushHook(hook func() int) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *ProcessContainerNumPrioritiesFunc) SetDefaultReturn(r0 int) {
	f.SetDefaultHook(func() int {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *ProcessContainerNumPrioritiesFunc) PushReturn(r0 int) {
	f.PushHook(func() int {
		return r0
	})
}

func (f *ProcessContainerNumPrioritiesFunc) nextHook() func() int {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ProcessContainerNumPrioritiesFunc) appendCall(r0 ProcessContainerNumPrioritiesFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of ProcessContainerNumPrioritiesFuncCall
// objects describing the invocations of this function.
func (f *ProcessContainerNumPrioritiesFunc) History() []ProcessContainerNumPrioritiesFuncCall {
	f.mutex.Lock()
	history := make([]ProcessContainerNumPrioritiesFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ProcessContainerNumPrioritiesFuncCall is an object that describes an
// invocation of method NumPriorities on an instance of
// MockProcessContainer.
type ProcessContainerNumPrioritiesFuncCall struct {
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 int
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c ProcessContainerNumPrioritiesFuncCall) Args() []interface{} {
	return []interface{}{}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c ProcessContainerNumPrioritiesFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// ProcessContainerNumProcessesFunc describes the behavior when the
// NumProcesses method of the parent MockProcessContainer instance is
// invoked.
type ProcessContainerNumProcessesFunc struct {
	defaultHook func() int
	hooks       []func() int
	history     []ProcessContainerNumProcessesFuncCall
	mutex       sync.Mutex
}

// NumProcesses delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockProcessContainer) NumProcesses() int {
	r0 := m.NumProcessesFunc.nextHook()()
	m.NumProcessesFunc.appendCall(ProcessContainerNumProcessesFuncCall{r0})
	return r0
}

// SetDefaultHook sets function that is called when the NumProcesses method
// of the parent MockProcessContainer instance is invoked and the hook queue
// is empty.
func (f *ProcessContainerNumProcessesFunc) SetDefaultHook(hook func() int) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// NumProcesses method of the parent MockProcessContainer instance invokes
// the hook at the front of the queue and discards it. After the queue is
// empty, the default hook function is invoked for any future action.
func (f *ProcessContainerNumProcessesFunc) PushHook(hook func() int) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *ProcessContainerNumProcessesFunc) SetDefaultReturn(r0 int) {
	f.SetDefaultHook(func() int {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *ProcessContainerNumProcessesFunc) PushReturn(r0 int) {
	f.PushHook(func() int {
		return r0
	})
}

func (f *ProcessContainerNumProcessesFunc) nextHook() func() int {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ProcessContainerNumProcessesFunc) appendCall(r0 ProcessContainerNumProcessesFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of ProcessContainerNumProcessesFuncCall
// objects describing the invocations of this function.
func (f *ProcessContainerNumProcessesFunc) History() []ProcessContainerNumProcessesFuncCall {
	f.mutex.Lock()
	history := make([]ProcessContainerNumProcessesFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ProcessContainerNumProcessesFuncCall is an object that describes an
// invocation of method NumProcesses on an instance of MockProcessContainer.
type ProcessContainerNumProcessesFuncCall struct {
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 int
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c ProcessContainerNumProcessesFuncCall) Args() []interface{} {
	return []interface{}{}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c ProcessContainerNumProcessesFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// ProcessContainerRegisterInitializerFunc describes the behavior when the
// RegisterInitializer method of the parent MockProcessContainer instance is
// invoked.
type ProcessContainerRegisterInitializerFunc struct {
	defaultHook func(process.Initializer, ...process.InitializerConfigFunc)
	hooks       []func(process.Initializer, ...process.InitializerConfigFunc)
	history     []ProcessContainerRegisterInitializerFuncCall
	mutex       sync.Mutex
}

// RegisterInitializer delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockProcessContainer) RegisterInitializer(v0 process.Initializer, v1 ...process.InitializerConfigFunc) {
	m.RegisterInitializerFunc.nextHook()(v0, v1...)
	m.RegisterInitializerFunc.appendCall(ProcessContainerRegisterInitializerFuncCall{v0, v1})
	return
}

// SetDefaultHook sets function that is called when the RegisterInitializer
// method of the parent MockProcessContainer instance is invoked and the
// hook queue is empty.
func (f *ProcessContainerRegisterInitializerFunc) SetDefaultHook(hook func(process.Initializer, ...process.InitializerConfigFunc)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// RegisterInitializer method of the parent MockProcessContainer instance
// invokes the hook at the front of the queue and discards it. After the
// queue is empty, the default hook function is invoked for any future
// action.
func (f *ProcessContainerRegisterInitializerFunc) PushHook(hook func(process.Initializer, ...process.InitializerConfigFunc)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *ProcessContainerRegisterInitializerFunc) SetDefaultReturn() {
	f.SetDefaultHook(func(process.Initializer, ...process.InitializerConfigFunc) {
		return
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *ProcessContainerRegisterInitializerFunc) PushReturn() {
	f.PushHook(func(process.Initializer, ...process.InitializerConfigFunc) {
		return
	})
}

func (f *ProcessContainerRegisterInitializerFunc) nextHook() func(process.Initializer, ...process.InitializerConfigFunc) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ProcessContainerRegisterInitializerFunc) appendCall(r0 ProcessContainerRegisterInitializerFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of ProcessContainerRegisterInitializerFuncCall
// objects describing the invocations of this function.
func (f *ProcessContainerRegisterInitializerFunc) History() []ProcessContainerRegisterInitializerFuncCall {
	f.mutex.Lock()
	history := make([]ProcessContainerRegisterInitializerFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ProcessContainerRegisterInitializerFuncCall is an object that describes
// an invocation of method RegisterInitializer on an instance of
// MockProcessContainer.
type ProcessContainerRegisterInitializerFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 process.Initializer
	// Arg1 is a slice containing the values of the variadic arguments
	// passed to this method invocation.
	Arg1 []process.InitializerConfigFunc
}

// Args returns an interface slice containing the arguments of this
// invocation. The variadic slice argument is flattened in this array such
// that one positional argument and three variadic arguments would result in
// a slice of four, not two.
func (c ProcessContainerRegisterInitializerFuncCall) Args() []interface{} {
	trailing := []interface{}{}
	for _, val := range c.Arg1 {
		trailing = append(trailing, val)
	}

	return append([]interface{}{c.Arg0}, trailing...)
}

// Results returns an interface slice containing the results of this
// invocation.
func (c ProcessContainerRegisterInitializerFuncCall) Results() []interface{} {
	return []interface{}{}
}

// ProcessContainerRegisterProcessFunc describes the behavior when the
// RegisterProcess method of the parent MockProcessContainer instance is
// invoked.
type ProcessContainerRegisterProcessFunc struct {
	defaultHook func(process.Process, ...process.ProcessConfigFunc)
	hooks       []func(process.Process, ...process.ProcessConfigFunc)
	history     []ProcessContainerRegisterProcessFuncCall
	mutex       sync.Mutex
}

// RegisterProcess delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockProcessContainer) RegisterProcess(v0 process.Process, v1 ...process.ProcessConfigFunc) {
	m.RegisterProcessFunc.nextHook()(v0, v1...)
	m.RegisterProcessFunc.appendCall(ProcessContainerRegisterProcessFuncCall{v0, v1})
	return
}

// SetDefaultHook sets function that is called when the RegisterProcess
// method of the parent MockProcessContainer instance is invoked and the
// hook queue is empty.
func (f *ProcessContainerRegisterProcessFunc) SetDefaultHook(hook func(process.Process, ...process.ProcessConfigFunc)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// RegisterProcess method of the parent MockProcessContainer instance
// invokes the hook at the front of the queue and discards it. After the
// queue is empty, the default hook function is invoked for any future
// action.
func (f *ProcessContainerRegisterProcessFunc) PushHook(hook func(process.Process, ...process.ProcessConfigFunc)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *ProcessContainerRegisterProcessFunc) SetDefaultReturn() {
	f.SetDefaultHook(func(process.Process, ...process.ProcessConfigFunc) {
		return
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *ProcessContainerRegisterProcessFunc) PushReturn() {
	f.PushHook(func(process.Process, ...process.ProcessConfigFunc) {
		return
	})
}

func (f *ProcessContainerRegisterProcessFunc) nextHook() func(process.Process, ...process.ProcessConfigFunc) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ProcessContainerRegisterProcessFunc) appendCall(r0 ProcessContainerRegisterProcessFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of ProcessContainerRegisterProcessFuncCall
// objects describing the invocations of this function.
func (f *ProcessContainerRegisterProcessFunc) History() []ProcessContainerRegisterProcessFuncCall {
	f.mutex.Lock()
	history := make([]ProcessContainerRegisterProcessFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ProcessContainerRegisterProcessFuncCall is an object that describes an
// invocation of method RegisterProcess on an instance of
// MockProcessContainer.
type ProcessContainerRegisterProcessFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 process.Process
	// Arg1 is a slice containing the values of the variadic arguments
	// passed to this method invocation.
	Arg1 []process.ProcessConfigFunc
}

// Args returns an interface slice containing the arguments of this
// invocation. The variadic slice argument is flattened in this array such
// that one positional argument and three variadic arguments would result in
// a slice of four, not two.
func (c ProcessContainerRegisterProcessFuncCall) Args() []interface{} {
	trailing := []interface{}{}
	for _, val := range c.Arg1 {
		trailing = append(trailing, val)
	}

	return append([]interface{}{c.Arg0}, trailing...)
}

// Results returns an interface slice containing the results of this
// invocation.
func (c ProcessContainerRegisterProcessFuncCall) Results() []interface{} {
	return []interface{}{}
}
