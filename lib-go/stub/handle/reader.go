package handle

import (
	ddt "differentiable/datatypes"
	"sync"
)

type ReaderHandle struct {
	BaseHandle

	Rop   ddt.ReadOp
	Mutex sync.Mutex
}

func NewReaderHandle(val ddt.DVal, rop ddt.ReadOp, writerCount int) *ReaderHandle {
	return &ReaderHandle{
		BaseHandle: *newBaseHandle(val, writerCount),
		Rop:        rop,
		Mutex:      sync.Mutex{},
	}
}

func (*ReaderHandle) IsWriter() bool {
	return false
}

func (h *ReaderHandle) Read(wop ddt.WriteOp, rank int) {
	// TODO: the WopBuffer is not used. Consider removal
	targetOp := wop.Diff(h.Rop, rank, h.Val)
	h.Val.Apply(targetOp)
	for _, reader := range h.Readers {
		// go func() {
		// 	reader.Lock()
		// 	reader.Read(targetOp, reader.Rank)
		// 	reader.Unlock()
		// }()
		reader.Lock()
		reader.Read(targetOp, reader.Rank)
		reader.Unlock()
	}
}

func (h *ReaderHandle) Lock() {
	h.Mutex.Lock()
}

func (h *ReaderHandle) Unlock() {
	h.Mutex.Unlock()
}
