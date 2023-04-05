package chans

type Chan[T any] <-chan T

func (c Chan[T]) Map(f func(T) T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for t := range c {
			out <- f(t)
		}
	}()
	return out
}

func (c Chan[T]) Filter(f func(T) bool) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for t := range c {
			if f(t) {
				out <- t
			}
		}
	}()
	return out
}

func (c Chan[T]) TakeWhile(f func(T) bool) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for t := range c {
			if !f(t) {
				return
			}
			out <- t
		}
		drain(c)
	}()
	return out
}

func (c Chan[T]) DropWhile(f func(T) bool) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for t := range c {
			if !f(t) {
				out <- t
				break
			}
		}
		pipe(c, out)
	}()
	return out
}

func (c Chan[T]) Buffer() <-chan []T {
	out := make(chan []T)
	go func() {
		defer close(out)
		for t := range c {
			buf := []T{t}
		Send:
			for {
				select {
				case t, ok := <-c:
					if !ok {
						return
					}
					buf = append(buf, t)
				case out <- buf:
					break Send
				}
			}
		}
	}()
	return out
}

func (c Chan[T]) Audit() <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for t := range c {
		Send:
			for {
				select {
				case v, ok := <-c:
					if !ok {
						return
					}
					t = v
				case out <- t:
					break Send
				}
			}
		}
	}()
	return out
}

func (c Chan[T]) Debounce() <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for t := range c {
		Send:
			for {
				select {
				case _, ok := <-c:
					if !ok {
						return
					}
				case out <- t:
					break Send
				}
			}
		}
	}()
	return out
}

func (c Chan[T]) Balance() (<-chan T, <-chan T) {
	out1 := make(chan T)
	out2 := make(chan T)
	go func() {
		defer close(out1)
		defer close(out2)
		for t := range c {
			select {
			case out1 <- t:
			case out2 <- t:
			}
		}
	}()
	return out1, out2
}

func (c Chan[T]) Broadcast() (<-chan T, <-chan T) {
	out1 := make(chan T)
	out2 := make(chan T)
	go func() {
		defer close(out1)
		defer close(out2)
		for t := range c {
			select {
			case out1 <- t:
			case out2 <- t:
			}
		}
	}()
	return out1, out2
}

func pipe[T any](in <-chan T, out chan<- T) {
	for t := range in {
		out <- t
	}
}

func drain[T any](in <-chan T) {
	for range in {
		// noop
	}
}

type Producer[T any] interface {
	Next() T
}

type Consumer[T any] interface {
	Accept(T)
}
