package process

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/derision-test/glock"
	mockassert "github.com/derision-test/go-mockgen/testutil/assert"
	"github.com/stretchr/testify/assert"
)

func TestMetaInit(t *testing.T) {
	wrapped := NewMockMaximumProcess()
	meta := newMeta(wrapped)

	assert.Nil(t, meta.Init(context.Background()))
	mockassert.CalledOnce(t, wrapped.InitFunc)
}

func TestMetaInitError(t *testing.T) {
	wrapped := NewMockMaximumProcess()
	wrapped.InitFunc.SetDefaultReturn(testErr1)
	meta := newMeta(wrapped, WithMetaName("test-service"))

	assert.EqualError(t, meta.Init(context.Background()), "test-service: init failed (oops1)")
	mockassert.CalledOnce(t, wrapped.InitFunc)
}

func TestMetaInitTimeout(t *testing.T) {
	clock := glock.NewMockClock()
	wrapped := NewMockMaximumProcess()
	initHook, _ := newBlockingSingleErrorFunc()
	wrapped.InitFunc.SetDefaultHook(initHook)
	meta := newMeta(wrapped, WithMetaName("test-service"), WithMetaInitTimeout(time.Second*5), withMetaInitClock(clock))

	results := runAsync(context.Background(), meta.Init)
	clock.BlockingAdvance(time.Second * 5)
	assertChannelContents(t, readErrorChannel(results), seq(errors.New("test-service: init timeout")))
}

func TestMetaStart(t *testing.T) {
	wrapped := NewMockMaximumProcess()
	startHook, started := newSingalingSingleErrorFunc()
	wrapped.StartFunc.SetDefaultHook(startHook)
	meta := newMeta(wrapped, WithMetaName("test-service"))

	assert.Nil(t, meta.Init(context.Background()))
	results := runAsync(context.Background(), meta.Start)

	<-started
	assert.Nil(t, meta.Stop(context.Background()))
	assertChannelContents(t, readErrorChannel(results), seq(nil))
	mockassert.CalledOnce(t, wrapped.StartFunc)
	mockassert.CalledOnce(t, wrapped.StopFunc)
}

func TestMetaStartupTimeout(t *testing.T) {
	health := NewHealth()
	healthComponent, _ := health.Register("test")
	healthComponent.Update(false)

	clock := glock.NewMockClock()
	wrapped := NewMockMaximumProcess()
	startHook, started := newSingalingSingleErrorFunc()
	wrapped.StartFunc.SetDefaultHook(startHook)
	meta := newMeta(wrapped, WithMetaName("test-service"), WithMetaHealth(health), WithMetaHealthKey("test"), WithMetaStartupTimeout(time.Second*5), withMetaStartupClock(clock))

	assert.Nil(t, meta.Init(context.Background()))
	results := runAsync(context.Background(), meta.Start)

	<-started
	clock.BlockingAdvance(time.Second)
	clock.BlockingAdvance(time.Second)
	clock.BlockingAdvance(time.Second)
	clock.BlockingAdvance(time.Second)
	clock.BlockingAdvance(time.Second)

	assertChannelContents(t, readErrorChannel(results), seq(ErrStartupTimeout))
}

func TestMetaStartTwice(t *testing.T) {
	wrapped := NewMockMaximumProcess()
	startHook, started := newSingalingSingleErrorFunc()
	wrapped.StartFunc.SetDefaultHook(startHook)
	meta := newMeta(wrapped, WithMetaName("test-service"))

	assert.Nil(t, meta.Init(context.Background()))
	results := runAsync(context.Background(), meta.Start)

	<-started
	assert.Nil(t, meta.Stop(context.Background()))
	assertChannelContents(t, readErrorChannel(results), seq(nil))
	mockassert.CalledOnce(t, wrapped.StartFunc)

	results = runAsync(context.Background(), meta.Start)
	assertChannelContents(t, readErrorChannel(results), seq(nil))
	mockassert.CalledOnce(t, wrapped.StartFunc)
}

func TestMetaStopTimeout(t *testing.T) {
	clock := glock.NewMockClock()
	wrapped := NewMockMaximumProcess()
	startHook, started := newSingalingSingleErrorFunc()
	wrapped.StartFunc.SetDefaultHook(startHook)
	stopHook, _ := newBlockingSingleErrorFunc()
	wrapped.StopFunc.SetDefaultHook(stopHook)
	meta := newMeta(wrapped, WithMetaName("test-service"), WithMetaStopTimeout(time.Second*5), withMetaStopClock(clock))

	assert.Nil(t, meta.Init(context.Background()))
	startResults := runAsync(context.Background(), meta.Start)

	<-started
	stopResults := runAsync(context.Background(), meta.Stop)

	clock.BlockingAdvance(time.Second * 5)
	assertChannelContents(t, readErrorChannel(startResults), seq(nil))
	assertChannelContents(t, readErrorChannel(stopResults), seq(errors.New("test-service: stop timeout")))
}

