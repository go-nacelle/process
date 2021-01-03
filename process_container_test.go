package process

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessContainerInitializers(t *testing.T) {
	c := NewProcessContainer()
	c.RegisterInitializer(InitializerFunc(func(ctx context.Context) error { return fmt.Errorf("a") }))
	c.RegisterInitializer(InitializerFunc(func(ctx context.Context) error { return fmt.Errorf("b") }), WithInitializerName("b"), WithInitializerPriority(5))
	c.RegisterInitializer(InitializerFunc(func(ctx context.Context) error { return fmt.Errorf("c") }), WithInitializerName("c"), WithInitializerPriority(2))
	c.RegisterInitializer(InitializerFunc(func(ctx context.Context) error { return fmt.Errorf("d") }), WithInitializerName("d"), WithInitializerPriority(3))
	c.RegisterInitializer(InitializerFunc(func(ctx context.Context) error { return fmt.Errorf("e") }), WithInitializerName("e"), WithInitializerPriority(2))
	c.RegisterInitializer(InitializerFunc(func(ctx context.Context) error { return fmt.Errorf("f") }), WithInitializerName("f"))

	require.Equal(t, 6, c.NumInitializers())
	require.Equal(t, 4, c.NumInitializerPriorities())

	p1 := c.GetInitializersAtPriorityIndex(0)
	p2 := c.GetInitializersAtPriorityIndex(1)
	p3 := c.GetInitializersAtPriorityIndex(2)
	p4 := c.GetInitializersAtPriorityIndex(3)

	require.Len(t, p1, 2)
	require.Len(t, p2, 2)
	require.Len(t, p3, 1)
	require.Len(t, p4, 1)

	// Test priorities
	assert.Equal(t, 0, p1[0].priority)
	assert.Equal(t, 2, p2[0].priority)
	assert.Equal(t, 3, p3[0].priority)
	assert.Equal(t, 5, p4[0].priority)

	// Test names + order
	assert.Equal(t, "<unnamed>", p1[0].Name())
	assert.Equal(t, "f", p1[1].Name())
	assert.Equal(t, "c", p2[0].Name())
	assert.Equal(t, "e", p2[1].Name())
	assert.Equal(t, "d", p3[0].Name())
	assert.Equal(t, "b", p4[0].Name())

	// Test inner function
	assert.EqualError(t, p1[0].Initializer.Init(context.Background()), "a")
	assert.EqualError(t, p1[1].Initializer.Init(context.Background()), "f")
	assert.EqualError(t, p2[0].Initializer.Init(context.Background()), "c")
	assert.EqualError(t, p2[1].Initializer.Init(context.Background()), "e")
	assert.EqualError(t, p3[0].Initializer.Init(context.Background()), "d")
	assert.EqualError(t, p4[0].Initializer.Init(context.Background()), "b")
}

func TestProcessContainerProcesses(t *testing.T) {
	c := NewProcessContainer()
	c.RegisterProcess(newInitFailProcess("a"))
	c.RegisterProcess(newInitFailProcess("b"), WithProcessName("b"), WithProcessPriority(5))
	c.RegisterProcess(newInitFailProcess("c"), WithProcessName("c"), WithProcessPriority(2))
	c.RegisterProcess(newInitFailProcess("d"), WithProcessName("d"), WithProcessPriority(3))
	c.RegisterProcess(newInitFailProcess("e"), WithProcessName("e"), WithProcessPriority(2))
	c.RegisterProcess(newInitFailProcess("f"), WithProcessName("f"))

	require.Equal(t, 6, c.NumProcesses())
	require.Equal(t, 4, c.NumProcessPriorities())

	p1 := c.GetProcessesAtPriorityIndex(0)
	p2 := c.GetProcessesAtPriorityIndex(1)
	p3 := c.GetProcessesAtPriorityIndex(2)
	p4 := c.GetProcessesAtPriorityIndex(3)

	require.Len(t, p1, 2)
	require.Len(t, p2, 2)
	require.Len(t, p3, 1)
	require.Len(t, p4, 1)

	// Test priorities
	assert.Equal(t, 0, p1[0].priority)
	assert.Equal(t, 2, p2[0].priority)
	assert.Equal(t, 3, p3[0].priority)
	assert.Equal(t, 5, p4[0].priority)

	// Test names + order
	assert.Equal(t, "<unnamed>", p1[0].Name())
	assert.Equal(t, "f", p1[1].Name())
	assert.Equal(t, "c", p2[0].Name())
	assert.Equal(t, "e", p2[1].Name())
	assert.Equal(t, "d", p3[0].Name())
	assert.Equal(t, "b", p4[0].Name())

	// Test inner function
	assert.EqualError(t, p1[0].Process.Init(context.Background()), "a")
	assert.EqualError(t, p1[1].Process.Init(context.Background()), "f")
	assert.EqualError(t, p2[0].Process.Init(context.Background()), "c")
	assert.EqualError(t, p2[1].Process.Init(context.Background()), "e")
	assert.EqualError(t, p3[0].Process.Init(context.Background()), "d")
	assert.EqualError(t, p4[0].Process.Init(context.Background()), "b")
}

//
//

type initFailProcess struct {
	name string
}

func newInitFailProcess(name string) Process {
	return &initFailProcess{name: name}
}

func (p *initFailProcess) Init(ctx context.Context) error {
	return fmt.Errorf("%s", p.name)
}

func (p *initFailProcess) Start(ctx context.Context) error { return nil }
func (p *initFailProcess) Stop(ctx context.Context) error  { return nil }
