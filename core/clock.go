package core

import "time"

type Tick struct {
	Start          time.Time
	Delta, Elapsed time.Duration
	N              uint64
}

func NewTick(start time.Time) Tick { return Tick{Start: start} }

func (t Tick) IsZero() bool { return t.Start.IsZero() }

func (t Tick) Now() time.Time { return t.Start.Add(t.Elapsed) }

func (t Tick) Step(now time.Time) Tick {
	elapsed := now.Sub(t.Start)
	return Tick{
		Start:   t.Start,
		Delta:   elapsed - t.Elapsed,
		Elapsed: elapsed,
		N:       t.N + 1,
	}
}

func newFrameClock(ck Clock, events <-chan Event) <-chan Frame {
	out := make(chan Frame)
	go func() {
		defer close(out)
		bufs := NewBuffers([]Event{}, []Event{})
		var tick Tick
		for {
			for tick.IsZero() {
				select {
				case e, ok := <-events:
					if !ok {
						events = nil
						continue
					}
					*bufs.Current() = append(*bufs.Current(), e)
				case t, ok := <-ck:
					if !ok {
						return
					}
					tick = t
				}
			}
			select {
			case e, ok := <-events:
				if !ok {
					events = nil
					continue
				}
				*bufs.Current() = append(*bufs.Current(), e)
			case t, ok := <-ck:
				if !ok {
					return
				}
				tick = t
			case out <- Frame{Tick: tick, Events: *bufs.Current()}:
				tick = Tick{}
				bufs.Swap()
				*bufs.Current() = (*bufs.Current())[:0]
			}
		}
	}()
	return out
}
