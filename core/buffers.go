package core

type Buffers[T any] [2]*T

func NewBuffers[T any](a, b T) Buffers[T] { return Buffers[T]{&a, &b} }

func (b *Buffers[T]) Current() *T { return b[0] }

func (b *Buffers[T]) Swap() { b[0], b[1] = b[1], b[0] }
