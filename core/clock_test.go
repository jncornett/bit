package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTick(t *testing.T) {
	var tick Tick
	assert.True(t, tick.IsZero())

	tick = NewTick(mustParse("2020-01-01T00:00:00Z"))
	assert.False(t, tick.IsZero())
	assert.Equal(t, mustParse("2020-01-01T00:00:00Z"), tick.Now())
	assert.Equal(t, Tick{
		Start: mustParse("2020-01-01T00:00:00Z"),
	}, tick)

	tick = tick.Step(mustParse("2020-01-01T00:00:01Z"))
	assert.Equal(t, mustParse("2020-01-01T00:00:01Z"), tick.Now())
	assert.Equal(t, Tick{
		Start:   mustParse("2020-01-01T00:00:00Z"),
		Delta:   time.Second,
		Elapsed: time.Second,
		N:       1,
	}, tick)

	tick = tick.Step(mustParse("2020-01-01T00:00:03Z"))
	assert.Equal(t, mustParse("2020-01-01T00:00:03Z"), tick.Now())
	assert.Equal(t, Tick{
		Start:   mustParse("2020-01-01T00:00:00Z"),
		Delta:   2 * time.Second,
		Elapsed: 3 * time.Second,
		N:       2,
	}, tick)
}

func TestStartClock(t *testing.T) {
	t.Run("sends ticks", func(t *testing.T) {
		clock, stop := StartClock(time.Millisecond)
		defer stop()
		tick := <-clock
		assert.False(t, tick.IsZero())
		assert.Equal(t, uint64(1), tick.N)
		assert.NotZero(t, tick.Elapsed)

		tick = <-clock
		assert.Equal(t, uint64(2), tick.N)
	})
	t.Run("doesn't send ticks after stop", func(t *testing.T) {
		clock, stop := StartClock(time.Millisecond)
		<-clock
		stop()
		_, ok := <-clock
		assert.False(t, ok)
	})
	t.Run("sends a tick with an updated delta", func(t *testing.T) {
		clock, stop := StartClock(time.Millisecond)
		defer stop()
		<-clock
		time.Sleep(3 * time.Millisecond)
		tick := <-clock
		assert.Less(t, int64(2*time.Millisecond), int64(tick.Delta))
	})
}

func TestNewFrameClock(t *testing.T) {
	clock := make(chan Tick, 1)
	events := make(chan Event, 2)
	fck := newFrameClock(clock, events)
	tick := NewTick(mustParse("2020-01-01T00:00:00Z"))
	clock <- tick
	events <- Event{}
	events <- Event{}
	got := <-fck
	assert.Equal(t, Frame{Tick: tick, Events: []Event{{}, {}}}, got)
}

func mustParse(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
