package bit

import (
	"context"
	"image"
	"sync/atomic"
	"time"

	"github.com/jncornett/doublebuf"
)

type Engine[GameState, RenderState any] struct {
	StartClock func() (ticks <-chan Tick, stop func())
	Update     func(Tick, GameState) (GameState, RenderState)
	Render     func(RenderState, *image.NRGBA)
	StartDraw  func(ReadBuffer) error
	Size       image.Point
	Metrics    EngineMetrics
}

func (e *Engine[GameState, RenderState]) Run(ctx context.Context, initialGameState GameState) error {
	db := doublebuf.New(
		image.NewNRGBA(image.Rectangle{Max: e.Size}),
		image.NewNRGBA(image.Rectangle{Max: e.Size}),
	)
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	defer func() { e.Metrics.Stop = time.Now() }()
	go func() {
		defer close(done)
		ticks, stop := e.StartClock()
		defer stop()
		gameState := initialGameState
		for {
			select {
			case <-ctx.Done():
				return
			case t, ok := <-ticks:
				if !ok {
					return
				}

				var renderState RenderState
				WithDurationMetric(&e.Metrics.Update, func() {
					gameState, renderState = e.Update(t, gameState)
				})

				if buf, ok := db.TryBack(); ok { // attempt to acquire the back buffer
					WithDurationMetric(&e.Metrics.Render, func() {
						e.Render(renderState, *buf)
						db.Ready()
					})
				}

				atomic.AddUint64(&e.Metrics.LoopCount, 1)
			}
		}
	}()
	err := e.StartDraw(db)
	cancel()
	<-done
	return err
}

type ReadBuffer interface {
	Front() *image.NRGBA
	Next() (img *image.NRGBA, changed bool)
}

type FPS float64

func (fps FPS) Duration() time.Duration {
	return time.Duration(float64(time.Second) / float64(fps))
}

type EngineMetrics struct {
	Start, Stop    time.Time
	LoopCount      uint64
	Update, Render DurationMetric
}

func MakeEngineMetrics(now time.Time) EngineMetrics {
	return EngineMetrics{Start: now}
}

func (m *EngineMetrics) Load() EngineMetrics {
	// Atomically load metrics field by field.
	return EngineMetrics{
		Start:     m.Start,
		Stop:      m.Stop,
		LoopCount: atomic.LoadUint64(&m.LoopCount),
		Update:    m.Update.Load(),
		Render:    m.Render.Load(),
	}
}

func (m EngineMetrics) LoopAt(now time.Time) DurationMetric {
	return DurationMetric{
		Count:         m.LoopCount,
		TotalDuration: now.Sub(m.Start),
	}
}

func (m EngineMetrics) Loop() DurationMetric {
	if m.Stop.IsZero() {
		return m.LoopAt(time.Now())
	}
	return m.LoopAt(m.Stop)
}
