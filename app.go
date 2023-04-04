package bit

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"sync/atomic"
	"time"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
)

type App struct {
	FPS          FPS
	Size         image.Point
	Steps        func() func(Tick) RenderState
	DebugEnabled bool
}

func NewApp(size image.Point, steps func() func(Tick) RenderState) *App {
	return &App{Steps: steps, FPS: 60, Size: size}
}

func (a *App) Debug() *App {
	a.DebugEnabled = true
	return a
}

func (a *App) Main() {
	e := Engine[RenderState]{
		Clock:  DefineClock(a.FPS),
		Steps:  a.Steps,
		Render: defaultRender,
		DrawUI: mainLoop(a.Size),
		Size:   a.Size,
	}
	if a.DebugEnabled {
		log.SetHandler(cli.Default)
		log.SetLevel(log.DebugLevel)
		go func() {
			for range time.Tick(time.Second) {
				m := e.Metrics.Load()
				log.
					WithField("averageLoopFps", m.Loop().AverageFPS()).
					WithField("averageLoopDuration", m.Loop().AverageDuration()).
					WithField("loopCount", m.LoopCount).
					WithField("averageUpdateDuration", m.Update().AverageDuration()).
					WithField("averageRenderDuration", m.Render().AverageDuration()).
					WithField("averageDrawFps", m.Draw().AverageFPS()).
					WithField("averageDrawDuration", m.Draw().AverageDuration()).
					WithField("drawCount", m.DrawCount).
					Info("metric")
			}
		}()
	}
	go func() {
		if err := e.Run(context.Background()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

func defaultRender(ctx context.Context, state RenderState, buf WriteBuffer) error {
	img, ok := buf.TryBack()
	if !ok {
		return nil
	}
	defer buf.Ready()
	draw.Draw(*img, (*img).Bounds(), image.Black, image.Point{}, draw.Src)
	for _, rect := range state {
		draw.Draw(*img, rect, image.NewUniform(color.NRGBA{R: 0xff, A: 0xff}), image.Point{}, draw.Src)
	}
	return nil
}

func mainLoop(size image.Point) func(buf ReadBuffer, count *uint64) error {
	return func(buf ReadBuffer, count *uint64) error {
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
			atomic.AddUint64(count, 1)
		}
	}
}

// RenderState is whatever we send to the render function to draw on the image buffer.
type RenderState []image.Rectangle
