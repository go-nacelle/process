package process

// Container is an immutable container used to hold registered processes.
// A container instance is constructed from a mutable container builder.
type Container struct {
	meta       map[int][]*Meta
	priorities []int
}

// Meta returns a new slice of meta values registered to the container.
func (c *Container) Meta() []*Meta {
	var all []*Meta
	for _, meta := range c.meta {
		all = append(all, meta...)
	}

	return all
}

// MetaForPriority returns a new slice of meta values registered to the given priority.
func (c *Container) MetaForPriority(priority int) []*Meta {
	meta := make([]*Meta, len(c.meta[priority]))
	copy(meta, c.meta[priority])
	return meta
}

// Priorities returns a new slice of unique priorities to which a meta value is registered.
func (c *Container) Priorities() []int {
	priorities := make([]int, len(c.priorities))
	copy(priorities, c.priorities)
	return priorities
}
