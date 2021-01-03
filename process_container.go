package process

import "sort"

// ProcessContainer is a collection of initializers and processes.
type ProcessContainer interface {
	// RegisterInitializer adds an initializer to the container
	// with the given configuration.
	RegisterInitializer(Initializer, ...InitializerConfigFunc)

	// NumInitializers returns the number of registered initializers.
	NumInitializers() int

	// NumInitializer returns the number of distinct registered
	// initializer priorities.
	NumInitializerPriorities() int

	// GetProcessesAtPriorityIndex returns a slice of meta objects
	// wrapping all initializers registered to this priority index,
	// where zero denotes the lowest priority, one the second
	// lowest, and so on. The index parameter is not checked for
	// validity before indexing an internal slice - caller beware.
	GetInitializersAtPriorityIndex(index int) []*InitializerMeta

	// RegisterProcess adds a process to the container with the
	// given configuration.
	RegisterProcess(Process, ...ProcessConfigFunc)

	// NumProcesses returns the number of registered processes.
	NumProcesses() int

	// NumProcessPriorities returns the number of distinct registered
	// process priorities.
	NumProcessPriorities() int

	// GetProcessesAtPriorityIndex returns a slice of meta objects
	// wrapping all processes registered to this priority index,
	// where zero denotes the lowest priority, one the second
	// lowest, and so on. The index parameter is not checked for
	// validity before indexing an internal slice - caller beware.
	GetProcessesAtPriorityIndex(index int) []*ProcessMeta
}

type container struct {
	initializers          map[int][]*InitializerMeta
	processes             map[int][]*ProcessMeta
	initializerPriorities []int
	processPriorities     []int
}

var _ ProcessContainer = &container{}

// NewProcessContainer creates an empty process container.
func NewProcessContainer() ProcessContainer {
	return &container{
		initializers:          map[int][]*InitializerMeta{},
		processes:             map[int][]*ProcessMeta{},
		initializerPriorities: []int{},
		processPriorities:     []int{},
	}
}

func (c *container) RegisterInitializer(
	initializer Initializer,
	initializerConfigs ...InitializerConfigFunc,
) {
	meta := newInitializerMeta(initializer)

	for _, f := range initializerConfigs {
		f(meta)
	}

	c.initializers[meta.priority] = append(c.initializers[meta.priority], meta)
	c.initializerPriorities = c.getInitializerPriorities()
}

func (c *container) NumInitializers() int {
	n := 0
	for _, is := range c.initializers {
		n += len(is)
	}

	return n
}

func (c *container) NumInitializerPriorities() int {
	return len(c.initializerPriorities)
}

func (c *container) GetInitializersAtPriorityIndex(index int) []*InitializerMeta {
	return c.initializers[c.initializerPriorities[index]]
}

func (c *container) getInitializerPriorities() []int {
	priorities := []int{}
	for priority := range c.initializers {
		priorities = append(priorities, priority)
	}

	sort.Ints(priorities)
	return priorities
}

func (c *container) RegisterProcess(
	process Process,
	processConfigs ...ProcessConfigFunc,
) {
	meta := newProcessMeta(process)

	for _, f := range processConfigs {
		f(meta)
	}

	c.processes[meta.priority] = append(c.processes[meta.priority], meta)
	c.processPriorities = c.getProcessPriorities()
}

func (c *container) NumProcesses() int {
	n := 0
	for _, ps := range c.processes {
		n += len(ps)
	}

	return n
}

func (c *container) NumProcessPriorities() int {
	return len(c.processPriorities)
}

func (c *container) GetProcessesAtPriorityIndex(index int) []*ProcessMeta {
	return c.processes[c.processPriorities[index]]
}

func (c *container) getProcessPriorities() []int {
	priorities := []int{}
	for priority := range c.processes {
		priorities = append(priorities, priority)
	}

	sort.Ints(priorities)
	return priorities
}
