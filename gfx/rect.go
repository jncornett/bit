package gfx

import "image"

type Rect struct {
	Min, Max Vec
}

func (r Rect) Size() Vec {
	return r.Max.Sub(r.Min)
}

func (r Rect) Center() Vec {
	return r.Min.Add(r.Size().Mul(0.5))
}

func (r Rect) Add(v Vec) Rect {
	return Rect{r.Min.Add(v), r.Max.Add(v)}
}

func (r Rect) Rectangle() image.Rectangle {
	return image.Rectangle{r.Min.Point(), r.Max.Point()}
}
