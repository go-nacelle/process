package process

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunInitializer(t *testing.T) {
	trace := make(chan string, 72)
	builder := NewContainerBuilder()

	initializer := NewMockMaximumProcess()
	initializer.InitFunc.SetDefaultHook(traceInit(nil, trace, "a", 0, nil))
	builder.RegisterInitializer(initializer)

	state := Run(context.Background(), builder.Build())
	require.True(t, state.Wait(context.Background()))
	require.Empty(t, state.Errors())

	close(trace)
	assertChannelContents(t, readStringChannel(trace), seq("a.0.init"))
}

func TestRunFinalizingInitializer(t *testing.T) {
	trace := make(chan string, 72)
	builder := NewContainerBuilder()

	initializer := NewMockMaximumProcess()
	initializer.InitFunc.SetDefaultHook(traceInit(nil, trace, "a", 0, nil))
	initializer.FinalizeFunc.SetDefaultHook(traceFinalize(trace, "a", 0, nil))
	builder.RegisterInitializer(initializer)

	state := Run(context.Background(), builder.Build())
	require.True(t, state.Wait(context.Background()))
	require.Empty(t, state.Errors())

	close(trace)
	assertChannelContents(t, readStringChannel(trace), seq("a.0.init", "a.0.finalize"))
}

func TestRunMultipleInitializers(t *testing.T) {
	trace := make(chan string, 72)
	builder := NewContainerBuilder()

	for _, value := range []string{"a", "b", "c", "d"} {
		for i := 1; i <= 3; i++ {
			initializer := NewMockMaximumProcess()
			initializer.InitFunc.SetDefaultHook(traceInit(nil, trace, value, i, nil))
			builder.RegisterInitializer(initializer, WithMetaPriority(i))
		}
	}

	state := Run(context.Background(), builder.Build())
	require.True(t, state.Wait(context.Background()))
	require.Empty(t, state.Errors())

	close(trace)
	assertChannelContents(t, readStringChannel(trace), seq(
		unordered("a.1.init", "b.1.init", "c.1.init", "d.1.init"),
		unordered("a.2.init", "b.2.init", "c.2.init", "d.2.init"),
		unordered("a.3.init", "b.3.init", "c.3.init", "d.3.init"),
	))
}

func TestRunMultipleFinalizingInitializers(t *testing.T) {
	trace := make(chan string, 72)
	builder := NewContainerBuilder()

	for _, value := range []string{"a", "b", "c", "d"} {
		for i := 1; i <= 3; i++ {
			initializer := NewMockMaximumProcess()
			initializer.InitFunc.SetDefaultHook(traceInit(nil, trace, value, i, nil))
			initializer.FinalizeFunc.SetDefaultHook(traceFinalize(trace, value, i, nil))
			builder.RegisterInitializer(initializer, WithMetaPriority(i))
		}
	}

	state := Run(context.Background(), builder.Build())

	assertChannelContents(t, readStringChannel(forwardN(trace, 12)), seq(
		unordered("a.1.init", "b.1.init", "c.1.init", "d.1.init"),
		unordered("a.2.init", "b.2.init", "c.2.init", "d.2.init"),
		unordered("a.3.init", "b.3.init", "c.3.init", "d.3.init"),
	))

	state.Shutdown(context.Background())
	require.True(t, state.Wait(context.Background()))
	require.Empty(t, state.Errors())

	close(trace)
	assertChannelContents(t, readStringChannel(trace), seq(
		unordered(
			"a.1.finalize", "b.1.finalize", "c.1.finalize", "d.1.finalize",
			"a.2.finalize", "b.2.finalize", "c.2.finalize", "d.2.finalize",
			"a.3.finalize", "b.3.finalize", "c.3.finalize", "d.3.finalize",
		),
	))
}

func TestRunMultipleInitializersWithPriorityZero(t *testing.T) {
	trace := make(chan string, 72)
	builder := NewContainerBuilder()

	for _, value := range []string{"a", "b", "c", "d"} {
		initializer := NewMockMaximumProcess()
		initializer.InitFunc.SetDefaultHook(traceInit(nil, trace, value, 0, nil))
		builder.RegisterInitializer(initializer)
	}

	state := Run(context.Background(), builder.Build())
	require.True(t, state.Wait(context.Background()))
	require.Empty(t, state.Errors())

	close(trace)
	assertChannelContents(t, readStringChannel(trace), seq("a.0.init", "b.0.init", "c.0.init", "d.0.init"))
}

