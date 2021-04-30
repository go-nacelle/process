package process

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var testErr1 = errors.New("oops1")
var testErr2 = errors.New("oops2")

// type testLogger struct{}

// func (l testLogger) WithFields(LogFields) Logger             { return l }
// func (l testLogger) Info(msg string, args ...interface{})    { fmt.Printf("INFO: "+msg+"\n", args...) }
// func (l testLogger) Warning(msg string, args ...interface{}) { fmt.Printf("WARN: "+msg+"\n", args...) }
// func (l testLogger) Error(msg string, args ...interface{})   { fmt.Printf("ERROR: "+msg+"\n", args...) }

// forwardN returns a channel that proxies the next n values read from the
// given channel. The returned channel will be closed after the nth value
// has been written.
func forwardN(ch <-chan string, n int) <-chan string {
	ch2 := make(chan string)
	go func() {
		defer close(ch2)
		for i := 0; i < n; i++ {
			ch2 <- <-ch
		}
	}()

	return ch2
}

// traceInit returns a function that can be used as an process's Init method. The returned function
// will register an initially unhealthy health component if the given health value is non-nil. The
// returned function will write a unique string built from the given name and index to the given
// channel once it has been invoked.
func traceInit(health *Health, trace chan<- string, name string, index int, err error) singleErrorFunc {
	return func(ctx context.Context) error {
		if health != nil {
			healthComponent, err := health.Register(testHealthKey(name, index))
			if err != nil {
				return err
			}

			healthComponent.Update(false)
		}

		trace <- fmt.Sprintf("%s.%d.init", name, index)
		return err
	}
}

// traceRun returns a function that can be used as a process's Run method. The returned
// function will update the health component registered in traceInit if the given health value
// is non-nil. The returned function will write a unique string built from the given name and
// index to the given channel once it has been invoked.
func traceRun(health *Health, trace chan<- string, name string, index int, err error) singleErrorFunc {
	return func(ctx context.Context) error {
		trace <- fmt.Sprintf("%s.%d.run", name, index)
		if err != nil {
			return err
		}

		if health != nil {
			healthComponent, ok := health.Get(testHealthKey(name, index))
			if !ok {
				return fmt.Errorf("unknown health key")
			}

			healthComponent.Update(true)
		}

		<-ctx.Done()
		return ctx.Err()
	}
}

// testHealthKey creates a deterministic health component key from the given name and index.
func testHealthKey(name string, index int) interface{} {
	return fmt.Sprintf("%s.%d", name, index)
}

// traceStop returns a function that can be used as a process's Stop method. The returned function
// will write a unique string built from the given name and index to the given channel once it has
// been invoked.
func traceStop(trace chan<- string, name string, index int, err error) singleErrorFunc {
	return func(ctx context.Context) error {
		trace <- fmt.Sprintf("%s.%d.stop", name, index)
		return err
	}
}

// traceFinalize returns a function that can be used as a process's Finalize method. The returned
// function  will write a unique string built from the given name and index to the given channel
// once it has been invoked.
func traceFinalize(trace chan<- string, name string, index int, err error) singleErrorFunc {
	return func(ctx context.Context) error {
		trace <- fmt.Sprintf("%s.%d.finalize", name, index)
		return err
	}
}

// newBlockingSingleErrorFunc returns a function that blocks indefinitely. This
// method also returns a channel closes once the function is invoked. The returned
// function is not safe to call multiple times.
func newBlockingSingleErrorFunc() (singleErrorFunc, <-chan struct{}) {
	scheduled := make(chan struct{})
	fn := func(ctx context.Context) error {
		close(scheduled)
		select {}
	}

	return fn, scheduled
}

// newSingalingSingleErrorFunc returns a function that blocks until the given context
// is canceled. This function also returns a channel closes once the function is invoked.
// The returned function returns the context error value. The returned function is not
// safe to call multiple times.
func newSingalingSingleErrorFunc() (singleErrorFunc, <-chan struct{}) {
	scheduled := make(chan struct{})
	fn := func(ctx context.Context) error {
		close(scheduled)
		<-ctx.Done()
		return ctx.Err()
	}

	return fn, scheduled
}

// newTracedSingleErrorFunc returns a function that sends the given list of values down
// the given channel when invoked.
func newTracedSingleErrorFunc(trace chan<- string, values <-chan string, err error) singleErrorFunc {
	return func(ctx context.Context) error {
		for v := range values {
			trace <- v
		}

		return err
	}
}

// traceValue writes the given value to a new channel, closes it, then returns it.
func traceValue(value string) <-chan string {
	ch := make(chan string, 1)
	ch <- value
	close(ch)

	return ch
}

type unorderedWrapper struct {
	values []interface{}
}

// seq returns the given values as a slice.
func seq(values ...interface{}) []interface{} {
	return values
}

// unordered wraps the given values in an instance of unorderedWrapper. This
// type is used to signify in assertChannelContents that all of the values in
// the set must occur before any other values, but they may occur in any order.
//
// For example, valid orderings of [a, b, unordered(c, d, e), f, g] are:
//
// - [a, b, c, d, e, f, g]
// - [a, b, c, e, d, f, g]
// - [a, b, d, c, e, f, g]
// - [a, b, d, e, c, f, g]
// - [a, b, e, c, d, f, g]
// - [a, b, e, d, c, f, g]
func unordered(values ...interface{}) unorderedWrapper {
	return unorderedWrapper{values}
}

