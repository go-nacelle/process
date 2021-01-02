package process

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-nacelle/log"
	"github.com/go-nacelle/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParallelInitializerRegisterConfiguration(t *testing.T) {
	container := service.NewServiceContainer()
	init := make(chan string, 3)
	finalize := make(chan string, 3)

	i1 := newTaggedFinalizer(init, finalize, "a")
	i2 := newTaggedInitializer(init, "b")
	i3 := newTaggedFinalizer(init, finalize, "c")

	pi := NewParallelInitializer(
		WithParallelInitializerContainer(container),
		WithParallelInitializerLogger(log.NewNilLogger()),
	)

	// Register things
	pi.RegisterInitializer(i1)
	pi.RegisterInitializer(i2)
	pi.RegisterInitializer(i3)

	pi.RegisterConfiguration(nil)
	assert.True(t, i1.registerConfigurationCalled)
	assert.True(t, i2.registerConfigurationCalled)
	assert.True(t, i3.registerConfigurationCalled)
}

func TestParallelInitializerInitialize(t *testing.T) {
	container := service.NewServiceContainer()
	init := make(chan string, 3)
	finalize := make(chan string, 3)

	i1 := newTaggedFinalizer(init, finalize, "a")
	i2 := newTaggedInitializer(init, "b")
	i3 := newTaggedFinalizer(init, finalize, "c")

	pi := NewParallelInitializer(
		WithParallelInitializerContainer(container),
		WithParallelInitializerLogger(log.NewNilLogger()),
	)

	// Register things
	pi.RegisterInitializer(i1)
	pi.RegisterInitializer(i2)
	pi.RegisterInitializer(i3)

	require.Nil(t, pi.Init(context.Background()))

	// May initialize in any order
	eventually(t, stringChanReceivesUnordered(init, "a", "b", "c"))
}

func TestParallelInitializerInitError(t *testing.T) {
	container := service.NewServiceContainer()
	init := make(chan string, 4)
	finalize := make(chan string, 4)

	i1 := newTaggedFinalizer(init, finalize, "a")
	i2 := newTaggedFinalizer(init, finalize, "b")
	i3 := newTaggedFinalizer(init, finalize, "c")
	i4 := newTaggedFinalizer(init, finalize, "d")
	m1 := newInitializerMeta(i1)
	m2 := newInitializerMeta(i2)
	m3 := newInitializerMeta(i3)
	m4 := newInitializerMeta(i4)

	pi := NewParallelInitializer(
		WithParallelInitializerContainer(container),
		WithParallelInitializerLogger(log.NewNilLogger()),
	)

	// Register things
	pi.RegisterInitializer(i1, WithInitializerName("a"))
	pi.RegisterInitializer(i2, WithInitializerName("b"))
	pi.RegisterInitializer(i3, WithInitializerName("c"))
	pi.RegisterInitializer(i4, WithInitializerName("d"))

	i2.initErr = fmt.Errorf("oops y")
	i3.initErr = fmt.Errorf("oops z")
	i4.finalizeErr = fmt.Errorf("oops w")

	WithInitializerName("a")(m1)
	WithInitializerName("b")(m2)
	WithInitializerName("c")(m3)
	WithInitializerName("d")(m4)

	err := pi.Init(context.Background())
	require.NotNil(t, err)

	expected := []errMeta{
		{err: fmt.Errorf("failed to initialize b (oops y)"), source: m2},
		{err: fmt.Errorf("failed to initialize c (oops z)"), source: m3},
		{err: fmt.Errorf("d returned error from finalize (oops w)"), source: m4},
	}
	for _, value := range expected {
		assert.Contains(t, err, value)
	}

	eventually(t, stringChanReceivesUnordered(init, "a", "b", "c", "d"))
	eventually(t, stringChanReceivesUnordered(finalize, "a", "d"))
}

func TestParallelInitializerFinalize(t *testing.T) {
	container := service.NewServiceContainer()
	init := make(chan string, 3)
	finalize := make(chan string, 3)

	i1 := newTaggedFinalizer(init, finalize, "a")
	i2 := newTaggedInitializer(init, "b")
	i3 := newTaggedFinalizer(init, finalize, "c")

	pi := NewParallelInitializer(
		WithParallelInitializerContainer(container),
		WithParallelInitializerLogger(log.NewNilLogger()),
	)

	// Register things
	pi.RegisterInitializer(i1)
	pi.RegisterInitializer(i2)
	pi.RegisterInitializer(i3)

	require.Nil(t, pi.Finalize(context.Background()))

	// May finalize in any order
	eventually(t, stringChanReceivesUnordered(finalize, "a", "c"))
}

func TestParallelInitializerFinalizeError(t *testing.T) {
	container := service.NewServiceContainer()
	init := make(chan string, 3)
	finalize := make(chan string, 3)

	i1 := newTaggedFinalizer(init, finalize, "a")
	i2 := newTaggedFinalizer(init, finalize, "b")
	i3 := newTaggedFinalizer(init, finalize, "c")
	m1 := newInitializerMeta(i1)
	m2 := newInitializerMeta(i2)
	m3 := newInitializerMeta(i3)

	pi := NewParallelInitializer(
		WithParallelInitializerContainer(container),
		WithParallelInitializerLogger(log.NewNilLogger()),
	)

	// Register things
	pi.RegisterInitializer(i1, WithInitializerName("a"))
	pi.RegisterInitializer(i2, WithInitializerName("b"))
	pi.RegisterInitializer(i3, WithInitializerName("c"))

	i1.finalizeErr = fmt.Errorf("oops x")
	i2.finalizeErr = fmt.Errorf("oops y")
	i3.finalizeErr = fmt.Errorf("oops z")

	WithInitializerName("a")(m1)
	WithInitializerName("b")(m2)
	WithInitializerName("c")(m3)

	err := pi.Finalize(context.Background())
	require.NotNil(t, err)

	expected := []errMeta{
		{err: fmt.Errorf("a returned error from finalize (oops x)"), source: m1},
		{err: fmt.Errorf("b returned error from finalize (oops y)"), source: m2},
		{err: fmt.Errorf("c returned error from finalize (oops z)"), source: m3},
	}
	for _, value := range expected {
		assert.Contains(t, err, value)
	}

	eventually(t, stringChanReceivesUnordered(finalize, "a", "b", "c"))
}
