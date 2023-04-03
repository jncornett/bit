package bit

import (
	"context"
	"fmt"
	"image"
	"time"
)

type Engine[RenderState any] struct {
	Clock  func() (ticks <-chan Tick, stop func())
	Steps  func() func(Tick) RenderState
	Render func(context.Context, RenderState, WriteBuffer) error
	DrawUI func(ReadBuffer) error
	Size   image.Point
}

func (e *Engine[RenderState]) Run(ctx context.Context) error {
	fb := NewFrameBuffer(image.Rectangle{Max: e.Size})
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticks, stop := e.Clock()
		defer stop()
		step := e.Steps()
		for {
			select {
			case <-ctx.Done():
				return
			case t, ok := <-ticks:
				if !ok {
					return
				}
				renderState := step(t)
				e.Render(ctx, renderState, fb)
			}
		}
	}()
	err := e.DrawUI(fb)
	cancel()
	fmt.Println("cancelled")
	<-done
	fmt.Println("done")
	return err
}

type Clock func() (ticks <-chan Tick, stop func())

func DefineClock(fps FPS) func() (ticks <-chan Tick, stop func()) {
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

type Tick [3]time.Time

func NewTick(t time.Time) Tick { return Tick{t, t, t} }

func (t Tick) Step(now time.Time) Tick { return Tick{t[0], t[2], now} }
func (t Tick) Delta() time.Duration    { return t[2].Sub(t[1]) }

type WriteBuffer interface {
	Write(context.Context, func(*image.NRGBA)) error
}

type ReadBuffer interface {
	Read() (img *image.NRGBA, changed bool)
}

type FPS float64

func (fps FPS) Duration() time.Duration {
	return time.Duration(float64(time.Second) / float64(fps))
}
