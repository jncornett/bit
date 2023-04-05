package bit

import (
	"github.com/jncornett/chans"
)

func runEngine(framerate FPS, events <-chan any, update func(Frame, RenderTarget), render func(RenderSource) Widget) (widgets <-chan Widget, stop func()) {
	rs := newRenderState()
	clock, stop := MakeClock(framerate)()
	frames := chans.MapSidechain(MakeFrame)(clock, chans.Batch(events))
	widgets = chans.Map(func(frame Frame) Widget {
		update(frame, rs)
		return render(rs)
	})(frames)
	return widgets, stop
}