func TestRunProcess(t *testing.T) {
	health := NewHealth()
	trace := make(chan string, 72)
	builder := NewContainerBuilder()

	process := NewMockMaximumProcess()
	process.InitFunc.SetDefaultHook(traceInit(health, trace, "a", 0, nil))
	process.RunFunc.SetDefaultHook(traceRun(health, trace, "a", 0, nil))
	process.StopFunc.SetDefaultHook(traceStop(trace, "a", 0, nil))
	builder.RegisterProcess(process, WithMetaHealthKey(testHealthKey("a", 0)))

	state := Run(context.Background(), builder.Build(WithMetaHealth(health)), WithHealth(health))
	assertChannelContents(t, readStringChannel(forwardN(trace, 2)), seq("a.0.init", "a.0.run"))
	state.Shutdown(context.Background())
	require.True(t, state.Wait(context.Background()))
	require.Empty(t, state.Errors())
	close(trace)
	assertChannelContents(t, readStringChannel(trace), seq("a.0.stop"))
}

func TestRunFinalizingProcess(t *testing.T) {
	health := NewHealth()
	trace := make(chan string, 72)
	builder := NewContainerBuilder()

	process := NewMockMaximumProcess()
	process.InitFunc.SetDefaultHook(traceInit(health, trace, "a", 0, nil))
	process.RunFunc.SetDefaultHook(traceRun(health, trace, "a", 0, nil))
	process.StopFunc.SetDefaultHook(traceStop(trace, "a", 0, nil))
	process.FinalizeFunc.SetDefaultHook(traceFinalize(trace, "a", 0, nil))
	builder.RegisterProcess(process, WithMetaHealthKey(testHealthKey("a", 0)))

	state := Run(context.Background(), builder.Build(WithMetaHealth(health)), WithHealth(health))
	assertChannelContents(t, readStringChannel(forwardN(trace, 2)), seq("a.0.init", "a.0.run"))
	state.Shutdown(context.Background())
	require.True(t, state.Wait(context.Background()))
	require.Empty(t, state.Errors())
	close(trace)
	assertChannelContents(t, readStringChannel(trace), seq("a.0.stop", "a.0.finalize"))
}

func TestRunMultipleProcesses(t *testing.T) {
	health := NewHealth()
	trace := make(chan string, 72)
	builder := NewContainerBuilder()

	for _, value := range []string{"a", "b", "c", "d"} {
		for i := 1; i <= 3; i++ {
			process := NewMockMaximumProcess()
			process.InitFunc.SetDefaultHook(traceInit(health, trace, value, i, nil))
			process.RunFunc.SetDefaultHook(traceRun(health, trace, value, i, nil))
			process.StopFunc.SetDefaultHook(traceStop(trace, value, i, nil))
			builder.RegisterProcess(process, WithMetaPriority(i), WithMetaHealthKey(testHealthKey(value, i)))
		}
	}

	state := Run(context.Background(), builder.Build(WithMetaHealth(health)), WithHealth(health))

	assertChannelContents(t, readStringChannel(forwardN(trace, 24)), seq(
		unordered("a.1.init", "b.1.init", "c.1.init", "d.1.init"),
		unordered("a.1.run", "b.1.run", "c.1.run", "d.1.run"),
		unordered("a.2.init", "b.2.init", "c.2.init", "d.2.init"),
		unordered("a.2.run", "b.2.run", "c.2.run", "d.2.run"),
		unordered("a.3.init", "b.3.init", "c.3.init", "d.3.init"),
		unordered("a.3.run", "b.3.run", "c.3.run", "d.3.run"),
	))

	state.Shutdown(context.Background())
	require.True(t, state.Wait(context.Background()))
	require.Empty(t, state.Errors())

	close(trace)
	assertChannelContents(t, readStringChannel(trace), seq(
		unordered("a.3.stop", "b.3.stop", "c.3.stop", "d.3.stop"),
		unordered("a.2.stop", "b.2.stop", "c.2.stop", "d.2.stop"),
		unordered("a.1.stop", "b.1.stop", "c.1.stop", "d.1.stop"),
	))
}

