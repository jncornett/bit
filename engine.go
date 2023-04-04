package bit

import (
	"context"
	"image"
	"sync/atomic"
	"time"

	"github.com/jncornett/doublebuf"
)

type Engine[RenderState any] struct {
	Clock   func() (ticks <-chan Tick, stop func())
	Steps   func() func(Tick) RenderState
	Render  func(context.Context, RenderState, WriteBuffer) error
	DrawUI  func(ReadBuffer, *uint64) error
	Size    image.Point
	Metrics Metrics
}

func (e *Engine[RenderState]) Run(ctx context.Context) error {
	db := doublebuf.New(
		image.NewNRGBA(image.Rectangle{Max: e.Size}),
		image.NewNRGBA(image.Rectangle{Max: e.Size}),
	)
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	e.Metrics = Metrics{Start: time.Now()}
	defer func() { e.Metrics.Stop = time.Now() }()
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
				updateStart := time.Now()
				renderState := step(t)
				atomic.AddInt64((*int64)(&e.Metrics.TotalTimeInUpdate), int64(time.Since(updateStart)))
				renderStart := time.Now()
				e.Render(ctx, renderState, db)
				atomic.AddInt64((*int64)(&e.Metrics.TotalTimeInRender), int64(time.Since(renderStart)))
				atomic.AddUint64(&e.Metrics.LoopCount, 1)
			}
		}
	}()
	err := e.DrawUI(db, &e.Metrics.DrawCount)
	cancel()
	<-done
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
func (t Tick) Zero() time.Time         { return t[0] }
func (t Tick) Age() time.Duration      { return t[2].Sub(t[0]) }

type WriteBuffer interface {
	Back(context.Context) (**image.NRGBA, error)
	TryBack() (**image.NRGBA, bool)
	Ready()
}

type ReadBuffer interface {
	Front() *image.NRGBA
	Next() (img *image.NRGBA, changed bool)
}

type FPS float64

func (fps FPS) Duration() time.Duration {
	return time.Duration(float64(time.Second) / float64(fps))
}

type Metrics struct {
	Start             time.Time
	LoopCount         uint64
	DrawCount         uint64
	TotalTimeInUpdate time.Duration
	TotalTimeInRender time.Duration
	Stop              time.Time
}

func (m *Metrics) Load() Metrics {
	// Atomically load metrics field by field.
	return Metrics{
		Start:             m.Start,
		Stop:              m.Stop,
		LoopCount:         atomic.LoadUint64(&m.LoopCount),
		DrawCount:         atomic.LoadUint64(&m.DrawCount),
		TotalTimeInUpdate: time.Duration(atomic.LoadInt64((*int64)(&m.TotalTimeInUpdate))),
		TotalTimeInRender: time.Duration(atomic.LoadInt64((*int64)(&m.TotalTimeInRender))),
	}
}

func (m Metrics) Loop() Metric {
	return Metric{m.LoopCount, m.totalDuration()}
}

func (m Metrics) Update() Metric {
	return Metric{m.LoopCount, m.TotalTimeInUpdate}
}

func (m Metrics) Render() Metric {
	return Metric{m.LoopCount, m.TotalTimeInRender}
}

func (m Metrics) Draw() Metric {
	return Metric{m.DrawCount, m.totalDuration()}
}

func (m Metrics) totalDuration() time.Duration {
	if m.Stop.IsZero() {
		return time.Since(m.Start)
	}
	return m.Stop.Sub(m.Start)
}

type Metric struct {
	Count     uint64
	TotalTime time.Duration
}

func (m Metric) AverageDuration() time.Duration {
	if m.Count == 0 {
		return 0
	}
	return time.Duration(float64(m.TotalTime) / float64(m.Count))
}

func (m Metric) AverageFPS() float64 {
	if m.TotalTime == 0 {
		return 0
	}
	return float64(m.Count) / m.TotalTime.Seconds()
}
