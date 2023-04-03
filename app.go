package bit

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"

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
		DrawUI: mainLoop,
		Size:   a.Size,
	}
	if a.DebugEnabled {
		log.SetHandler(cli.Default)
		log.SetLevel(log.DebugLevel)
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

func defaultRender(ctx context.Context, state RenderState, fb WriteBuffer) error {
	return fb.Write(ctx, func(img *image.NRGBA) {
		draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)
		for _, rect := range state {
			draw.Draw(img, rect, image.NewUniform(color.NRGBA{R: 0xff, A: 0xff}), image.Point{}, draw.Src)
		}
	})
}

func mainLoop(fb ReadBuffer) error {
	w := app.NewWindow()
	var ops op.Ops
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			img, changed := fb.Read()
			if changed {
				paint.NewImageOp(img).Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
			}
			op.InvalidateOp{}.Add(gtx.Ops)
			e.Frame(gtx.Ops)
		}
	}

}

type RenderState []image.Rectangle
