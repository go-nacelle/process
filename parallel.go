package process

import (
	"context"
	"sync"
)

// singleErrorFunc is a function that returns a single error value.
type singleErrorFunc func(ctx context.Context) error

// streamErrorFunc is a function that returns a stream of error values.
type streamErrorFunc func(ctx context.Context) <-chan error

// toStreamErrorFunc converts a singleErrorFunc into a streamErrorFunc.
func toStreamErrorFunc(fn singleErrorFunc) streamErrorFunc {
	return func(ctx context.Context) <-chan error {
		return withErrors(func(errs chan<- error) {
			if err := fn(ctx); err != nil {
				errs <- err
			}
		})
	}
}

// chain returns a streamErrorFunc that invokes each of the given functions
// sequentially. Errors emitted from each source function will be emitted
// unchanged by the returned function. If a function emits an error, the
// remaining functions will not be called.
func chain(fns ...streamErrorFunc) streamErrorFunc {
	return combineInternal(true, fns...)
}

// sequence returns a streamErrorFunc that invokes each of the given functions
// sequentially. Errors emitted from each source function will be emitted
// unchanged by the returned function. Every function will be called whether or
// not a function earlier in the sequence emitted an error.
func sequence(fns ...streamErrorFunc) streamErrorFunc {
	return combineInternal(false, fns...)
}

func combineInternal(stopOnError bool, fns ...streamErrorFunc) streamErrorFunc {
	if len(fns) == 1 {
		return fns[0]
	}

	return func(ctx context.Context) <-chan error {
		return withErrors(func(errs chan<- error) {
			hasError := false

			for _, fn := range fns {
				if hasError && stopOnError {
					break
				}

				for err := range fn(ctx) {
					hasError = true
					errs <- err
				}
			}
		})
	}
}

// parallel returns a streamErrorFunc that invokes the given functions
// concurrently and blocks until each function returns. An error emitted
// from any source function will be immediately emitted from the returned
// function.
func parallel(funcs ...streamErrorFunc) streamErrorFunc {
	return func(ctx context.Context) <-chan error {
		return withErrors(func(errs chan<- error) {
			var wg sync.WaitGroup
			for _, f := range funcs {
				wg.Add(1)

				go func(f func(ctx context.Context) <-chan error) {
					defer wg.Done()

					for err := range f(ctx) {
						errs <- err
					}
				}(f)
			}

			wg.Wait()
		})
	}
}

// withErrors creates a channel of errors, invokes the given function in
// a goroutine with the channel as a parameter, then returns the channel.
// The channel is closed after the given function unblocks (asynchronously).
func withErrors(fn func(chan<- error)) <-chan error {
	errs := make(chan error)

	go func() {
		defer close(errs)
		fn(errs)
	}()

	return errs
}

// runAsync creates a channel of errors, invokes the given function in a
// goroutine with the given context and the channel as a parameter, then
// returns the channel. The channel is closed after the given function
// unblocks (asynchronously).
func runAsync(ctx context.Context, fn func(ctx context.Context) error) <-chan error {
	return withErrors(func(results chan<- error) {
		results <- fn(ctx)
	})
}
