package process

import (
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthReasons(t *testing.T) {
	now := time.Now()
	clock := glock.NewMockClock()
	clock.SetCurrent(now)

	health := NewHealth(WithHealthClock(clock))
	assert.Equal(t, time.Duration(0), health.LastChange())

	clock.Advance(time.Hour)
	require.Nil(t, health.AddReason("foo"))
	assert.True(t, health.HasReason("foo"))
	clock.Advance(time.Minute)
	require.Nil(t, health.AddReason("bar"))
	assert.True(t, health.HasReason("foo"))
	assert.True(t, health.HasReason("bar"))
	clock.Advance(time.Minute)
	require.Nil(t, health.AddReason("baz"))
	assert.True(t, health.HasReason("foo"))
	assert.True(t, health.HasReason("bar"))
	assert.True(t, health.HasReason("baz"))
	assert.Nil(t, health.RemoveReason("bar"))
	assert.False(t, health.HasReason("bar"))

	reasons := health.Reasons()
	sort.Slice(reasons, func(i, j int) bool {
		keyi := reasons[i].Key.(string)
		keyj := reasons[j].Key.(string)

		return strings.Compare(keyi, keyj) < 0
	})

	expected := []Reason{
		{Key: "baz", Added: now.Add(time.Hour + time.Minute*2)},
		{Key: "foo", Added: now.Add(time.Hour)},
	}
	assert.Equal(t, expected, reasons)
}

func TestHealthLastChangedTime(t *testing.T) {
	clock := glock.NewMockClock()
	clock.SetCurrent(time.Now())

	health := NewHealth(WithHealthClock(clock))
	assert.Equal(t, time.Duration(0), health.LastChange())

	// Changed
	clock.Advance(time.Hour)
	require.Nil(t, health.AddReason("foo"))
	assert.True(t, health.HasReason("foo"))

	clock.Advance(time.Minute * 2)
	assert.Equal(t, time.Minute*2, health.LastChange())

	// No change
	require.Nil(t, health.AddReason("bar"))
	assert.True(t, health.HasReason("bar"))
	clock.Advance(time.Minute * 4)
	assert.Equal(t, time.Minute*6, health.LastChange())

	// No change
	require.Nil(t, health.RemoveReason("foo"))
	assert.False(t, health.HasReason("foo"))
	clock.Advance(time.Minute * 2)
	assert.Equal(t, time.Minute*8, health.LastChange())

	// Changed
	require.Nil(t, health.RemoveReason("bar"))
	assert.False(t, health.HasReason("bar"))
	assert.Equal(t, time.Minute*0, health.LastChange())
}

func TestHealthAddReasonError(t *testing.T) {
	health := NewHealth()
	health.AddReason("foo")
	assert.EqualError(t, health.AddReason("foo"), "reason foo already registered")
}

func TestHealthRemoveReasonError(t *testing.T) {
	health := NewHealth()
	assert.EqualError(t, health.RemoveReason("foo"), "reason foo not registered")
}
