package main

import (
	"image"
	"math"
	"math/rand"

	"github.com/jncornett/bit"
	"github.com/jncornett/bit/gfx"
)

func main() {
	bit.
		NewApp(image.Pt(800, 600), func() func(bit.Tick) bit.RenderState {
			renderState := make(bit.RenderState, 10000)
			positions := make([]gfx.Vec, len(renderState))
			velocities := make([]gfx.Vec, len(renderState))
			halfSize := 2.0
			for i := range renderState {
				center := gfx.Vec{X: float64(20 + rand.Intn(760)), Y: float64(20 + rand.Intn(560))}
				rect := gfx.Rect{
					Min: center.Sub(gfx.Vec{X: halfSize, Y: halfSize}),
					Max: center.Add(gfx.Vec{X: halfSize, Y: halfSize}),
				}
				positions[i] = center
				theta := rand.Float64() * 2 * math.Pi
				velocities[i] = gfx.Vec{X: math.Cos(theta), Y: math.Sin(theta)}
				renderState[i] = rect.Rectangle()
			}
			const speed = 250.0
			return func(t bit.Tick) bit.RenderState {
				for i, v := range velocities {
					p := positions[i]
					d := v.Mul(speed * t.Delta().Seconds())
					p = p.Add(d)
					positions[i] = p
					rect := gfx.Rect{
						Min: p.Sub(gfx.Vec{X: halfSize, Y: halfSize}),
						Max: p.Add(gfx.Vec{X: halfSize, Y: halfSize}),
					}
					abs := func(x float64) float64 {
						if x < 0 {
							return -x
						}
						return x
					}
					if rect.Min.X < 0 {
						v = gfx.V(abs(v.X), v.Y)
					}
					if rect.Min.Y < 0 {
						v = gfx.V(v.X, abs(v.Y))
					}
					if rect.Max.X > 800 {
						v = gfx.V(-abs(v.X), v.Y)
					}
					if rect.Max.Y > 600 {
						v = gfx.V(v.X, -abs(v.Y))
					}
					velocities[i] = v
					renderState[i] = rect.Rectangle()
				}
				return renderState
			}
		}).
		Debug().
		Main()
}