func TestRunMultipleFinalizingProcesses(t *testing.T) {
	health := NewHealth()
	trace := make(chan string, 72)
	builder := NewContainerBuilder()

	for _, value := range []string{"a", "b", "c", "d"} {
		for i := 1; i <= 3; i++ {
			process := NewMockMaximumProcess()
			process.InitFunc.SetDefaultHook(traceInit(health, trace, value, i, nil))
			process.RunFunc.SetDefaultHook(traceRun(health, trace, value, i, nil))
			process.StopFunc.SetDefaultHook(traceStop(trace, value, i, nil))
			process.FinalizeFunc.SetDefaultHook(traceFinalize(trace, value, i, nil))
			builder.RegisterProcess(process, WithMetaPriority(i), WithMetaHealthKey(testHealthKey(value, i)))
		}
	}

	state := Run(context.Background(), builder.Build(WithMetaHealth(health)), WithHealth(health))

	assertChannelContents(t, readStringChannel(forwardN(trace, 24)), seq(
		unordered("a.1.init", "b.1.init", "c.1.init", "d.1.init"),
		unordered("a.1.run", "b.1.run", "c.1.run", "d.1.run"),
		unordered("a.2.init", "b.2.init", "c.2.init", "d.2.init"),
		unordered("a.2.run", "b.2.run", "c.2.run", "d.2.run"),
		unordered("a.3.init", "b.3.init", "c.3.init", "d.3.init"),
		unordered("a.3.run", "b.3.run", "c.3.run", "d.3.run"),
	))

	state.Shutdown(context.Background())
	require.True(t, state.Wait(context.Background()))
	require.Empty(t, state.Errors())

	close(trace)
	assertChannelContents(t, readStringChannel(trace), seq(
		unordered("a.3.stop", "b.3.stop", "c.3.stop", "d.3.stop"),
		unordered("a.2.stop", "b.2.stop", "c.2.stop", "d.2.stop"),
		unordered("a.1.stop", "b.1.stop", "c.1.stop", "d.1.stop"),
		unordered(
			"a.1.finalize", "b.1.finalize", "c.1.finalize", "d.1.finalize",
			"a.2.finalize", "b.2.finalize", "c.2.finalize", "d.2.finalize",
			"a.3.finalize", "b.3.finalize", "c.3.finalize", "d.3.finalize",
		),
	))
}

func TestRunMultipleFinalizingInitializersAndProcesses(t *testing.T) {
	health := NewHealth()
	trace := make(chan string, 72)
	builder := NewContainerBuilder()

	for _, value := range []string{"w", "x", "y", "z"} {
		for i := 1; i <= 3; i++ {
			initializer := NewMockMaximumProcess()
			initializer.InitFunc.SetDefaultHook(traceInit(nil, trace, value, i, nil))
			initializer.FinalizeFunc.SetDefaultHook(traceFinalize(trace, value, i, nil))
			builder.RegisterInitializer(initializer, WithMetaPriority(i))
		}
	}

	for _, value := range []string{"a", "b", "c", "d"} {
		for i := 1; i <= 3; i++ {
			process := NewMockMaximumProcess()
			process.InitFunc.SetDefaultHook(traceInit(health, trace, value, i, nil))
			process.RunFunc.SetDefaultHook(traceRun(health, trace, value, i, nil))
			process.StopFunc.SetDefaultHook(traceStop(trace, value, i, nil))
			process.FinalizeFunc.SetDefaultHook(traceFinalize(trace, value, i, nil))
			builder.RegisterProcess(process, WithMetaPriority(i), WithMetaHealthKey(testHealthKey(value, i)))
		}
	}

	state := Run(context.Background(), builder.Build(WithMetaHealth(health)), WithHealth(health))

	assertChannelContents(t, readStringChannel(forwardN(trace, 36)), seq(
		unordered("w.1.init", "x.1.init", "y.1.init", "z.1.init", "a.1.init", "b.1.init", "c.1.init", "d.1.init"),
		unordered("a.1.run", "b.1.run", "c.1.run", "d.1.run"),
		unordered("w.2.init", "x.2.init", "y.2.init", "z.2.init", "a.2.init", "b.2.init", "c.2.init", "d.2.init"),
		unordered("a.2.run", "b.2.run", "c.2.run", "d.2.run"),
		unordered("w.3.init", "x.3.init", "y.3.init", "z.3.init", "a.3.init", "b.3.init", "c.3.init", "d.3.init"),
		unordered("a.3.run", "b.3.run", "c.3.run", "d.3.run"),
	))

	state.Shutdown(context.Background())
	require.True(t, state.Wait(context.Background()))
	require.Empty(t, state.Errors())

	close(trace)
	assertChannelContents(t, readStringChannel(trace), seq(
		unordered("a.3.stop", "b.3.stop", "c.3.stop", "d.3.stop"),
		unordered("a.2.stop", "b.2.stop", "c.2.stop", "d.2.stop"),
		unordered("a.1.stop", "b.1.stop", "c.1.stop", "d.1.stop"),
		unordered(
			"a.1.finalize", "b.1.finalize", "c.1.finalize", "d.1.finalize",
			"a.2.finalize", "b.2.finalize", "c.2.finalize", "d.2.finalize",
			"a.3.finalize", "b.3.finalize", "c.3.finalize", "d.3.finalize",
			"w.1.finalize", "x.1.finalize", "y.1.finalize", "z.1.finalize",
			"w.2.finalize", "x.2.finalize", "y.2.finalize", "z.2.finalize",
			"w.3.finalize", "x.3.finalize", "y.3.finalize", "z.3.finalize",
		),
	))
}

