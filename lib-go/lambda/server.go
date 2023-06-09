package lambda

import (
	"bytes"
	"context"
	ddt "differentiable/datatypes"
	"differentiable/protocol"
	"differentiable/server"
	"differentiable/stub"
	"differentiable/stub/handle"
	"differentiable/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type KeyWithDag struct {
	DagId string `json:"dagId"`
	Key   string `json:"key"`
}

type ChituServer struct {
	objs        map[string]handle.ObjectHandle // objId -> local
	invocations map[string]stub.AgentStub
	// keyeds      map[KeyWithDag]string // key, dagId -> objId. Here are only writers, readers and imports

	keyedExports map[KeyWithDag]string
	keyedImports map[KeyWithDag]string

	writeChan chan ddt.DWriteMsg
	endChan   chan ddt.DEndMsg
	consChan  chan ddt.DConsMsg
	awaitChan chan ddt.DAwaitMsg
	readChan  chan ddt.DReadMsg

	exportChan  chan ddt.ExportMsg
	importChan  chan ddt.ImportMsg
	triggerChan chan ddt.TriggerMsg

	period time.Duration
}

const CHANNEL_BUFFER_SIZE = 4096

func NewChituServer(f server.FunctionType) (*ChituServer, context.CancelFunc) {
	_, cancel := context.WithCancel(context.Background())
	s := &ChituServer{
		objs:        map[string]handle.ObjectHandle{},
		invocations: map[string]stub.AgentStub{},
		// keyeds:      map[KeyWithDag]string{},
		keyedExports: map[KeyWithDag]string{},
		keyedImports: map[KeyWithDag]string{},
		consChan:     make(chan ddt.DConsMsg, CHANNEL_BUFFER_SIZE),
		writeChan:    make(chan ddt.DWriteMsg, CHANNEL_BUFFER_SIZE),
		endChan:      make(chan ddt.DEndMsg, CHANNEL_BUFFER_SIZE),
		awaitChan:    make(chan ddt.DAwaitMsg, CHANNEL_BUFFER_SIZE),
		readChan:     make(chan ddt.DReadMsg, CHANNEL_BUFFER_SIZE),
		exportChan:   make(chan ddt.ExportMsg, CHANNEL_BUFFER_SIZE),
		importChan:   make(chan ddt.ImportMsg, CHANNEL_BUFFER_SIZE),
		triggerChan:  make(chan ddt.TriggerMsg, CHANNEL_BUFFER_SIZE),
	}

	return s, cancel
}

func (s *ChituServer) serveUserRequest(ctx context.Context) {
	defer log.Printf("service stopped")
	for {
		select {
		case write := <-s.writeChan:
			s.handleWrite(write)
		case cons := <-s.consChan:
			s.handleCons(cons)
		case end := <-s.endChan:
			s.handleUserEnd(end)
		case await := <-s.awaitChan:
			go s.handleAwait(await) // await is a blocking function
		case read := <-s.readChan:
			s.handleRead(read)
		case export := <-s.exportChan:
			s.handleExport(export)
		case importMsg := <-s.importChan:
			s.handleImport(importMsg)
		case trigger := <-s.triggerChan:
			s.handleTrigger(trigger)
		case <-ctx.Done():
			return
		}
	}
}

func (s *ChituServer) handleWrite(dWrite ddt.DWriteMsg) {
	obj := s.lookupObj(dWrite.Id)
	if obj.IsWriter() {
		writer := obj.(*handle.WriterHandle)
		writer.Update(dWrite.Wop)
		dWrite.Res <- ddt.DWriteRes{}
	} else {
		panic("User error: write on a Reader")
	}
}

func (s *ChituServer) handleCons(cons ddt.DConsMsg) {
	objId := utils.NewUId()
	val := ddt.DVAL_CONS[cons.DType]()
	s.objs[objId] = handle.NewWriterHandle(val)
	cons.Res <- ddt.DConsRes{
		Id: objId,
	}
}

func (s *ChituServer) handleAwait(await ddt.DAwaitMsg) {
	obj := s.lookupObj(await.Id)
	obj.Await()
	obj.ApplyWopBuffer()
	await.Res <- ddt.DAwaitRes{
		Val: obj.GetVal(),
	}
}

func (s *ChituServer) handleUserEnd(end ddt.DEndMsg) {
	obj := s.lookupObj(end.Id)
	if obj.IsWriter() {
		obj.End(handle.EXPORT_DEFAULT_RANK)
	} else {
		panic("User error: end on a reader")
	}
}

