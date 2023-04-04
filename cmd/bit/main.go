package main

import (
	"image"
	"math"
	"math/rand"

	"github.com/jncornett/bit"
	"github.com/jncornett/bit/gfx"
)

func main() {
	var window = image.Pt(1600, 1200)
	type gameState struct {
		renderState bit.RenderState
		positions   []gfx.Vec
		velocities  []gfx.Vec
	}
	var initialGameState = gameState{}
	halfSize := 2.0
	{
		initialGameState.renderState = make(bit.RenderState, 10000)
		initialGameState.positions = make([]gfx.Vec, len(initialGameState.renderState))
		initialGameState.velocities = make([]gfx.Vec, len(initialGameState.renderState))
		for i := range initialGameState.renderState {
			center := gfx.Vec{X: float64(20 + rand.Intn(window.X-40)), Y: float64(20 + rand.Intn(window.Y-40))}
			rect := gfx.Rect{
				Min: center.Sub(gfx.Vec{X: halfSize, Y: halfSize}),
				Max: center.Add(gfx.Vec{X: halfSize, Y: halfSize}),
			}
			initialGameState.positions[i] = center
			theta := rand.Float64() * 2 * math.Pi
			initialGameState.velocities[i] = gfx.Vec{X: math.Cos(theta), Y: math.Sin(theta)}
			initialGameState.renderState[i] = rect.Rectangle()
		}
	}
	const speed = 500.0
	bit.
		NewApp(window, &initialGameState, func(t bit.Tick, state *gameState) (*gameState, bit.RenderState) {
			for i, v := range state.velocities {
				p := state.positions[i]
				d := v.Mul(speed * t.Delta().Seconds())
				p = p.Add(d)
				state.positions[i] = p
				rect := gfx.Rect{
					Min: p.Round().Sub(gfx.Vec{X: halfSize, Y: halfSize}),
					Max: p.Round().Add(gfx.Vec{X: halfSize, Y: halfSize}),
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
				if rect.Max.X > float64(window.X) {
					v = gfx.V(-abs(v.X), v.Y)
				}
				if rect.Max.Y > float64(window.Y) {
					v = gfx.V(v.X, -abs(v.Y))
				}
				state.velocities[i] = v
				state.renderState[i] = rect.Rectangle()
			}
			return state, state.renderState
		}).
		Debug().
		Main()
}
