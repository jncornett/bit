package bit

import (
	"image"

	"gioui.org/layout"
)

type UpdateHandler func(Frame, RenderTarget)

type RenderHandler func(RenderSource) Widget

type WidgetFunc func(layout.Context) layout.Dimensions

type Widget struct {
	Func    WidgetFunc
	Dispose func()
}

func (w Widget) Layout(gtx layout.Context) layout.Dimensions {
	return w.Func(gtx)
}

type RenderSource interface {
	Sprites() []Sprite
}

type RenderTarget interface {
	AddSprite(id any, img image.Image, rect image.Rectangle)
}

type Sprite struct {
	ID any
	image.Image
	image.Rectangle
}

type Frame struct {
	Tick   Tick
	Events []any
}

func MakeFrame(tick Tick, events []any) Frame {
	return Frame{Tick: tick, Events: events}
}
