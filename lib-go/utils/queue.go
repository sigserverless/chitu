package utils

type Queue[T any] struct {
	items []T
}

func (q *Queue[T]) Enqueue(item T) {
	q.items = append(q.items, item)
}

func (q *Queue[T]) Dequeue() T {
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

func (q *Queue[T]) IsEmpty() bool {
	return len(q.items) == 0
}

func (q *Queue[T]) Len() int {
	return len(q.items)
}

// iteration on queue
func (q *Queue[T]) Iter() <-chan T {
	ch := make(chan T)
	go func() {
		for _, item := range q.items {
			ch <- item
		}
		close(ch)
	}()
	return ch
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{items: []T{}}
}
