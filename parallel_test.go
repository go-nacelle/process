package process

import (
	"context"
	"testing"
)

func TestToStreamErrorFuncNil(t *testing.T) {
	trace := make(chan string, 4)
	fn := toStreamErrorFunc(newTracedSingleErrorFunc(trace, traceValue("a"), nil))

	errs := fn(context.Background())
	assertChannelContents(t, readErrorChannel(errs), nil)

	close(trace)
	assertChannelContents(t, readStringChannel(trace), seq("a"))
}

func TestToStreamErrorFuncNonNil(t *testing.T) {
	trace := make(chan string, 4)
	fn := toStreamErrorFunc(newTracedSingleErrorFunc(trace, traceValue("a"), testErr1))

	errs := fn(context.Background())
	assertChannelContents(t, readErrorChannel(errs), seq(testErr1))

	close(trace)
	assertChannelContents(t, readStringChannel(trace), seq("a"))
}

func TestChain(t *testing.T) {
	trace := make(chan string, 4)

	fn := chain(
		toStreamErrorFunc(newTracedSingleErrorFunc(trace, traceValue("a"), nil)),
		toStreamErrorFunc(newTracedSingleErrorFunc(trace, traceValue("b"), testErr1)),
		toStreamErrorFunc(newTracedSingleErrorFunc(trace, traceValue("c"), testErr2)),
		toStreamErrorFunc(newTracedSingleErrorFunc(trace, traceValue("d"), nil)),
	)

	errs := fn(context.Background())
	assertChannelContents(t, readErrorChannel(errs), seq(testErr1))

	close(trace)
	assertChannelContents(t, readStringChannel(trace), seq("a", "b"))
}

func TestSequence(t *testing.T) {
	trace := make(chan string, 4)

	fn := sequence(
		toStreamErrorFunc(newTracedSingleErrorFunc(trace, traceValue("a"), nil)),
		toStreamErrorFunc(newTracedSingleErrorFunc(trace, traceValue("b"), testErr1)),
		toStreamErrorFunc(newTracedSingleErrorFunc(trace, traceValue("c"), testErr2)),
		toStreamErrorFunc(newTracedSingleErrorFunc(trace, traceValue("d"), nil)),
	)

	errs := fn(context.Background())
	assertChannelContents(t, readErrorChannel(errs), seq(testErr1, testErr2))

	close(trace)
	assertChannelContents(t, readStringChannel(trace), seq("a", "b", "c", "d"))
}

func TestParallel(t *testing.T) {
	t1 := make(chan string)
	t2 := make(chan string)
	t3 := make(chan string)
	t4 := make(chan string)
	trace := make(chan string)
	traceBuffered := make(chan string, 8)

	fn := parallel(
		toStreamErrorFunc(newTracedSingleErrorFunc(trace, t1, nil)),
		toStreamErrorFunc(newTracedSingleErrorFunc(trace, t2, testErr1)),
		toStreamErrorFunc(newTracedSingleErrorFunc(trace, t3, testErr2)),
		toStreamErrorFunc(newTracedSingleErrorFunc(trace, t4, nil)),
	)

	go func() {
		type pair struct {
			ch    chan<- string
			value string
		}

		for _, pair := range []pair{
			{t4, "d"},
			{t3, "c"},
			{t1, "a"},
			{t2, "b"},
			{t4, "D"},
			{t3, "C"},
			{t1, "A"},
			{t2, "B"},
		} {
			pair.ch <- pair.value
			traceBuffered <- <-trace
		}

		for _, ch := range []chan string{t1, t2, t3, t4} {
			close(ch)
		}
	}()

	errs := fn(context.Background())
	assertChannelContents(t, readErrorChannel(errs), seq(unordered(testErr1, testErr2)))

	close(trace)
	close(traceBuffered)
	assertChannelContents(t, readStringChannel(traceBuffered), seq("d", "c", "a", "b", "D", "C", "A", "B"))
}
