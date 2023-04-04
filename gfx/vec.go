package gfx

import (
	"image"
	"math"
)

type Vec struct{ X, Y float64 }

func V(x, y float64) Vec { return Vec{x, y} }

func (v Vec) Map(f func(float64) float64) Vec { return Vec{f(v.X), f(v.Y)} }
func (v Vec) Add(v2 Vec) Vec                  { return Vec{v.X + v2.X, v.Y + v2.Y} }
func (v Vec) Sub(v2 Vec) Vec                  { return Vec{v.X - v2.X, v.Y - v2.Y} }
func (v Vec) Mul(s float64) Vec               { return Vec{v.X * s, v.Y * s} }
func (v Vec) Div(s float64) Vec               { return Vec{v.X / s, v.Y / s} }
func (v Vec) LenSquared() float64             { return v.X*v.X + v.Y*v.Y }
func (v Vec) Len() float64                    { return float64(math.Sqrt(float64(v.LenSquared()))) }

func (v Vec) Unit() Vec {
	if len := v.Len(); len != 0 {
		return v.Div(len)
	}
	return Vec{}
}

func (v Vec) Point() image.Point { return image.Point{int(v.X), int(v.Y)} }
func (v Vec) Round() Vec         { return Vec{math.Round(v.X), math.Round(v.Y)} }
func (v Vec) Floor() Vec         { return Vec{math.Floor(v.X), math.Floor(v.Y)} }
func (v Vec) Ceil() Vec          { return Vec{math.Ceil(v.X), math.Ceil(v.Y)} }
func (v Vec) Abs() Vec           { return Vec{math.Abs(v.X), math.Abs(v.Y)} }
func (v Vec) Min(v2 Vec) Vec     { return Vec{math.Min(v.X, v2.X), math.Min(v.Y, v2.Y)} }
func (v Vec) Max(v2 Vec) Vec     { return Vec{math.Max(v.X, v2.X), math.Max(v.Y, v2.Y)} }