func TestRunProcessExitsWithEarlyExit(t *testing.T) {
	health := NewHealth()
	trace := make(chan string, 72)
	builder := NewContainerBuilder()

	for _, value := range []string{"w", "x", "y", "z"} {
		for i := 1; i <= 3; i++ {
			initializer := NewMockMaximumProcess()
			initializer.InitFunc.SetDefaultHook(traceInit(nil, trace, value, i, nil))
			initializer.FinalizeFunc.SetDefaultHook(traceFinalize(trace, value, i, nil))
			builder.RegisterInitializer(initializer, WithMetaPriority(i))
		}
	}

	var processes []*MockMaximumProcess
	for _, value := range []string{"a", "b", "c", "d"} {
		for i := 1; i <= 3; i++ {
			process := NewMockMaximumProcess()
			process.InitFunc.SetDefaultHook(traceInit(health, trace, value, i, nil))
			process.RunFunc.SetDefaultHook(traceRun(health, trace, value, i, nil))
			process.StopFunc.SetDefaultHook(traceStop(trace, value, i, nil))
			process.FinalizeFunc.SetDefaultHook(traceFinalize(trace, value, i, nil))
			builder.RegisterProcess(process, WithMetaPriority(i), WithEarlyExit(true), WithMetaHealthKey(testHealthKey(value, i)))
			processes = append(processes, process)
		}
	}

	//  0  1  2  3  4  5  6  7 ...
	// a1 a2 a3 b1 b2 b3 c1 c2 ...
	processes[7].RunFunc.SetDefaultHook(func(ctx context.Context) error {
		healthComponent, _ := health.Get("c.2")
		healthComponent.Update(true)
		trace <- "boom"
		return nil
	})

	state := Run(context.Background(), builder.Build(WithMetaHealth(health)), WithHealth(health))

	assertChannelContents(t, readStringChannel(forwardN(trace, 36)), seq(
		unordered("w.1.init", "x.1.init", "y.1.init", "z.1.init", "a.1.init", "b.1.init", "c.1.init", "d.1.init"),
		unordered("a.1.run", "b.1.run", "c.1.run", "d.1.run"),
		unordered("w.2.init", "x.2.init", "y.2.init", "z.2.init", "a.2.init", "b.2.init", "c.2.init", "d.2.init"),
		unordered("a.2.run", "b.2.run", "boom", "d.2.run"),
		unordered("a.3.init", "b.3.init", "c.3.init", "d.3.init", "w.3.init", "x.3.init", "y.3.init", "z.3.init"),
		unordered("a.3.run", "b.3.run", "c.3.run", "d.3.run"),
	))

	state.Shutdown(context.Background())
	require.True(t, state.Wait(context.Background()))
	require.Empty(t, state.Errors())

	close(trace)
	assertChannelContents(t, readStringChannel(trace), seq(
		unordered("a.3.stop", "b.3.stop", "c.3.stop", "d.3.stop"),
		unordered("a.2.stop", "b.2.stop", "d.2.stop"),
		unordered("a.1.stop", "b.1.stop", "c.1.stop", "d.1.stop"),
		unordered(
			"a.1.finalize", "b.1.finalize", "c.1.finalize", "d.1.finalize",
			"a.2.finalize", "b.2.finalize", "c.2.finalize", "d.2.finalize",
			"a.3.finalize", "b.3.finalize", "c.3.finalize", "d.3.finalize",
			"w.1.finalize", "x.1.finalize", "y.1.finalize", "z.1.finalize",
			"w.2.finalize", "x.2.finalize", "y.2.finalize", "z.2.finalize",
			"w.3.finalize", "x.3.finalize", "y.3.finalize", "z.3.finalize",
		),
	))
}

