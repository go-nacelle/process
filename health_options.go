package process

import "github.com/derision-test/glock"

type HealthConfigFunc func(*health)

func WithHealthClock(clock glock.Clock) HealthConfigFunc {
	return func(h *health) { h.clock = clock }
}
