package main

import (
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
)

func main() {
	go func() {
		err := func() error {
			inputs := make(chan Event, 128)
			windowSize := image.Pt(800, 600)
			clock, stop := newClock(time.Second / 60)
			defer stop()
			frames := newFrameClock(clock, inputs)
			state := newGameState(windowSize)
			toRender := mapChannel(frames, bindState(state, update))
			renders := newRenderer(toRender, render, newBuffers(defineImageBuffer(windowSize), clearImageBuffer))
			w := app.NewWindow()
			var ops op.Ops
		Resize:
			for {
				e := <-w.Events()
				switch e := e.(type) {
				case system.DestroyEvent:
					return e.Err
				case system.FrameEvent:
					gtx := layout.NewContext(&ops, e)
					w.Option(app.Size(gtx.Metric.PxToDp(windowSize.X), gtx.Metric.PxToDp(windowSize.Y)))
					op.InvalidateOp{}.Add(gtx.Ops)
					e.Frame(gtx.Ops)
					break Resize
				}
			}
			var imgBuf *image.NRGBA
		ImageBuffer:
			for {
				select {
				case e := <-w.Events():
					switch e := e.(type) {
					case system.DestroyEvent:
						return e.Err
					case system.FrameEvent:
						gtx := layout.NewContext(&ops, e)
						op.InvalidateOp{}.Add(gtx.Ops)
						e.Frame(gtx.Ops)
					}
				case img := <-renders:
					imgBuf = img
					break ImageBuffer
				}
			}
			// MainLoop:
			tag := new(int)
			for {
				e := <-w.Events()
				switch e := e.(type) {
				case system.DestroyEvent:
					return e.Err
				case system.FrameEvent:
					for _, ev := range e.Queue.Events(tag) {
						if x, ok := ev.(key.Event); ok {
							inputs <- Event{Key: x}
						}
					}
					gtx := layout.NewContext(&ops, e)
					key.InputOp{Tag: tag}.Add(gtx.Ops)
					select {
					case img := <-renders:
						imgBuf = img
					default:
					}
					paint.NewImageOp(imgBuf).Add(gtx.Ops)
					paint.PaintOp{}.Add(gtx.Ops)
					op.InvalidateOp{}.Add(gtx.Ops)
					e.Frame(gtx.Ops)
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

func defineImageBuffer(size image.Point) func() *image.NRGBA {
	return func() *image.NRGBA {
		return image.NewNRGBA(image.Rectangle{Max: size})
	}
}

func clearImageBuffer(img *image.NRGBA) {
	draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)
}

type gameState struct {
	entities  []entity
	player    entity
	renderSet renderSet
	bounds    image.Rectangle
	speed     float64
	keyState  map[string]bool
}

type entity struct {
	color color.NRGBA
	size  float64
	pos   [2]float64
	vel   [2]float64
}

func newGameState(windowSize image.Point) *gameState {
	const n = 100
	const margin = 20
	const size = 20
	var colors = []color.NRGBA{
		{R: 0x8f, A: 0xff},
		{G: 0x8f, A: 0xff},
	}
	randPos := func(min, max [2]float64) [2]float64 {
		return [2]float64{
			rand.Float64()*(max[0]-min[0]) + min[0],
			rand.Float64()*(max[1]-min[1]) + min[1],
		}
	}
	randVel := func() [2]float64 {
		theta := rand.Float64() * 2 * math.Pi
		return [2]float64{math.Cos(theta), math.Sin(theta)}
	}
	spawnBounds := [][2]float64{{float64(margin), float64(margin)}, {float64(windowSize.X - margin), float64(windowSize.Y - margin)}}
	entities := make([]entity, n)
	renderSet := make(renderSet, n+1)
	for i := range entities {
		entities[i] = entity{
			color: colors[rand.Intn(len(colors))],
			pos:   randPos(spawnBounds[0], spawnBounds[1]),
			vel:   randVel(),
			size:  size,
		}
		renderSet[i] = entityToRender(entities[i])
	}
	player := entity{
		color: color.NRGBA{B: 0xff, A: 0xff},
		size:  30,
		pos:   [2]float64{float64(windowSize.X) / 2, float64(windowSize.Y) / 2},
	}
	renderSet[n] = entityToRender(player)
	return &gameState{
		entities:  entities,
		player:    player,
		renderSet: renderSet,
		bounds:    image.Rectangle{Max: windowSize},
		speed:     250,
		keyState:  map[string]bool{},
	}
}

func entityToRender(e entity) renderOne {
	p := image.Pt(int(e.pos[0]), int(e.pos[1]))
	sz := image.Pt(int(e.size), int(e.size)).Div(2)
	return renderOne{
		src: image.NewUniform(e.color),
		rect: image.Rectangle{
			Min: p.Sub(sz),
			Max: p.Add(sz),
		},
	}
}

func update(state *gameState, frame Frame) (renderSet, *gameState) {
	delta := frame.Tick[1].Sub(frame.Tick[0]).Seconds()
	abs := func(x float64) float64 {
		if x < 0 {
			return -x
		}
		return x
	}
	for i := range state.entities {
		e := &state.entities[i]
		e.pos[0] += e.vel[0] * state.speed * delta
		e.pos[1] += e.vel[1] * state.speed * delta
		if e.pos[0] < float64(state.bounds.Min.X) {
			e.vel[0] = abs(e.vel[0])
		} else if e.pos[0] > float64(state.bounds.Max.X) {
			e.vel[0] = -abs(e.vel[0])
		}
		if e.pos[1] < float64(state.bounds.Min.Y) {
			e.vel[1] = abs(e.vel[1])
		} else if e.pos[1] > float64(state.bounds.Max.Y) {
			e.vel[1] = -abs(e.vel[1])
		}
		state.renderSet[i] = entityToRender(*e)
	}
	for _, e := range frame.Events {
		state.keyState[e.Key.Name] = e.Key.State == key.Press
	}
	var pv [2]float64
	for k, v := range keyVelMap {
		if state.keyState[k] {
			pv[0] += v[0]
			pv[1] += v[1]
		}
	}
	state.player.pos[0] += pv[0] * state.speed * delta
	state.player.pos[1] += pv[1] * state.speed * delta
	state.renderSet[len(state.renderSet)-1] = entityToRender(state.player)
	return state.renderSet, state
}

var keyVelMap = map[string][2]float64{
	"↑": {0, -1},
	"↓": {0, 1},
	"←": {-1, 0},
	"→": {1, 0},
}

func render(rs renderSet, img *image.NRGBA) {
	for _, r := range rs {
		draw.Draw(img, r.rect, r.src, image.Point{}, draw.Src)
	}
}

func newClock(d time.Duration) (<-chan [2]time.Time, func()) {
	out := make(chan [2]time.Time)
	done := make(chan struct{})
	go func() {
		defer close(out)
		tk := time.NewTicker(d)
		defer tk.Stop()
		var tm [2]time.Time
		tm[1] = <-tk.C
		for {
			select {
			case <-done:
				return
			case t := <-tk.C:
				tm[0], tm[1] = tm[1], t
				select {
				case <-done:
					return
				case out <- tm:
				}
			}
		}
	}()
	return out, func() {
		close(done)
		<-out
	}
}

type Event struct {
	Key key.Event
}

type Frame struct {
	Tick   [2]time.Time
	Events []Event
}

func newFrameClock(ck <-chan [2]time.Time, events <-chan Event) <-chan Frame {
	out := make(chan Frame)
	go func() {
		defer close(out)
		for {
			var frame Frame
		Receive:
			for {
				select {
				case e, ok := <-events:
					if !ok {
						events = nil
						continue
					}
					frame.Events = append(frame.Events, e)
				case tm, ok := <-ck:
					if !ok {
						return
					}
					frame.Tick = tm
					break Receive
				}
			}
		Send:
			for {
				select {
				case e, ok := <-events:
					if !ok {
						events = nil
						continue
					}
					frame.Events = append(frame.Events, e)
				case tm, ok := <-ck:
					if !ok {
						return
					}
					frame.Tick = tm
				case out <- frame:
					break Send
				}
			}
		}
	}()
	return out
}

func newRenderer[T, R any](in <-chan T, render func(T, *R), bufs buffers[R]) <-chan *R {
	out := make(chan *R)
	go func() {
		defer close(out)
		for arg := range in {
		Send:
			for {
				render(arg, bufs.Active())
				select {
				case a, ok := <-in:
					if !ok {
						return
					}
					arg = a
				case out <- bufs.Active():
					bufs.Swap()
					break Send
				}
			}
		}
	}()
	return out
}

type renderSet []renderOne

type renderOne struct {
	src  image.Image
	rect image.Rectangle
}

type buffers[T any] struct {
	b     [2]*T
	reset func(*T)
}

func newBuffers[T any](new func() *T, reset func(*T)) buffers[T] {
	return buffers[T]{b: [2]*T{new(), new()}, reset: reset}
}

func (b *buffers[T]) Active() *T { return b.b[0] }

func (b *buffers[T]) Swap() {
	b.b[0], b.b[1] = b.b[1], b.b[0]
	if b.reset != nil {
		b.reset(b.b[0])
	}
}

func mapChannel[T, R any](in <-chan T, f func(T) R) <-chan R {
	out := make(chan R)
	go func() {
		defer close(out)
		for v := range in {
			out <- f(v)
		}
	}()
	return out
}

func bindState[S, T, R any](init S, f func(S, T) (R, S)) func(T) R {
	state := init
	return func(t T) R {
		r, s := f(state, t)
		state = s
		return r
	}
}
