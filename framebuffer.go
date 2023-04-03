package bit

import (
	"context"
	"image"
)

type Framebuffer struct {
	front  *image.NRGBA
	next   chan *image.NRGBA
	back   chan *image.NRGBA
	bounds image.Rectangle
}

func NewFrameBuffer(bounds image.Rectangle) *Framebuffer {
	fb := &Framebuffer{bounds: bounds}
	fb.next = make(chan *image.NRGBA, 1)
	fb.back = make(chan *image.NRGBA, 1)
	fb.back <- image.NewNRGBA(bounds)
	fb.front = image.NewNRGBA(bounds)
	return fb
}

func (fr *Framebuffer) Write(ctx context.Context, f func(*image.NRGBA)) error {
	var buf *image.NRGBA
	select {
	case <-ctx.Done():
		return ctx.Err()
	case buf = <-fr.back:
	}
	f(buf)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case fr.next <- buf:
	}
	return nil
}

func (fr *Framebuffer) Read() (img *image.NRGBA, changed bool) {
	select {
	case buf := <-fr.next:
		old := fr.front
		fr.front = buf
		fr.back <- old
		return fr.front, true
	default:
		return fr.front, false
	}
}
