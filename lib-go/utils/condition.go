package utils

type Condition struct {
	cond    bool
	waiting []chan<- struct{}
}

func NewCondition() *Condition {
	return &Condition{
		cond:    false,
		waiting: []chan<- struct{}{},
	}
}

func (c *Condition) Wait() {
	if c.cond {
		return
	}
	b := make(chan struct{})
	c.waiting = append(c.waiting, b)
	<-b
}

func (c *Condition) Fire() {
	if c.cond {
		// panic("A condition cannot be fired more than once!")
		return
	}
	c.cond = true
	for _, ch := range c.waiting {
		ch <- struct{}{}
	}
}

func (c *Condition) IsFired() bool {
	return c.cond
}
