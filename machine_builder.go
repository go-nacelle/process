package process

import (
	"context"
	"sync"
)

type machineBuilder struct {
	injecter Injecter
	health   *Health
}

// closedErrorsChannel is a global, always closed channel of error values.
var closedErrorsChannel = make(chan error)

func init() {
	close(closedErrorsChannel)
}

func newMachineBuilder(configs ...MachineConfigFunc) *machineBuilder {
	b := &machineBuilder{
		injecter: InjecterFunc(func(ctx context.Context, meta *Meta) error { return nil }),
		health:   NewHealth(),
	}

	for _, f := range configs {
		f(b)
	}

	return b
}

// buildRun creates a function that initializers, runs, and monitors the processes
// registered to the given container. For each priority from low values to high values,
// the function will:
//
// 1. Run the inject hook for each process registered to the target priority. If the
// injecter provided to the builder is nil, this step is skipped. The injecter will
// assign fields of the processes registered at this priority. Initializers registered
// to lower priorities may provide values to be injected into processes registered to
// higher priorities.
//
// 2. Initialize each process. These methods are invoked in parallel. If an error occurs
// during initialization, the application will skip the following step for this and
// remaining priority groups.
//
// 3. Start each process. These methods are invoked in parallel as well. As processes
// are expected to be long-running (with exceptions), this phase has no obvious end.
// The next priority group of processes will begin to execute in the same manner after
// all of the processes registered to this priority have started and the process becomes
// healthy (or the health timeout for an unhealthyprocess elapses).
//
// On shutdown due to a user signal, an explicit request, or a process error, all of the
// processes registered to the given container are finalized. All finalizer methods are
// invoked in parallel.
func (b *machineBuilder) buildRun(container *Container) streamErrorFunc {
	n := 0
	for _, meta := range container.meta {
		n += len(meta)
	}

	var wg sync.WaitGroup
	processErrors := make(chan error, n)
	healthCheckCtx, healthCheckCancel := context.WithCancel(context.Background())

	var initAndRunEachPriority []streamErrorFunc
	for _, priority := range container.priorities {
		meta := container.meta[priority]

		var partitions [][]*Meta
		if priority == 0 {
			for _, meta := range meta {
				partitions = append(partitions, []*Meta{meta})
			}
		} else {
			partitions = append(partitions, meta)
		}

		var initEachPriority []streamErrorFunc
		for _, meta := range partitions {
			injectAtPriority := mapMetaParallel(meta, func(meta *Meta) streamErrorFunc {
				return toStreamErrorFunc(func(ctx context.Context) error {
					if b.injecter == nil {
						return nil
					}

					meta.logger.Info("Running inject hook for %s", meta.Name())

					if err := b.injecter.Inject(ctx, meta); err != nil {
						return &opError{
							source:   err,
							metaName: meta.Name(),
							opName:   "inject hook",
							message:  "failed",
						}
					}

					return nil
				})
			})

			initAtPriority := mapMetaParallel(meta, func(m *Meta) streamErrorFunc {
				return toStreamErrorFunc(m.Init)
			})

			initEachPriority = append(initEachPriority, chain(
				injectAtPriority,
				initAtPriority,
			))
		}

		runAtPriority := mapMetaParallel(meta, func(meta *Meta) streamErrorFunc {
			return func(ctx context.Context) <-chan error {
				wg.Add(1)

				go func() {
					defer wg.Done()

					if err := meta.Run(ctx); err != nil {
						healthCheckCancel()
						processErrors <- err
					}
				}()

				return closedErrorsChannel
			}
		})

		healthKeyMap := map[interface{}]struct{}{}
		for _, meta := range meta {
			for _, key := range meta.options.healthKeys {
				healthKeyMap[key] = struct{}{}
			}
		}
		healthKeys := make([]interface{}, 0, len(healthKeyMap))
		for key := range healthKeyMap {
			healthKeys = append(healthKeys, key)
		}

		waitUntilHealthy := toStreamErrorFunc(func(ctx context.Context) error {
			components, err := b.health.GetAll(healthKeys...)
			if err != nil {
				return err
			}
			if len(components) == 0 {
				return nil
			}

			ch, cancel := b.health.Subscribe()
			defer cancel()

			for {
				select {
				case <-ch:
				case <-healthCheckCtx.Done():
					return ErrHealthCheckCanceled
				}

				healthy := true
				for _, component := range components {
					if !component.Healthy() {
						healthy = false
						break
					}
				}

				if healthy {
					return nil
				}
			}
		})

		initAndRunEachPriority = append(initAndRunEachPriority, chain(
			chain(initEachPriority...),
			runAtPriority,
			waitUntilHealthy,
		))
	}

	forwardProcessErrors := func(ctx context.Context) <-chan error {
		go func() {
			wg.Wait()
			close(processErrors)
			healthCheckCancel()
		}()

		return processErrors
	}

	runFinalizers := mapMetaParallel(container.Meta(), func(m *Meta) streamErrorFunc {
		return toStreamErrorFunc(func(ctx context.Context) error {
			// TODO - does this neglect values from the original ctx?
			return m.Finalize(context.Background())
		})
	})

	return sequence(
		chain(initAndRunEachPriority...),
		forwardProcessErrors,
		runFinalizers,
	)
}

// buildShutdown creates a function that invokes the Stop function of each process registered
//to the given container. Processes registered to the same priority are stopped in parallel and
// processes with a higher priority are stopped before those registered to a lower priority.
func (b *machineBuilder) buildShutdown(container *Container) streamErrorFunc {
	var stopEachPriority []streamErrorFunc
	for i := len(container.priorities) - 1; i >= 0; i-- {
		metaAtPriority := container.meta[container.priorities[i]]

		stopEachPriority = append(stopEachPriority, mapMetaParallel(metaAtPriority, func(m *Meta) streamErrorFunc {
			return toStreamErrorFunc(m.Stop)
		}))
	}

	return sequence(stopEachPriority...)
}

// mapMetaParallel creates a function that executes the given function in parallel
// over each of the given meta values.
func mapMetaParallel(meta []*Meta, fn func(meta *Meta) streamErrorFunc) streamErrorFunc {
	var mapped []streamErrorFunc
	for _, meta := range meta {
		mapped = append(mapped, fn(meta))
	}

	return parallel(mapped...)
}
