package handle

import (
	ddt "differentiable/datatypes"
)

type WriterHandle struct {
	BaseHandle
}

func (*WriterHandle) IsWriter() bool {
	return true
}

func NewWriterHandle(val ddt.DVal) *WriterHandle {
	return &WriterHandle{
		BaseHandle: *newBaseHandle(val, 1),
	}
}

func (h *WriterHandle) Update(wop ddt.WriteOp) {
	h.WopBuffer.Enqueue(wop)
	// h.Val.Apply(wop)
	for _, reader := range h.Readers {
		reader.Lock()
		reader.Read(wop, reader.Rank)
		reader.Unlock()
		// go func() {
		// 	reader.Lock()
		// 	reader.Read(targetOp, reader.Rank)
		// 	reader.Unlock()
		// }()
	}
}
