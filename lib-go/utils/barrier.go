package utils

type Barrier struct {
	conds []*Condition
}

func NewBarrier(n int) *Barrier {
	conds := []*Condition{}
	for i := 0; i < n; i++ {
		conds = append(conds, NewCondition())
	}
	return &Barrier{
		conds: conds,
	}
}

func (b *Barrier) IsFired() bool {
	pass := true
	for _, c := range b.conds {
		pass = pass && c.IsFired()
	}
	return pass
}

func (b *Barrier) Wait() {
	for _, c := range b.conds {
		c.Wait()
	}
}

func (b *Barrier) Fire(i int) {
	b.conds[i].Fire()
}
