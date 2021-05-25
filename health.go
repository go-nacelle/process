package process

import (
	"fmt"
	"sync"
)

// Health is an aggregate container reporting the current health status of
// individual application components.
type Health struct {
	mu          sync.Mutex
	components  map[interface{}]*HealthComponentStatus
	subscribers []chan<- struct{}
}

// NewHealth creates an empty Health instance.
func NewHealth() *Health {
	return &Health{
		components: map[interface{}]*HealthComponentStatus{},
	}
}

// Healthy returns true if all registered components are healthy.
func (h *Health) Healthy() bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, component := range h.components {
		if !component.healthy {
			return false
		}
	}

	return true
}

// Subscribe returns a notification channel that receives a value whenever
// the set of components change, or the status of a component changes. This
// method also returns a cancellation function that should be called once
// the user wishes to unsubscribe.
func (h *Health) Subscribe() (<-chan struct{}, func()) {
	h.mu.Lock()
	defer h.mu.Unlock()

	n := 0
	for n < len(h.subscribers) && h.subscribers[n] != nil {
		n++
	}

	ch := make(chan struct{}, 1)
	ch <- struct{}{}
	h.subscribers = append(h.subscribers, ch)

	unsubscribe := func() {
		h.mu.Lock()
		h.subscribers[n] = nil
		h.mu.Unlock()
	}

	var once sync.Once
	return ch, func() { once.Do(unsubscribe) }
}

// Get returns the component status value registered to the given key.
func (h *Health) Get(key interface{}) (*HealthComponentStatus, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	component, ok := h.components[key]
	return component, ok
}

// GetAll returns the component status values registered to the given keys.
func (h *Health) GetAll(keys ...interface{}) ([]*HealthComponentStatus, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	components := make([]*HealthComponentStatus, 0, len(keys))
	for _, key := range keys {
		component, ok := h.components[key]
		if !ok {
			return nil, fmt.Errorf("health component %q not registered", key)
		}

		components = append(components, component)
	}

	return components, nil
}

// Register creates and returns a new component status value for the given key.
// It an error to register the same key twice.
func (h *Health) Register(key interface{}) (*HealthComponentStatus, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.components[key]; ok {
		return nil, ErrHealthComponentAlreadyRegistered
	}

	component := newHealthComponentStatus(h, key)
	h.components[key] = component
	h.notify()
	return component, nil
}

// notify writes a signal to all subscribed channels. Callers MUST lock b.mu.
func (h *Health) notify() {
	for _, subscriber := range h.subscribers {
		if subscriber == nil {
			continue
		}

		select {
		case subscriber <- struct{}{}:
		default:
		}
	}
}