func TestRunProcessExitsWithoutEarlyExit(t *testing.T) {
	health := NewHealth()
	trace := make(chan string, 72)
	builder := NewContainerBuilder()

	for _, value := range []string{"w", "x", "y", "z"} {
		for i := 1; i <= 3; i++ {
			initializer := NewMockMaximumProcess()
			initializer.InitFunc.SetDefaultHook(traceInit(nil, trace, value, i, nil))
			initializer.FinalizeFunc.SetDefaultHook(traceFinalize(trace, value, i, nil))
			builder.RegisterInitializer(initializer, WithMetaPriority(i))
		}
	}

	var processes []*MockMaximumProcess
	for _, value := range []string{"a", "b", "c", "d"} {
		for i := 1; i <= 3; i++ {
			process := NewMockMaximumProcess()
			process.InitFunc.SetDefaultHook(traceInit(health, trace, value, i, nil))
			process.RunFunc.SetDefaultHook(traceRun(health, trace, value, i, nil))
			process.StopFunc.SetDefaultHook(traceStop(trace, value, i, nil))
			process.FinalizeFunc.SetDefaultHook(traceFinalize(trace, value, i, nil))
			builder.RegisterProcess(process, WithMetaPriority(i), WithMetaHealthKey(testHealthKey(value, i)))
			processes = append(processes, process)
		}
	}

	//  0  1  2  3  4  5  6  7 ...
	// a1 a2 a3 b1 b2 b3 c1 c2 ...
	processes[7].RunFunc.SetDefaultHook(func(ctx context.Context) error {
		trace <- "boom"
		return nil
	})

	state := Run(context.Background(), builder.Build(WithMetaHealth(health)), WithHealth(health))

	assertChannelContents(t, readStringChannel(forwardN(trace, 24)), seq(
		unordered("w.1.init", "x.1.init", "y.1.init", "z.1.init", "a.1.init", "b.1.init", "c.1.init", "d.1.init"),
		unordered("a.1.run", "b.1.run", "c.1.run", "d.1.run"),
		unordered("w.2.init", "x.2.init", "y.2.init", "z.2.init", "a.2.init", "b.2.init", "c.2.init", "d.2.init"),
		unordered("a.2.run", "b.2.run", "boom", "d.2.run"),
	))

	require.False(t, state.Wait(context.Background()))

	assert.ElementsMatch(t,
		state.Errors(),
		[]error{
			fmt.Errorf("health check canceled"),
			fmt.Errorf("unexpected return from process"),
		},
	)

	close(trace)
	assertChannelContents(t, readStringChannel(trace), seq(
		unordered("a.2.stop", "b.2.stop", "d.2.stop"),
		unordered("a.1.stop", "b.1.stop", "c.1.stop", "d.1.stop"),
		unordered(
			"a.1.finalize", "b.1.finalize", "c.1.finalize", "d.1.finalize",
			"a.2.finalize", "b.2.finalize", "c.2.finalize", "d.2.finalize",
			"w.1.finalize", "x.1.finalize", "y.1.finalize", "z.1.finalize",
			"w.2.finalize", "x.2.finalize", "y.2.finalize", "z.2.finalize",
		),
	))
}

