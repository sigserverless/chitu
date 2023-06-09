package handle

import (
	ddt "differentiable/datatypes"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type ExportHandle struct {
	Shipper   Shipper
	WopChan   chan<- ddt.WriteOp
	Mutex     sync.Mutex
	Heartbeat atomic.Uint32
}

func (*ExportHandle) IsWriter() bool {
	return false
}

func NewExportHandle(key, ip, invId, dagId string, period time.Duration) *ExportHandle {
	log.Printf("Exporting: %s, ip: %s", key, ip)
	ch := make(chan ddt.WriteOp, 1024)
	shipper := NewShipper(key, ip, invId, dagId, ch, period)
	go shipper.Run()
	return &ExportHandle{
		Shipper:   *shipper,
		WopChan:   ch,
		Mutex:     sync.Mutex{},
		Heartbeat: atomic.Uint32{},
	}
}

func (h *ExportHandle) Read(wop ddt.WriteOp, _rank int) {
	h.WopChan <- wop
	h.Heartbeat.Store(0)
}

func (h *ExportHandle) End(_rank int) {
	// ticker := time.NewTicker(h.Shipper.Period * 2)
	ticker := time.NewTicker(time.Millisecond * 50)
	go func() {
		for {
			<-ticker.C
			hb := h.Heartbeat.Add(1)
			if hb >= 4 {
				log.Printf("Ends saving %s: %v", h.Shipper.Key, time.Now())
				close(h.WopChan)
				return
			} else {
				// ticker.Reset(h.Shipper.Period * 2)
				ticker.Reset(time.Millisecond * 50)
			}
		}
	}()
}

func (h *ExportHandle) Lock() {
	h.Mutex.Lock()
}

func (h *ExportHandle) Unlock() {
	h.Mutex.Unlock()
}
