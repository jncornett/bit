package framework

import (
	"image"
	"sync"
)

type Canvas struct {
	back, front *image.NRGBA
	ready       bool
	bounds      image.Rectangle
	mu          sync.RWMutex
}

type CanvasContextKey struct{}

func NewCanvas(width, height int) *Canvas {
	bounds := image.Rect(0, 0, width, height)
	return &Canvas{
		back:   image.NewNRGBA(bounds),
		front:  image.NewNRGBA(bounds),
		bounds: bounds,
	}
}

func (c *Canvas) Bounds() image.Rectangle { return c.bounds }

func (c *Canvas) WriteBuffer(f func(img *image.NRGBA)) {
	WithLock(c.mu.RLocker(), func() {
		f(c.back)
		c.ready = true
	})
}

func (c *Canvas) NextReadBuffer() *image.NRGBA {
	WithTryLock(&c.mu, func() {
		if c.ready {
			c.back, c.front = c.front, c.back
			c.ready = false
		}
	})
	return c.front
}

func WithLock(l sync.Locker, f func()) {
	l.Lock()
	defer l.Unlock()
	f()
}

type TryLocker interface {
	sync.Locker
	TryLock() bool
}

func WithTryLock(l TryLocker, f func()) (ok bool) {
	if l.TryLock() {
		defer l.Unlock()
		f()
		return true
	}
	return false
}