func TestRunInitializerInitError(t *testing.T) {
	health := NewHealth()
	trace := make(chan string, 72)
	builder := NewContainerBuilder()

	var initializers []*MockMaximumProcess
	for _, value := range []string{"w", "x", "y", "z"} {
		for i := 1; i <= 3; i++ {
			initializer := NewMockMaximumProcess()
			initializer.InitFunc.SetDefaultHook(traceInit(nil, trace, value, i, nil))
			initializer.FinalizeFunc.SetDefaultHook(traceFinalize(trace, value, i, nil))
			builder.RegisterInitializer(initializer, WithMetaPriority(i))
			initializers = append(initializers, initializer)
		}
	}

	//  0  1  2  3  4  5  6  7 ...
	// w1 w2 w3 x1 x2 x3 y1 y2 ...
	initializers[7].InitFunc.SetDefaultHook(func(ctx context.Context) error {
		trace <- "boom"
		return testErr1
	})

	state := Run(context.Background(), builder.Build(WithMetaHealth(health)), WithHealth(health))

	assertChannelContents(t, readStringChannel(forwardN(trace, 8)), seq(
		unordered("w.1.init", "x.1.init", "y.1.init", "z.1.init"),
		unordered("w.2.init", "x.2.init", "boom", "z.2.init"),
	))

	require.False(t, state.Wait(context.Background()))

	assert.ElementsMatch(t,
		state.Errors(),
		[]error{
			&opError{
				source:   fmt.Errorf("oops1"),
				opName:   "init",
				message:  "failed",
				metaName: "<unnamed *process.MockMaximumProcess>",
			},
		},
	)

	close(trace)
	assertChannelContents(t, readStringChannel(trace), seq(
		unordered(
			"w.1.finalize", "x.1.finalize", "y.1.finalize", "z.1.finalize",
			"w.2.finalize", "x.2.finalize", "z.2.finalize",
		),
	))
}

func TestRunProcessInitError(t *testing.T) {
	health := NewHealth()
	trace := make(chan string, 72)
	builder := NewContainerBuilder()

	for _, value := range []string{"w", "x", "y", "z"} {
		for i := 1; i <= 3; i++ {
			initializer := NewMockMaximumProcess()
			initializer.InitFunc.SetDefaultHook(traceInit(nil, trace, value, i, nil))
			initializer.FinalizeFunc.SetDefaultHook(traceFinalize(trace, value, i, nil))
			builder.RegisterInitializer(initializer, WithMetaPriority(i))
		}
	}

	var processes []*MockMaximumProcess
	for _, value := range []string{"a", "b", "c", "d"} {
		for i := 1; i <= 3; i++ {
			process := NewMockMaximumProcess()
			process.InitFunc.SetDefaultHook(traceInit(health, trace, value, i, nil))
			process.RunFunc.SetDefaultHook(traceRun(health, trace, value, i, nil))
			process.StopFunc.SetDefaultHook(traceStop(trace, value, i, nil))
			process.FinalizeFunc.SetDefaultHook(traceFinalize(trace, value, i, nil))
			builder.RegisterProcess(process, WithMetaPriority(i), WithMetaHealthKey(testHealthKey(value, i)))
			processes = append(processes, process)
		}
	}

	//  0  1  2  3  4  5  6  7 ...
	// a1 a2 a3 b1 b2 b3 c1 c2 ...
	processes[7].InitFunc.SetDefaultHook(func(ctx context.Context) error {
		trace <- "boom"
		return testErr1
	})

	state := Run(context.Background(), builder.Build(WithMetaHealth(health)), WithHealth(health))

	assertChannelContents(t, readStringChannel(forwardN(trace, 20)), seq(
		unordered("w.1.init", "x.1.init", "y.1.init", "z.1.init", "a.1.init", "b.1.init", "c.1.init", "d.1.init"),
		unordered("a.1.run", "b.1.run", "c.1.run", "d.1.run"),
		unordered("w.2.init", "x.2.init", "y.2.init", "z.2.init", "a.2.init", "b.2.init", "boom", "d.2.init"),
	))

	require.False(t, state.Wait(context.Background()))

	assert.ElementsMatch(t,
		state.Errors(),
		[]error{
			&opError{
				source:   fmt.Errorf("oops1"),
				opName:   "init",
				message:  "failed",
				metaName: "<unnamed *process.MockMaximumProcess>",
			},
		},
	)

	close(trace)
	assertChannelContents(t, readStringChannel(trace), seq(
		unordered("a.1.stop", "b.1.stop", "c.1.stop", "d.1.stop"),
		unordered(
			"a.1.finalize", "b.1.finalize", "c.1.finalize", "d.1.finalize",
			"a.2.finalize", "b.2.finalize", "d.2.finalize",
			"w.1.finalize", "x.1.finalize", "y.1.finalize", "z.1.finalize",
			"w.2.finalize", "x.2.finalize", "y.2.finalize", "z.2.finalize",
		),
	))
}

