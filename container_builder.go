package process

import "sort"

// ContainerBuilder is a mutable container used to register processes at
// application boot. The container builder can be frozen into an immutable
// container.
type ContainerBuilder struct {
	registrations []registration
}

// registration is a raw process value paired with the set of configuration
// values supppiled when the value ws first registered to a container builder.
type registration struct {
	wrapped interface{}
	configs []MetaConfigFunc
}

// NewContainerBuilder creates an empty container builder.
func NewContainerBuilder() *ContainerBuilder {
	return &ContainerBuilder{}
}

// RegisterInitializer registers an initializer with the given configs.
func (b *ContainerBuilder) RegisterInitializer(wrapped interface{}, configs ...MetaConfigFunc) {
	b.RegisterProcess(wrapped, append(configs, WithEarlyExit(true))...)
}

// RegisterProcess registers a process with the given configs.
func (b *ContainerBuilder) RegisterProcess(wrapped interface{}, configs ...MetaConfigFunc) {
	b.registrations = append(b.registrations, registration{wrapped: wrapped, configs: configs})
}

// Build creates a frozen and immutable version of the container containing
// all of the processes registered to the container builder thus far.
func (b *ContainerBuilder) Build(configs ...MetaConfigFunc) *Container {
	processes := map[int][]*Meta{}
	for _, registration := range b.registrations {
		configs := unionConfigs(configs, registration.configs)
		meta := newMeta(registration.wrapped, configs...)
		priority := meta.options.priority
		processes[priority] = append(processes[priority], meta)
	}

	return &Container{
		meta:       processes,
		priorities: sortPriorities(processes),
	}
}

// unionConfigs returns a new slice consisting of all the elements of left
// followed by all the elements or right, both in their original order.
func unionConfigs(left, right []MetaConfigFunc) []MetaConfigFunc {
	union := make([]MetaConfigFunc, len(left)+len(right))
	copy(union, left)
	copy(union[len(left):], right)

	return union
}

// sortPriorities returns a sorted slice of the keys forming the given map.
func sortPriorities(m map[int][]*Meta) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Ints(keys)
	return keys
}
