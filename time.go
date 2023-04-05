package bit

import (
	"context"
	"time"

	"github.com/jncornett/chans"
)

type FPS float64

func (fps FPS) Duration() time.Duration {
	return time.Duration(float64(time.Second) / float64(fps))
}

type Tick [3]time.Time

func NewTick(t time.Time) Tick { return Tick{t, t, t} }

func (t Tick) Step(now time.Time) Tick { return Tick{t[0], t[2], now} }
func (t Tick) Delta() time.Duration    { return t[2].Sub(t[1]) }
func (t Tick) Zero() time.Time         { return t[0] }
func (t Tick) Age() time.Duration      { return t[2].Sub(t[0]) }

func MakeClock(fps FPS) (start func() (ticks chans.Chan[Tick], stop func())) {
	return func() (ch chans.Chan[Tick], stop func()) {
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
