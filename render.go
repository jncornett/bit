package bit

import (
	"image"
	"image/draw"
	"sync"

	"gioui.org/layout"
	"gioui.org/op/paint"
)

type renderState struct {
	sprites []Sprite
	index   map[any]int
}

var (
	_ RenderSource = &renderState{}
	_ RenderTarget = &renderState{}
)

func newRenderState() *renderState {
	return &renderState{}
}

func (rs *renderState) AddSprite(id any, img image.Image, rect image.Rectangle) {
	sprite := Sprite{
		ID:        id,
		Image:     img,
		Rectangle: rect,
	}
	if i, ok := rs.index[id]; ok {
		rs.sprites[i] = sprite
	} else {
		rs.index[id] = len(rs.sprites)
		rs.sprites = append(rs.sprites, sprite)
	}
}

func (rs *renderState) Sprites() []Sprite {
	return rs.sprites
}

type imageBufferPool struct {
	dimensions image.Point
	pool       *sync.Pool
}

func newImageBufferPool(dimensions image.Point) *imageBufferPool {
	return &imageBufferPool{
		dimensions: dimensions,
		pool: &sync.Pool{
			New: newImageBuffer(dimensions),
		},
	}
}

func newImageBuffer(dimensions image.Point) func() any {
	return func() any {
		return image.NewNRGBA(image.Rect(0, 0, dimensions.X, dimensions.Y))
	}
}

func (pool *imageBufferPool) get() *image.NRGBA {
	return pool.pool.Get().(*image.NRGBA)
}

func (pool *imageBufferPool) put(img *image.NRGBA) {
	if img.Bounds().Max != pool.dimensions {
		return
	}
	pool.pool.Put(img)
}

func newRenderer(dimensions image.Point) RenderHandler {
	pool := newImageBufferPool(dimensions)
	return func(rs RenderSource) Widget {
		img := pool.get()
		draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)
		for _, sprite := range rs.Sprites() {
			draw.Draw(img, sprite.Rectangle, sprite.Image, image.Point{}, draw.Over)
		}
		return Widget{
			Func: func(gtx layout.Context) layout.Dimensions {
				paint.NewImageOp(img).Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
				return layout.Dimensions{Size: pool.dimensions}
			},
			Dispose: func() { pool.put(img) },
		}
	}
}

var dummyWidget = Widget{
	Func:    func(layout.Context) layout.Dimensions { return layout.Dimensions{} },
	Dispose: func() {},
}