func TestMetaShutdownTimeout(t *testing.T) {
	clock := glock.NewMockClock()
	wrapped := NewMockMaximumProcess()
	startHook, started := newBlockingSingleErrorFunc()
	wrapped.StartFunc.SetDefaultHook(startHook)
	meta := newMeta(wrapped, WithMetaName("test-service"), WithMetaShutdownTimeout(time.Second*5), withMetaShutdownClock(clock))

	assert.Nil(t, meta.Init(context.Background()))
	results := runAsync(context.Background(), meta.Start)

	<-started
	assert.Nil(t, meta.Stop(context.Background()))

	clock.BlockingAdvance(time.Second * 5)
	assertChannelContents(t, readErrorChannel(results), seq(ErrShutdownTimeout))
}

func TestMetaStartError(t *testing.T) {
	wrapped := NewMockMaximumProcess()
	wrapped.StartFunc.SetDefaultReturn(testErr1)
	meta := newMeta(wrapped, WithMetaName("test-service"))

	assert.Nil(t, meta.Init(context.Background()))
	assert.EqualError(t, meta.Start(context.Background()), "test-service: start failed (oops1)")
	mockassert.CalledOnce(t, wrapped.StartFunc)
}

func TestMetaStopError(t *testing.T) {
	wrapped := NewMockMaximumProcess()
	startHook, scheduled := newSingalingSingleErrorFunc()
	wrapped.StartFunc.SetDefaultHook(startHook)
	wrapped.StopFunc.SetDefaultReturn(testErr1)
	meta := newMeta(wrapped, WithMetaName("test-service"))

	assert.Nil(t, meta.Init(context.Background()))
	results := runAsync(context.Background(), meta.Start)

	<-scheduled
	assert.EqualError(t, meta.Stop(context.Background()), "test-service: stop failed (oops1)")
	assertChannelContents(t, readErrorChannel(results), seq(nil))
	mockassert.CalledOnce(t, wrapped.StopFunc)
}

func TestMetaStartExitsWithEarlyExit(t *testing.T) {
	wrapped := NewMockMaximumProcess()
	meta := newMeta(wrapped, WithMetaName("test-service"), WithEarlyExit(true))

	assert.Nil(t, meta.Init(context.Background()))
	assert.Nil(t, meta.Start(context.Background()))
	mockassert.CalledOnce(t, wrapped.StartFunc)
}

func TestMetaStartExitsWithoutEarlyExit(t *testing.T) {
	wrapped := NewMockMaximumProcess()
	meta := newMeta(wrapped, WithMetaName("test-service"))

	assert.Nil(t, meta.Init(context.Background()))
	assert.Equal(t, ErrUnexpectedReturn, meta.Start(context.Background()))
	mockassert.CalledOnce(t, wrapped.StartFunc)
}

func TestMetaStartCalledUninitialized(t *testing.T) {
	wrapped := NewMockMaximumProcess()
	meta := newMeta(wrapped)

	assert.Nil(t, meta.Start(context.Background()))
	mockassert.NotCalled(t, wrapped.StartFunc)
}

func TestMetaStopCalledUninitialized(t *testing.T) {
	wrapped := NewMockMaximumProcess()
	meta := newMeta(wrapped)

	assert.Nil(t, meta.Stop(context.Background()))
	mockassert.NotCalled(t, wrapped.StopFunc)
}

func TestMetaStopCalledNotRunning(t *testing.T) {
	wrapped := NewMockMaximumProcess()
	meta := newMeta(wrapped)

	assert.Nil(t, meta.Init(context.Background()))
	assert.Nil(t, meta.Stop(context.Background()))
	mockassert.NotCalled(t, wrapped.StopFunc)
}

func TestMetaFinalize(t *testing.T) {
	wrapped := NewMockMaximumProcess()
	meta := newMeta(wrapped)

	assert.Nil(t, meta.Init(context.Background()))
	assert.Nil(t, meta.Finalize(context.Background()))
	mockassert.CalledOnce(t, wrapped.FinalizeFunc)
}

func TestMetaFinalizeError(t *testing.T) {
	wrapped := NewMockMaximumProcess()
	wrapped.FinalizeFunc.SetDefaultReturn(testErr1)
	meta := newMeta(wrapped, WithMetaName("test-service"))

	assert.Nil(t, meta.Init(context.Background()))
	assert.EqualError(t, meta.Finalize(context.Background()), "test-service: finalize failed (oops1)")
	mockassert.CalledOnce(t, wrapped.FinalizeFunc)
}

func TestMetaFinalizeTimeout(t *testing.T) {
	clock := glock.NewMockClock()
	wrapped := NewMockMaximumProcess()
	finalizeHook, _ := newBlockingSingleErrorFunc()
	wrapped.FinalizeFunc.SetDefaultHook(finalizeHook)
	meta := newMeta(wrapped, WithMetaName("test-service"), WithMetaFinalizeTimeout(time.Second*5), withMetaFinalizeClock(clock))

	assert.Nil(t, meta.Init(context.Background()))
	results := runAsync(context.Background(), meta.Finalize)
	clock.BlockingAdvance(time.Second * 5)
	assertChannelContents(t, readErrorChannel(results), seq(errors.New("test-service: finalize timeout")))
}

func TestMetaFinalizeCalledUninitialized(t *testing.T) {
	wrapped := NewMockMaximumProcess()
	meta := newMeta(wrapped)

	assert.Nil(t, meta.Finalize(context.Background()))
	mockassert.NotCalled(t, wrapped.FinalizeFunc)
}
