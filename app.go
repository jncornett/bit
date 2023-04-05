package bit

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
)

type App[GameState any] struct {
	FPS              FPS
	Size             image.Point
	InitialGameState GameState
	Update           func(Tick, GameState) (GameState, RenderState)
	DebugEnabled     bool
}

func NewApp[GameState any](size image.Point, initialGameState GameState, update func(Tick, GameState) (GameState, RenderState)) *App[GameState] {
	return &App[GameState]{
		FPS:              60,
		Size:             size,
		InitialGameState: initialGameState,
		Update:           update,
	}
}

func (a *App[GameState]) Debug() *App[GameState] {
	a.DebugEnabled = true
	return a
}

func (a *App[GameState]) Main() {
	e := Engine[GameState, RenderState]{
		Update:    a.Update,
		Render:    defaultRender,
		StartDraw: mainLoop(a.Size),
		Size:      a.Size,
		Metrics:   MakeEngineMetrics(time.Now()),
	}
	if a.DebugEnabled {
		log.SetHandler(cli.Default)
		log.SetLevel(log.DebugLevel)
		go func() {
			for range time.Tick(time.Second) {
				m := e.Metrics.Load()
				log.
					WithField("loop", m.Loop().String()).
					WithField("update", m.Update.String()).
					WithField("render", m.Render.String()).
					Info("metric")
			}
		}()
	}
	clock, stop := NewClock(a.FPS)
	go func() {
		defer stop()
		if err := e.Run(clock, a.InitialGameState); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

func defaultRender(state RenderState, img *image.NRGBA) {
	draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)
	for _, rect := range state {
		draw.Draw(img, rect, image.NewUniform(color.NRGBA{R: 0xff, A: 0xff}), image.Point{}, draw.Src)
	}
}

func mainLoop(size image.Point) func(buf ReadBuffer) error {
	return func(buf ReadBuffer) error {
		w := app.NewWindow()
		var ops op.Ops
		var resized bool
		for {
			e := <-w.Events()
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				if !resized {
					resized = true
					w.Option(app.Size(
						gtx.Metric.PxToDp(size.X),
						gtx.Metric.PxToDp(size.Y),
					))
				} else {
					img, _ := buf.Next()
					paint.NewImageOp(img).Add(gtx.Ops)
					paint.PaintOp{}.Add(gtx.Ops)
				}
				op.InvalidateOp{}.Add(gtx.Ops)
				e.Frame(gtx.Ops)
			}
		}
	}
}

// RenderState is whatever we send to the render function to draw on the image buffer.
type RenderState []image.Rectangle
