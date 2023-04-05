package bit

import (
	"image"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

type App struct {
	Dimensions image.Point
	Framerate  FPS
	Update     UpdateHandler
}

type AppOption func(*App)

func Update(f UpdateHandler) AppOption {
	return func(a *App) { a.Update = f }
}

const DefaultFramerate FPS = 60

func Framerate(fps FPS) AppOption {
	return func(a *App) { a.Framerate = fps }
}

func DefaultDimensions() image.Point {
	return image.Pt(800, 600)
}

func Dimensions(width, height int) AppOption {
	return func(a *App) { a.Dimensions = image.Pt(width, height) }
}

func NewApp(opts ...AppOption) *App {
	app := &App{
		Dimensions: DefaultDimensions(),
		Framerate:  DefaultFramerate,
	}
	return app.Option(opts...)
}

func (app *App) Option(opts ...AppOption) *App {
	for _, opt := range opts {
		opt(app)
	}
	return app
}

func (a *App) Main() {
	windowOptions := []app.Option{
		resizeApp(a.Dimensions),
	}
	go func() {
		err := func() error {
			widgets, stop := runEngine(
				a.Framerate,
				nil, // events
				a.Update,
				newRenderer(a.Dimensions),
			)
			defer stop()
			widget := <-widgets // get the initial widget
			prevWidget := dummyWidget
			w := app.NewWindow(windowOptions...)
			var ops op.Ops
			for {
				select {
				case w := <-widgets:
					prevWidget = widget
					widget = w
				case e := <-w.Events():
					switch e := e.(type) {
					case system.DestroyEvent:
						return e.Err
					case system.FrameEvent:
						gtx := layout.NewContext(&ops, e)
						prevWidget.Dispose()
						_ = widget.Func(gtx)
						op.InvalidateOp{}.Add(gtx.Ops)
						e.Frame(gtx.Ops)
					}
				}
			}
		}()
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func resizeApp(dimensions image.Point) app.Option {
	return func(m unit.Metric, c *app.Config) {
		app.Size(m.PxToDp(dimensions.X), m.PxToDp(dimensions.Y))(m, c)
	}
}
