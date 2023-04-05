package core

func Run[ToRender, Rendered any](
	clock <-chan Tick,
	events <-chan Event,
	update func(Frame) ToRender,
	render func(ToRender, *Rendered),
	bufs Buffers[Rendered],
	rendered <-chan Rendered,
) {
	pipeline := make(chan ToRender)
	go func() {
		defer close(pipeline)
		frames := newFrameClock(clock, events)
		for frame := range frames {
			pipeline <- update(frame)
		}
	}()
	for {
		var toRender ToRender
		select {
		case r, ok := <-pipeline:
			if !ok {
				return
			}
			toRender = r
		}
		for {
			select {
			case r, ok := <-pipeline:
				if !ok {
					return
				}
			}
		}
	}
	for toRender := range pipeline {
		render(toRender, bufs.Current())
	}
}

type Clock <-chan Tick

type Frame struct {
	Tick   Tick
	Events []Event
}

type Event struct{}

type Update[State, ToRender any] func(State, Frame) (ToRender, State)

func (f Update[State, ToRender]) BindState(initial State) func(Frame) ToRender {
	state := initial
	return func(frame Frame) (toRender ToRender) {
		toRender, state = f(state, frame)
		return toRender
	}
}