func TestRunProcessRunError(t *testing.T) {
	health := NewHealth()
	trace := make(chan string, 72)
	builder := NewContainerBuilder()

	for _, value := range []string{"w", "x", "y", "z"} {
		for i := 1; i <= 3; i++ {
			initializer := NewMockMaximumProcess()
			initializer.InitFunc.SetDefaultHook(traceInit(nil, trace, value, i, nil))
			initializer.FinalizeFunc.SetDefaultHook(traceFinalize(trace, value, i, nil))
			builder.RegisterInitializer(initializer, WithMetaPriority(i))
		}
	}

	var processes []*MockMaximumProcess
	for _, value := range []string{"a", "b", "c", "d"} {
		for i := 1; i <= 3; i++ {
			process := NewMockMaximumProcess()
			process.InitFunc.SetDefaultHook(traceInit(health, trace, value, i, nil))
			process.RunFunc.SetDefaultHook(traceRun(health, trace, value, i, nil))
			process.StopFunc.SetDefaultHook(traceStop(trace, value, i, nil))
			process.FinalizeFunc.SetDefaultHook(traceFinalize(trace, value, i, nil))
			builder.RegisterProcess(process, WithMetaPriority(i), WithMetaHealthKey(testHealthKey(value, i)))
			processes = append(processes, process)
		}
	}

	//  0  1  2  3  4  5  6  7 ...
	// a1 a2 a3 b1 b2 b3 c1 c2 ...
	processes[7].RunFunc.SetDefaultHook(func(ctx context.Context) error {
		trace <- "boom"
		return testErr1
	})

	state := Run(context.Background(), builder.Build(WithMetaHealth(health)), WithHealth(health))

	assertChannelContents(t, readStringChannel(forwardN(trace, 24)), seq(
		unordered("w.1.init", "x.1.init", "y.1.init", "z.1.init", "a.1.init", "b.1.init", "c.1.init", "d.1.init"),
		unordered("a.1.run", "b.1.run", "c.1.run", "d.1.run"),
		unordered("w.2.init", "x.2.init", "y.2.init", "z.2.init", "a.2.init", "b.2.init", "c.2.init", "d.2.init"),
		unordered("a.2.run", "b.2.run", "boom", "d.2.run"),
	))

	require.False(t, state.Wait(context.Background()))

	assert.ElementsMatch(t,
		state.Errors(),
		[]error{
			fmt.Errorf("health check canceled"),
			&opError{
				source:   fmt.Errorf("oops1"),
				metaName: "<unnamed *process.MockMaximumProcess>",
				opName:   "run",
				message:  "failed",
			},
		},
	)

	close(trace)
	assertChannelContents(t, readStringChannel(trace), seq(
		unordered("a.2.stop", "b.2.stop", "d.2.stop"),
		unordered("a.1.stop", "b.1.stop", "c.1.stop", "d.1.stop"),
		unordered(
			"a.1.finalize", "b.1.finalize", "c.1.finalize", "d.1.finalize",
			"a.2.finalize", "b.2.finalize", "c.2.finalize", "d.2.finalize",
			"w.1.finalize", "x.1.finalize", "y.1.finalize", "z.1.finalize",
			"w.2.finalize", "x.2.finalize", "y.2.finalize", "z.2.finalize",
		),
	))
}

func TestCancelContextOnUnhealthyProcess(t *testing.T) {
	health := NewHealth()
	trace := make(chan string, 72)
	builder := NewContainerBuilder()

	process := NewMockMaximumProcess()
	process.InitFunc.SetDefaultHook(traceInit(health, trace, "a", 0, nil))
	process.RunFunc.SetDefaultHook(traceRun(nil, trace, "a", 0, nil)) // note: no health here
	builder.RegisterProcess(process, WithMetaHealthKey(testHealthKey("a", 0)))

	ctx, cancel := context.WithCancel(context.Background())
	state := Run(ctx, builder.Build(WithMetaHealth(health)), WithHealth(health))
	assertChannelContents(t, readStringChannel(forwardN(trace, 2)), seq("a.0.init", "a.0.run"))
	cancel()
	require.False(t, state.Wait(context.Background()))
	close(trace)
}

