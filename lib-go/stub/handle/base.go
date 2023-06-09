package handle

import (
	ddt "differentiable/datatypes"
	"differentiable/utils"
)

// ReaderHandle, WriterHandle and ImportHandle implement  
// ObjectHandle
type ObjectHandle interface {
	AddReader(Reader, int)
	Await()
	End(int)
	ApplyWopBuffer()
	IsWriter() bool
	GetVal() ddt.DVal
}

// ExportHandle, ReaderHandle implement Reader
type Reader interface {
	Read(ddt.WriteOp, int)
	End(int)
	Lock() // The mutex is primarily for AddReader
	Unlock()
}

type ReaderWithRank struct {
	Reader
	Rank int
}

type BaseHandle struct {
	Val       ddt.DVal
	WopBuffer utils.Queue[ddt.WriteOp]
	Readers   []*ReaderWithRank
	Barrier   utils.Barrier
}

func newBaseHandle(val ddt.DVal, writerCount int) *BaseHandle {
	return &BaseHandle{
		Val:       val,
		WopBuffer: *utils.NewQueue[ddt.WriteOp](),
		Readers:   []*ReaderWithRank{},
		Barrier:   *utils.NewBarrier(writerCount),
	}
}

const EXPORT_DEFAULT_RANK = 0

// rank only works when the struct is exact ReaderHandle.
// When adding an ExportHandle, the rank is EXPORT_DEFAULT_RANK
// which will always be discarded.
//
// AddReader needs a mutex to ensure when a reader is added,
// it can read all the wops already in the buffer.
func (h *BaseHandle) AddReader(reader Reader, rank int) {
	readerWithRank := ReaderWithRank{
		reader,
		rank,
	}
	reader.Lock()
	h.Readers = append(h.Readers, &readerWithRank)
	reader.Read(h.Val.ToWop(), rank)
	for wop := range h.WopBuffer.Iter() {
		reader.Read(wop, rank)
	}
	if h.IsEnd() {
		reader.End(rank)
	}
	reader.Unlock()
}

func (h *BaseHandle) ApplyWopBuffer() {
	for !h.WopBuffer.IsEmpty() {
		h.Val.Apply(h.WopBuffer.Dequeue())
	}
	// for wop := range h.WopBuffer.Iter() {
	// 	h.Val.Apply(wop)
	// }
}

func (h *BaseHandle) GetVal() ddt.DVal {
	return h.Val
}

func (h *BaseHandle) End(n int) {
	h.Barrier.Fire(n)
	if h.IsEnd() {
		for _, reader := range h.Readers {
			reader.End(reader.Rank)
		}
	}
}

func (h *BaseHandle) Await() {
	h.Barrier.Wait()
}

func (h *BaseHandle) IsEnd() bool {
	return h.Barrier.IsFired()
}
