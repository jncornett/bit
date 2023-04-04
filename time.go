package bit

import (
	"context"
	"time"
)

type Tick [3]time.Time

func NewTick(t time.Time) Tick { return Tick{t, t, t} }

func (t Tick) Step(now time.Time) Tick { return Tick{t[0], t[2], now} }
func (t Tick) Delta() time.Duration    { return t[2].Sub(t[1]) }
func (t Tick) Zero() time.Time         { return t[0] }
func (t Tick) Age() time.Duration      { return t[2].Sub(t[0]) }

func MakeClock(fps FPS) (start func() (ticks <-chan Tick, stop func())) {
	return func() (ticks <-chan Tick, stop func()) {
		ctx, cancel := context.WithCancel(context.Background())
		out := make(chan Tick)
		go func() {
			defer close(out)
			ticker := time.NewTicker(fps.Duration())
			defer ticker.Stop()
			var tick Tick
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				tick = NewTick(t)
			}
			for {
				select {
				case <-ctx.Done():
					return
				case t := <-ticker.C:
					next := tick.Step(t)
					select {
					case <-ctx.Done():
						return
					case out <- next:
						tick = next
					}
				}
			}
		}()
		return out, cancel
	}
}