// assertChannelContents invokes the given read function, which returns a slice
// of values from the channel and a boolean flag indicating whether or not a read
// timed out.
//
// The values in the slice are compared against the given expected values. This
// method will report a fatal test error if the content of the slices differ in
// value or (generally) order. Runs of values can be marked as occurring in any
// order by wrapping them via the unordered function.
func assertChannelContents(t *testing.T, read readFunc, expectedValues []interface{}, msgAndArgs ...interface{}) {
	values, timedOut := read(t)
	if timedOut {
		msgAndArgs = wrapTestError("Channel did not close", msgAndArgs...)
	}

	if len(expectedValues) == 0 {
		// Normalize for empty case
		require.Empty(t, values, msgAndArgs...)
		return
	}

	i := 0 // index for values
	j := 0 // index for expectedValues
	var next []interface{}

	for {
		for len(next) > 0 {
			if i >= len(values) {
				reportMismatchedContent(t, "Expected more values.", expectedValues, values)
			}

			found := false
			for k, candidate := range next {
				if objectsAreEqualValues(values[i], candidate) {
					found = true
					next[0], next[k] = next[k], next[0]
					break
				}
			}
			if !found {
				reportMismatchedContent(t, "Mismatched values.", expectedValues, values)
			}

			i++
			next = next[1:]
		}

		if j == len(expectedValues) {
			break
		}

		if u, ok := expectedValues[j].(unorderedWrapper); ok {
			next = u.values
		} else {
			next = []interface{}{expectedValues[j]}
		}

		j++
	}

	if i < len(values) {
		reportMismatchedContent(t, "Expected fewer values.", expectedValues, values)
	}
}

// reportMismatchedContent invokes a fatal test error with a serialized version of the
// given expected values and actual values.
func reportMismatchedContent(t *testing.T, message string, expectedValues, values []interface{}) {
	var flattened []string
	for _, expectedValue := range expectedValues {
		if u, ok := expectedValue.(unorderedWrapper); ok {
			var strValues []string
			for _, value := range u.values {
				strValues = append(strValues, fmt.Sprintf("%s", value))
			}

			flattened = append(flattened, fmt.Sprintf("{ %s }", strings.Join(strValues, ", ")))
		} else {
			flattened = append(flattened, fmt.Sprintf("%v", expectedValue))
		}
	}

	t.Fatalf(fmt.Sprintf("%s want=%s, have=%v", message, strings.Join(flattened, ", "), values))
}

// objectsAreEqualValues returns true if the given values are logically equal. We do
// this currently by casting them both to strings and comparing the resulting serialized
// value. This works because we're currently only testing with string and error values,
// which have well-formed and deterministic string representations.
func objectsAreEqualValues(expected, value interface{}) bool {
	return fmt.Sprintf("%s", expected) == fmt.Sprintf("%s", value)
}

// permute returns a slice of permutations of the given values.
func permute(values []interface{}) [][]interface{} {
	ch := make(chan []interface{})
	go func() {
		defer close(ch)
		permuteHelper(values, 0, ch)
	}()

	var permutations [][]interface{}
	for permutation := range ch {
		permutations = append(permutations, permutation)
	}

	return permutations
}

// permuteHelper recursively computes permutations of the given values. When a
// full permutation is constructed, it is written to the given channel.
//
// The given values paramete stores a mutable version of the input undergoing
// permutation. The index parameter specifies the size of the slice prefix that
// should not be modified on this chain of recursive calls.
func permuteHelper(values []interface{}, index int, ch chan<- []interface{}) {
	if index > len(values) {
		// values is mutable, make a copy before sending it
		valuesCopy := make([]interface{}, len(values))
		copy(valuesCopy, values)
		ch <- valuesCopy
		return
	}

	permuteHelper(values, index+1, ch)

	for j := index + 1; j < len(values); j++ {
		values[index], values[j] = values[j], values[index]
		permuteHelper(values, index+1, ch)
		values[index], values[j] = values[j], values[index]
	}
}

type readFunc func(t *testing.T) (values []interface{}, complete bool)

// wrapTestError appends the given prefix to the error message template given
// in the values composing msgAndArgs.
func wrapTestError(prefix string, msgAndArgs ...interface{}) []interface{} {
	if len(msgAndArgs) == 0 {
		return []interface{}{prefix}
	}

	if msg, ok := msgAndArgs[0].(string); ok {
		prefixedMessage := fmt.Sprintf("%s (%s)", prefix, msg)
		return append([]interface{}{prefixedMessage}, msgAndArgs[1:]...)
	}

	return msgAndArgs
}

// readErrorChannel returns a function that reads from the given channel with
// a timeout. If no value was ready to be read from the channel for the timeout
// period, a false-valued flag is returned.
func readErrorChannel(ch <-chan error) readFunc {
	return func(t *testing.T) (values []interface{}, complete bool) {
		for {
			select {
			case value, ok := <-ch:
				if !ok {
					return values, true
				}

				values = append(values, value)

			case <-time.After(time.Second):
				return values, false
			}
		}
	}
}

// readStringChannel returns a function that reads from the given channel with
// a timeout. If no value was ready to be read from the channel for the timeout
// period, a false-valued flag is returned.
func readStringChannel(ch <-chan string) readFunc {
	return func(t *testing.T) (values []interface{}, complete bool) {
		for {
			select {
			case value, ok := <-ch:
				if !ok {
					return values, true
				}

				values = append(values, value)

			case <-time.After(time.Second):
				return values, false
			}
		}
	}
}
