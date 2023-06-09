package handle

import (
	"bytes"
	ddt "differentiable/datatypes"
	"differentiable/pb/pdict"
	"differentiable/protocol"
	"differentiable/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"google.golang.org/protobuf/proto"
)

type Shipper struct {
	WopChan   <-chan ddt.WriteOp
	WopBuffer utils.Queue[ddt.WriteOp]
	Offset    int
	Key       string
	Ip        string
	InvId     string
	DagId     string
	Period    time.Duration
}

func NewShipper(key string, ip string, invId string, dagId string, ch <-chan ddt.WriteOp, period time.Duration) *Shipper {
	return &Shipper{
		Key:       key,
		Ip:        ip,
		InvId:     invId,
		DagId:     dagId,
		WopChan:   ch,
		WopBuffer: *utils.NewQueue[ddt.WriteOp](),
		Offset:    0,
		Period:    period,
	}
}

func (s *Shipper) Run() {
	if s.Period == 0 {
		for {
			select {
			case data, ok := <-s.WopChan:
				if ok {
					s.WopBuffer.Enqueue(data)
				} else {
					s.Ship()
					s.End()
					return
				}
			}
		}
	} else {
		ticker := time.NewTicker(s.Period)
		defer ticker.Stop()

		for {
			select {
			case data, ok := <-s.WopChan:
				if ok {
					s.WopBuffer.Enqueue(data)
				} else {
					s.End()
					return
				}
			case <-ticker.C:
				// start := time.Now()
				s.Ship()
				// s.Period = time.Since(start)
				// ticker.Reset(s.Period)
			}
		}
	}
}

func (s *Shipper) Ship() {
	size := s.WopBuffer.Len()
	if size == 0 {
		return
	}

	dst := fmt.Sprintf("http://%s:8080/update", s.Ip)
	wopsBytes := [][]byte{}
	for !s.WopBuffer.IsEmpty() {
		wop := s.WopBuffer.Dequeue()
		kindIsDictSet := wop.Kind() == "DictSet"
		wopDictSet, ok := wop.(*ddt.DictSetOp)
		if kindIsDictSet && ok {
			_, isAnySlice := wopDictSet.V.([]any)
			_, isFloatSlice := wopDictSet.V.([]float64)

			ok = isAnySlice || isFloatSlice
		}
		if ok {
			var dictSetMsg pdict.DictSetOp
			dictSetMsg.Key = wopDictSet.K
			dictSetMsg.Value = utils.ConvertToSlice[float64](wopDictSet.V)
			msgBytes, err := proto.Marshal(&dictSetMsg)
			if err != nil {
				panic(fmt.Sprintf("pb marshalling error: %v", err))
			}
			wopsBytes = append(wopsBytes, msgBytes)
		} else {
			msgBytes, err := utils.Marshal(wop)
			if err != nil {
				panic(fmt.Sprintf("marshalling error: %v", err))
			}
			wopsBytes = append(wopsBytes, msgBytes)
		}
	}
	to := s.Offset + size
	req := protocol.UpdateReq{
		Key:   s.Key,
		Wops:  wopsBytes,
		DagId: s.DagId,
		From:  s.Offset,
		To:    to,
	}

	s.Offset = to

	reqJson, err := json.Marshal(req)
	if err != nil {
		panic(fmt.Sprintf("json marshalling error: %v", err))
	}
	// log.Printf("Shipping %s, from: %d, to: %d", s.Key, to-size, to)
	if to-size == 0 {
		log.Printf("Starts saving %s: %v", s.Key, time.Now())
	}
	http.Post(dst, "application/json", bytes.NewReader(reqJson))
}

func (s *Shipper) End() {
	dst := fmt.Sprintf("http://%s:8080/end", s.Ip)

	req := protocol.EndReq{
		Key:   s.Key,
		DagId: s.DagId,
	}
	reqJson, err := json.Marshal(req)
	if err != nil {
		panic(fmt.Sprintf("json marshalling error: %v", err))
	}
	http.Post(dst, "application/json", bytes.NewReader(reqJson))
}
