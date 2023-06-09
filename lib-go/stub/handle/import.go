package handle

import (
	ddt "differentiable/datatypes"
	"log"
	"time"
)

type ImportHandle struct {
	BaseHandle

	Key     string
	Unready UnreadyBuffer
}

func NewImportHandle(val ddt.DVal, key string) *ImportHandle {
	log.Printf("Importing: %s", key)
	return &ImportHandle{
		BaseHandle: *newBaseHandle(val, 1),
		Key:        key,
		Unready:    *NewUnreadyBuffer(),
	}
}

func (h *ImportHandle) IsWriter() bool {
	return true
}

// TODO: removal
// Same as BaseHandle
func (h *ImportHandle) End(n int) {
	log.Printf("Ends loading %s: %v", h.Key, time.Now())
	h.Barrier.Fire(n)
	if h.IsEnd() {
		for _, reader := range h.Readers {
			reader.End(reader.Rank)
		}
	}
}

func (h *ImportHandle) UpdateRemote(wops []ddt.WriteOp, from, to int) {
	// log.Printf("Updating: %s, from: %d, to: %d", h.Key, from, to)
	if from == 0 {
		log.Printf("Starts loading %s: %v", h.Key, time.Now())
	}
	h.Unready.Insert(wops, from, to)
	readys := h.Unready.RemoveUntilBreak()
	for _, wop := range readys {
		h.update(wop)
	}
}

func (h *ImportHandle) update(wop ddt.WriteOp) {
	h.WopBuffer.Enqueue(wop)
	// h.Val.Apply(wop)
	for _, reader := range h.Readers {
		reader.Lock()
		reader.Read(wop, reader.Rank)
		reader.Unlock()
	}
}
