package process

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockTestingT struct{}

func (mockTestingT) Errorf(format string, args ...interface{}) {}

func eventually(t *testing.T, cond func() bool) bool {
	return assert.Eventually(t, cond, time.Second, 10*time.Millisecond)
}

func consistently(t *testing.T, cond func() bool) bool {
	if !assert.Eventually(mockTestingT{}, func() bool { return !cond() }, 100*time.Millisecond, 10*time.Millisecond) {
		return true
	}

	return assert.Fail(t, "Condition not met during test period")
}

func consistentlyNot(t *testing.T, cond func() bool) bool {
	return consistently(t, func() bool { return !cond() })
}

func errorChanClosed(ch <-chan error) func() bool {
	return func() bool {
		select {
		case _, ok := <-ch:
			return !ok
		default:
			return false
		}
	}
}

func errorChanReceivesUnordered(ch <-chan error, expectedValues ...string) func() bool {
	containsError := func(errs []error, substring string) bool {
		for _, err := range errs {
			if err != nil && strings.Contains(err.Error(), substring) {
				return true
			}
		}

		return false
	}

	return errChanReceives(ch, len(expectedValues), func(errs []error) bool {
		for _, expectedValue := range expectedValues {
			if !containsError(errs, expectedValue) {
				return false
			}
		}

		return true
	})
}

func errChanReceives(ch <-chan error, n int, cond func(errs []error) bool) func() bool {
	var errs []error

	return func() bool {
		select {
		case value, ok := <-ch:
			if !ok {
				return false
			}

			errs = append(errs, value)

			if len(errs) == n {
				return cond(errs)
			}
		default:
		}

		return false
	}
}

func errorChanDoesNotReceive(ch <-chan error) func() bool {
	return func() bool {
		select {
		case <-ch:
			return false
		default:
			return true
		}
	}
}

func structChanClosed(ch <-chan struct{}) func() bool {
	return func() bool {
		select {
		case _, ok := <-ch:
			return !ok
		default:
			return false
		}
	}
}

func structChanDoesNotReceive(ch <-chan struct{}) func() bool {
	return func() bool {
		select {
		case <-ch:
			return false
		default:
			return true
		}
	}
}

func stringChanReceivesOrdered(ch <-chan string, expectedValues ...string) func() bool {
	return stringChanReceives(ch, len(expectedValues), func(values []string) bool {
		for i, expectedValue := range expectedValues {
			if values[i] != expectedValue {
				return false
			}
		}

		return true
	})
}

func stringChanReceivesUnordered(ch <-chan string, expectedValues ...string) func() bool {
	contains := func(values []string, value string) bool {
		for _, v := range values {
			if v == value {
				return true
			}
		}

		return false
	}

	return stringChanReceives(ch, len(expectedValues), func(values []string) bool {
		for _, expectedValue := range expectedValues {
			if !contains(values, expectedValue) {
				return false
			}
		}

		return true
	})
}

func stringChanReceivesN(ch <-chan string, n int) func() bool {
	return stringChanReceives(ch, n, func(values []string) bool { return true })
}

func stringChanReceives(ch <-chan string, n int, cond func(values []string) bool) func() bool {
	var values []string

	return func() bool {
		select {
		case value, ok := <-ch:
			if !ok {
				return false
			}

			values = append(values, value)

			if len(values) == n {
				return cond(values)
			}
		default:
		}

		return false
	}
}

func stringChanDoesNotReceive(ch <-chan string) func() bool {
	return func() bool {
		select {
		case <-ch:
			return false
		default:
			return true
		}
	}
}
