package process

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-nacelle/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessContainerInitializers(t *testing.T) {
	i1 := InitializerFunc(func(config.Config) error { return fmt.Errorf("a") })
	i2 := InitializerFunc(func(config.Config) error { return fmt.Errorf("b") })
	i3 := InitializerFunc(func(config.Config) error { return fmt.Errorf("c") })

	c := NewProcessContainer()
	c.RegisterInitializer(i1)
	c.RegisterInitializer(i2, WithInitializerName("b"))
	c.RegisterInitializer(i3, WithInitializerName("c"), WithInitializerTimeout(time.Minute*2))

	initializers := c.GetInitializers()
	require.Len(t, initializers, 3)

	// Test names
	assert.Equal(t, "<unnamed>", initializers[0].Name())
	assert.Equal(t, "b", initializers[1].Name())
	assert.Equal(t, "c", initializers[2].Name())

	// Test timeout
	assert.Equal(t, time.Second*0, initializers[0].InitTimeout())
	assert.Equal(t, time.Second*0, initializers[1].InitTimeout())
	assert.Equal(t, time.Minute*2, initializers[2].InitTimeout())

	// Test inner function
	assert.EqualError(t, initializers[0].Initializer.Init(nil), "a")
	assert.EqualError(t, initializers[1].Initializer.Init(nil), "b")
	assert.EqualError(t, initializers[2].Initializer.Init(nil), "c")
}

func TestProcessContainerProcesses(t *testing.T) {
	c := NewProcessContainer()
	c.RegisterProcess(newInitFailProcess("a"))
	c.RegisterProcess(newInitFailProcess("b"), WithProcessName("b"), WithPriority(5))
	c.RegisterProcess(newInitFailProcess("c"), WithProcessName("c"), WithPriority(2))
	c.RegisterProcess(newInitFailProcess("d"), WithProcessName("d"), WithPriority(3))
	c.RegisterProcess(newInitFailProcess("e"), WithProcessName("e"), WithPriority(2))
	c.RegisterProcess(newInitFailProcess("f"), WithProcessName("f"))

	require.Equal(t, 6, c.NumProcesses())
	require.Equal(t, 4, c.NumPriorities())

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
	assert.EqualError(t, p1[0].Process.Init(nil), "a")
	assert.EqualError(t, p1[1].Process.Init(nil), "f")
	assert.EqualError(t, p2[0].Process.Init(nil), "c")
	assert.EqualError(t, p2[1].Process.Init(nil), "e")
	assert.EqualError(t, p3[0].Process.Init(nil), "d")
	assert.EqualError(t, p4[0].Process.Init(nil), "b")
}

//
//

type initFailProcess struct {
	name string
}

func newInitFailProcess(name string) Process {
	return &initFailProcess{name: name}
}

func (p *initFailProcess) Init(config config.Config) error { return fmt.Errorf("%s", p.name) }
func (p *initFailProcess) Start() error                    { return nil }
func (p *initFailProcess) Stop() error                     { return nil }