func TestRunErrorsDuringShutdown(t *testing.T) {
	health := NewHealth()
	trace := make(chan string, 72)
	builder := NewContainerBuilder()

	var expectedErrs []error
	for _, value := range []string{"w", "x", "y", "z"} {
		for i := 1; i <= 3; i++ {
			sourceErr := fmt.Errorf("%s.%d", value, i)
			initializer := NewMockMaximumProcess()
			initializer.InitFunc.SetDefaultHook(traceInit(nil, trace, value, i, nil))
			initializer.FinalizeFunc.SetDefaultHook(traceFinalize(trace, value, i, sourceErr))
			builder.RegisterInitializer(initializer, WithMetaPriority(i))

			expectedErrs = append(expectedErrs, &opError{
				source:   sourceErr,
				opName:   "finalize",
				message:  "failed",
				metaName: "<unnamed *process.MockMaximumProcess>",
			})
		}
	}

	for _, value := range []string{"a", "b", "c", "d"} {
		for i := 1; i <= 3; i++ {
			sourceErr := fmt.Errorf("%s.%d", value, i)
			process := NewMockMaximumProcess()
			process.InitFunc.SetDefaultHook(traceInit(health, trace, value, i, nil))
			process.RunFunc.SetDefaultHook(traceRun(health, trace, value, i, nil))
			process.StopFunc.SetDefaultHook(traceStop(trace, value, i, sourceErr))
			process.FinalizeFunc.SetDefaultHook(traceFinalize(trace, value, i, sourceErr))
			builder.RegisterProcess(process, WithMetaPriority(i), WithMetaHealthKey(testHealthKey(value, i)))

			expectedErrs = append(expectedErrs, []error{
				&opError{
					source:   sourceErr,
					opName:   "stop",
					message:  "failed",
					metaName: "<unnamed *process.MockMaximumProcess>",
				},
				&opError{
					source:   sourceErr,
					opName:   "finalize",
					message:  "failed",
					metaName: "<unnamed *process.MockMaximumProcess>",
				},
			}...)
		}
	}

	state := Run(context.Background(), builder.Build(WithMetaHealth(health)), WithHealth(health))

	assertChannelContents(t, readStringChannel(forwardN(trace, 36)), []interface{}{
		unordered("w.1.init", "x.1.init", "y.1.init", "z.1.init", "a.1.init", "b.1.init", "c.1.init", "d.1.init"),
		unordered("a.1.run", "b.1.run", "c.1.run", "d.1.run"),
		unordered("a.2.init", "b.2.init", "c.2.init", "d.2.init", "w.2.init", "x.2.init", "y.2.init", "z.2.init"),
		unordered("a.2.run", "b.2.run", "c.2.run", "d.2.run"),
		unordered("w.3.init", "x.3.init", "y.3.init", "z.3.init", "a.3.init", "b.3.init", "c.3.init", "d.3.init"),
		unordered("a.3.run", "b.3.run", "c.3.run", "d.3.run"),
	})

	state.Shutdown(context.Background())
	require.False(t, state.Wait(context.Background()))

	assert.ElementsMatch(t, state.Errors(), expectedErrs)

	close(trace)
	assertChannelContents(t, readStringChannel(trace), []interface{}{
		unordered("a.3.stop", "b.3.stop", "c.3.stop", "d.3.stop"),
		unordered("a.2.stop", "b.2.stop", "c.2.stop", "d.2.stop"),
		unordered("a.1.stop", "b.1.stop", "c.1.stop", "d.1.stop"),
		unordered(
			"a.1.finalize", "b.1.finalize", "c.1.finalize", "d.1.finalize",
			"a.2.finalize", "b.2.finalize", "c.2.finalize", "d.2.finalize",
			"a.3.finalize", "b.3.finalize", "c.3.finalize", "d.3.finalize",
			"w.1.finalize", "x.1.finalize", "y.1.finalize", "z.1.finalize",
			"w.2.finalize", "x.2.finalize", "y.2.finalize", "z.2.finalize",
			"w.3.finalize", "x.3.finalize", "y.3.finalize", "z.3.finalize",
		),
	})
}
