package core

import "time"

func StartClock(d time.Duration) (ck Clock, stop func()) {
	out := make(chan Tick)
	done := make(chan struct{})
	go func() {
		defer close(out)
		ticker := time.NewTicker(d)
		defer ticker.Stop()
		var prev Tick
		select {
		case <-done:
			return
		case t := <-ticker.C:
			prev = NewTick(t)
		}
		for {
			var next Tick
			select {
			case <-done:
				return
			case t := <-ticker.C:
				next = prev.Step(t)
			}
		Inner:
			for {
				select {
				case <-done:
					return
				case out <- next:
					// successfully sent the next tick
					prev = next
					break Inner
				case t := <-ticker.C:
					// failed to deliver the next tick, drop it
					// and try again with an updated time
					next = prev.Step(t)
				}
			}
		}
	}()
	return out, func() { close(done); <-out }
}

func BindEventsToClock(ck <-chan Tick, events <-chan Event) <-chan Frame {
	out := make(chan Frame)
	go func() {
		defer close(out)
		bufs := NewBuffers([]Event{}, []Event{})
		for {
			var tick Tick
			for tick.IsZero() {
				select {
				case e, ok := <-events:
					if !ok {
						events = nil
					} else {
						*bufs.Current() = append(*bufs.Current(), e)
					}
				case t, ok := <-ck:
					if !ok {
						return
					}
					tick = t
				}
			}
			for !tick.IsZero() {
				select {
				case e, ok := <-events:
					if !ok {
						events = nil
					} else {
						*bufs.Current() = append(*bufs.Current(), e)
					}
				case t, ok := <-ck:
					if !ok {
						return
					}
					tick = t // try again with a new tick
				case out <- Frame{tick, *bufs.Current()}:
					bufs.Swap()
					*bufs.Current() = (*bufs.Current())[:0]
					tick = Tick{}
				}
			}
		}
	}()
	return out
}

func CreateUpdateLoop[R any](frames <-chan Frame, update func(Frame) R) <-chan R {
	return MapChannel(frames, update)
}

func CreateRenderLoop[R, D any](updates <-chan R, render func(R, *D), bufs Buffers[D]) <-chan D {
	out := make(chan D)
	go func() {
		defer close(out)
		for toRender := range updates {
			var sent bool
			for !sent {
				render(toRender, bufs.Current())
				select {
				case r, ok := <-updates:
					if !ok {
						return
					}
					toRender = r // try again with a new update
				case out <- *bufs.Current():
					sent = true
					bufs.Swap()
				}
			}
		}
	}()
	return out
}

func MapChannel[T, R any](in <-chan T, f func(T) R) <-chan R {
	out := make(chan R)
	go func() {
		defer close(out)
		for t := range in {
			out <- f(t)
		}
	}()
	return out
}
