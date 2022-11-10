package collector

type collector[T any] struct {
	Add          chan T
	todo         []T
	errs         []error
	ErrCollector chan error
	done         chan struct{}
}

func New[T any]() *collector[T] {
	c := collector[T]{
		Add:          make(chan T, 8),
		todo:         make([]T, 0, 4),
		ErrCollector: make(chan error, 8),
		done:         make(chan struct{}),
	}
	go func() {
		for {
			select {
			case <-c.done:
				return
			case a := <-c.Add:
				c.todo = append(c.todo, a)
			}
		}
	}()
	go func() {
		for {
			select {
			case err := <-c.ErrCollector:
				if err != nil {
					c.errs = append(c.errs, err)
				}
			case <-c.done:
				return
			}
		}
	}()
	return &c
}

func (c *collector[T]) Done() ([]T, []error) {
	c.done <- struct{}{}
	c.done <- struct{}{}
	return c.todo, c.errs
}
