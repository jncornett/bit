package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
)

func main() {
	img := image.NewNRGBA(image.Rect(0, 0, 640, 480))
	draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)
	rect := image.Rectangle{Min: image.Pt(320, 240), Max: image.Pt(340, 260)}
	draw.Draw(img, rect, image.NewUniform(color.NRGBA{R: 0xff, A: 0xff}), image.Point{}, draw.Over)
	func() {
		f, err := os.Create("test.png")
		if err != nil {
			log.Println(err)
			return
		}
		defer f.Close()
		if err := png.Encode(f, img); err != nil {
			log.Println(err)
			return
		}
	}()
}

type FillRect struct {
	image.Rectangle
	m color.Model
	c color.NRGBA
}

func (f *FillRect) ColorModel() color.Model { return f.m }
func (f *FillRect) At(x, y int) color.Color { return f.c }
