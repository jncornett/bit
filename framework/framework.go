package framework

import (
	"context"
	"errors"
	"image"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"golang.org/x/sync/errgroup"
)

type Framework struct {
	funcs  []func(context.Context) error
	values [][2]any
	canvas *Canvas
}

func NewFramework(width, height int) *Framework {
	return &Framework{canvas: NewCanvas(width, height)}
}

func (fw *Framework) Value(key, value any) *Framework {
	fw.values = append(fw.values, [2]any{key, value})
	return fw
}

func (fw *Framework) Go(f func(context.Context) error) *Framework {
	fw.funcs = append(fw.funcs, f)
	return fw
}

func (fw *Framework) Render(fps float64, f func(context.Context, <-chan time.Time, image.Rectangle, func(func(*image.NRGBA)))) *Framework {
	return fw.Go(func(ctx context.Context) error {
		ticker := time.NewTicker(time.Duration(float64(time.Second) / fps))
		defer ticker.Stop()
		f(ctx, ticker.C, fw.canvas.bounds, fw.canvas.WriteBuffer)
		return nil
	})
}

func (fw *Framework) Main() {
	ctx, cancel := context.WithCancel(context.Background())
	for _, kv := range fw.values {
		ctx = context.WithValue(ctx, kv[0], kv[1])
	}
	g, ctx := errgroup.WithContext(ctx)
	for _, f := range fw.funcs {
		f := f
		g.Go(func() error {
			if err := f(ctx); err != nil {
				if !errors.Is(err, context.Canceled) {
					return err
				}
			}
			return nil
		})
	}
	g.Go(func() error {
		defer cancel()
		return mainLoop(fw.canvas.NextReadBuffer)
	})
	go func() {
		if err := g.Wait(); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func mainLoop(nextReadBuffer func() *image.NRGBA) error {
	w := app.NewWindow()
	var ops op.Ops
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			img := nextReadBuffer()
			paint.NewImageOp(img).Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
			op.InvalidateOp{}.Add(gtx.Ops)
			e.Frame(gtx.Ops)
		}
	}
}
