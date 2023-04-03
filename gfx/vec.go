package gfx

import (
	"image"
	"math"
)

type Vec struct{ X, Y float64 }

func V(x, y float64) Vec { return Vec{x, y} }

func (v Vec) Add(v2 Vec) Vec {
	return Vec{v.X + v2.X, v.Y + v2.Y}
}

func (v Vec) Sub(v2 Vec) Vec {
	return Vec{v.X - v2.X, v.Y - v2.Y}
}

func (v Vec) Mul(s float64) Vec {
	return Vec{v.X * s, v.Y * s}
}

func (v Vec) Div(s float64) Vec {
	return Vec{v.X / s, v.Y / s}
}

func (v Vec) LenSquared() float64 { return v.X*v.X + v.Y*v.Y }

func (v Vec) Len() float64 { return float64(math.Sqrt(float64(v.LenSquared()))) }

func (v Vec) Unit() Vec {
	if len := v.Len(); len != 0 {
		return v.Div(len)
	}
	return Vec{}
}

func (v Vec) Point() image.Point { return image.Point{int(v.X), int(v.Y)} }