func (s *ChituServer) handleRead(read ddt.DReadMsg) {
	readerId := utils.NewUId()
	val := ddt.DVAL_CONS[read.Rop.GetRetDType()]()
	writerCount := len(read.Ids)
	readerObj := handle.NewReaderHandle(val, read.Rop, writerCount)
	s.objs[readerId] = readerObj
	for rank, id := range read.Ids {
		writer := s.lookupObj(id)
		writer.AddReader(readerObj, rank)
	}
	read.Res <- ddt.DReadRes{
		ReaderId: readerId,
	}
}

func (s *ChituServer) handleExport(export ddt.ExportMsg) {
	key := export.Key
	dagId := export.DagId
	keyWithDag := KeyWithDag{
		Key:   key,
		DagId: dagId,
	}
	// s.keyeds[keyWithDag] = export.Id
	s.keyedExports[keyWithDag] = export.Id
	req := protocol.CoordPutReq{
		Key:   key,
		DagId: dagId,
		InvId: export.InvId,
	}
	reqJson, err := json.Marshal(req)
	if err != nil {
		panic("Error in marshalling export request: marshalling error")
	}
	res, err := http.Post(protocol.COORDINATOR_URL+"/put", "application/json", bytes.NewBuffer(reqJson))
	if err != nil {
		panic("Request to coordinator failed")
	}

	defer res.Body.Close()
	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		panic(fmt.Sprintf("Response from coordinator is invalid: %v", err))
	}
	var response = protocol.CoordPutRes{}
	err = json.Unmarshal(resBytes, &response)
	if err != nil {
		panic(fmt.Sprintf("Response from coordinator is not a valid CoordPutRes: %v", err))
	}
	// obj := s.lookupObjByKey(keyWithDag)
	obj := s.lookupKeyedExport(keyWithDag)
	for i := 0; i < len(response.Ips); i++ {
		ip := response.Ips[i]
		invId := response.InvIds[i]
		eh := handle.NewExportHandle(key, ip, invId, dagId, s.period)
		obj.AddReader(eh, handle.EXPORT_DEFAULT_RANK)
	}
}

func (s *ChituServer) handleImport(importMsg ddt.ImportMsg) {
	id := utils.NewUId()
	val := ddt.DVAL_CONS[importMsg.DType]()
	importHandle := handle.NewImportHandle(val, importMsg.Key)
	s.objs[id] = importHandle
	keyWithDag := KeyWithDag{
		Key:   importMsg.Key,
		DagId: importMsg.DagId,
	}
	// s.keyeds[keyWithDag] = id
	s.keyedImports[keyWithDag] = id

	req := protocol.CoordGetReq{
		Key:   importMsg.Key,
		InvId: importMsg.InvId,
		DagId: importMsg.DagId,
	}
	reqJson, err := json.Marshal(req)
	if err != nil {
		panic("Error in marshalling export request: marshalling error")
	}
	http.Post(protocol.COORDINATOR_URL+"/get", "application/json", bytes.NewBuffer(reqJson))

	importMsg.Res <- ddt.ImportRes{
		Id: id,
	}
}

func (s *ChituServer) handleTrigger(trigger ddt.TriggerMsg) {
	req := protocol.CoordTriggerReq{
		Fname: trigger.Fname,
		Args:  trigger.Args,
		DagId: trigger.DagId,
	}
	reqJson, err := json.Marshal(req)
	if err != nil {
		panic("Error in marshalling trigger request: marshalling error")
	}
	http.Post(protocol.COORDINATOR_URL+"/trigger", "application/json", bytes.NewBuffer(reqJson))
}

func (s *ChituServer) lookupObj(id string) handle.ObjectHandle {
	obj, ok := s.objs[id]
	if !ok {
		panic("write with unknown id")
	}
	return obj
}

func (s *ChituServer) lookupKeyedExport(key KeyWithDag) handle.ObjectHandle {
	id, ok := s.keyedExports[key]
	if !ok {
		panic(fmt.Sprintf("looking for export with unknown key: %s", key.Key))
	}
	return s.lookupObj(id)
}

func (s *ChituServer) lookupKeyedImport(key KeyWithDag) handle.ObjectHandle {
	id, ok := s.keyedImports[key]
	if !ok {
		panic(fmt.Sprintf("looking for import with unknown key: %s", key.Key))
	}
	return s.lookupObj(id)
}
