package handle

import ddt "differentiable/datatypes"

type UnreadyBuffer struct {
	Head *UnreadyWop
	Last int
}

type UnreadyWop struct {
	Wops []ddt.WriteOp
	From int
	To   int
	Next *UnreadyWop
}

func NewUnreadyBuffer() *UnreadyBuffer {
	return &UnreadyBuffer{
		Head: nil,
		Last: 0,
	}
}

func (b *UnreadyBuffer) Insert(wops []ddt.WriteOp, from, to int) {
	newNode := &UnreadyWop{
		Wops: wops,
		From: from,
		To:   to,
		Next: nil,
	}

	if b.Head == nil {
		b.Head = newNode
		return
	}

	if from < b.Head.From {
		newNode.Next = b.Head
		b.Head = newNode
		return
	}

	curr := b.Head
	for curr.Next != nil && curr.Next.From < from {
		curr = curr.Next
	}
	newNode.Next = curr.Next
	curr.Next = newNode
}

func (b *UnreadyBuffer) RemoveUntilBreak() []ddt.WriteOp {
	wops := []ddt.WriteOp{}
	last := b.Last

	curr := b.Head
	for curr != nil && curr.From == last {
		wops = append(wops, b.Head.Wops...)
		last = curr.To
		curr = curr.Next
	}

	b.Head = curr
	b.Last = last

	return wops
}
