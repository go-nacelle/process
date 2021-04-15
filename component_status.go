package process

import "time"

// HealthComponentStatus manages the current status of an application component.
type HealthComponentStatus struct {
	health      *Health
	key         interface{}
	healthy     bool
	lastUpdated time.Time
}

func newHealthComponentStatus(health *Health, key interface{}) *HealthComponentStatus {
	return &HealthComponentStatus{
		health: health,
		key:    key,
	}
}

// Healthy returns true if the component is currently healthy.
func (s *HealthComponentStatus) Healthy() bool {
	s.health.mu.Lock()
	defer s.health.mu.Unlock()

	return s.healthy
}

// Update sets the current health status of the application component.
func (s *HealthComponentStatus) Update(healthy bool) {
	s.health.mu.Lock()
	defer s.health.mu.Unlock()

	if s.healthy == healthy {
		return
	}

	s.healthy = healthy
	s.lastUpdated = time.Now()
	s.health.notify()
}
