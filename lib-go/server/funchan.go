package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	ddt "differentiable/datatypes"
	"differentiable/pb/pdict"
	"differentiable/protocol"
	"differentiable/stub"
	"differentiable/stub/handle"
	"differentiable/utils"

	"google.golang.org/protobuf/proto"
)

const IP_PORT = ":8082"

type KeyWithDag struct {
	DagId string `json:"dagId"`
	Key   string `json:"key"`
}

type FunchanServer struct {
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

func NewFunchanServer(period time.Duration, f FunctionType) (*FunchanServer, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	s := &FunchanServer{
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
		period:       period,
	}
	go s.serveUserRequest(ctx)
	usercode := &GoUsercode{f}
	s.serveOuterRequest(usercode)
	return s, cancel
}

func (s *FunchanServer) serveUserRequest(ctx context.Context) {
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

func (s *FunchanServer) serveOuterRequest(usercode Usercode) {
	http.HandleFunc("/async-invoke", handleAsyncInvoke(s, usercode))
	http.HandleFunc("/invoke", handleInvoke(s, usercode))
	http.HandleFunc("/get-notify", handleGetNotify(s))
	http.HandleFunc("/update", handleUpdate(s))
	http.HandleFunc("/end", handleOtherEnd(s))
	http.HandleFunc("/", handleUserRequest(s, usercode))
	http.ListenAndServe(IP_PORT, nil)
}

func makeHandler[T any](fn func(T) ([]byte, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var args []byte

		if r.Body != nil {
			defer r.Body.Close()

			bodyBytes, err := io.ReadAll(r.Body)

			if err != nil {
				log.Fatalf("Read body error: %v", err)
			}

			args = bodyBytes
		}

		var req T
		if err := json.Unmarshal(args, &req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := fn(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			w.Write(res)
		}
	}
}

func (s *FunchanServer) forkAgent(dagId string) stub.AgentStub {
	invId := utils.NewUId()
	agent := stub.NewFunchanStub(
		invId,
		dagId,
		s.writeChan,
		s.endChan,
		s.awaitChan,
		s.consChan,
		s.readChan,
		s.exportChan,
		s.importChan,
		s.triggerChan,
	)
	s.invocations[invId] = agent
	return agent
}

func handleAsyncInvoke(s *FunchanServer, usercode Usercode) func(w http.ResponseWriter, r *http.Request) {
	return makeHandler(func(req protocol.InvokeReq) ([]byte, error) {
		dagId := req.DagId
		agent := s.forkAgent(dagId)
		go usercode.Handle(agent, []byte(req.Req))
		return []byte("Asynchronous invocation success. "), nil
	})
}

func handleInvoke(s *FunchanServer, usercode Usercode) func(w http.ResponseWriter, r *http.Request) {
	return makeHandler(func(req protocol.InvokeReq) ([]byte, error) {
		dagId := req.DagId
		agent := s.forkAgent(dagId)
		res, err := usercode.Handle(agent, []byte(req.Req))
		return res, err
	})
}

func handleGetNotify(s *FunchanServer) func(w http.ResponseWriter, r *http.Request) {
	return makeHandler(func(req protocol.GetNotifyReq) ([]byte, error) {
		key := req.Key
		dagId := req.DagId
		keyWithDag := KeyWithDag{
			DagId: dagId,
			Key:   key,
		}
		// obj := s.lookupObjByKey(keyWithDag)
		obj := s.lookupKeyedExport(keyWithDag)
		eh := handle.NewExportHandle(key, req.IP, req.InvId, req.DagId, s.period)
		obj.AddReader(eh, handle.EXPORT_DEFAULT_RANK)
		return []byte(""), nil
	})
}

func handleUpdate(s *FunchanServer) func(w http.ResponseWriter, r *http.Request) {
	return makeHandler(func(req protocol.UpdateReq) ([]byte, error) {
		key := req.Key
		dagId := req.DagId
		from := req.From
		to := req.To
		keyWithDag := KeyWithDag{
			DagId: dagId,
			Key:   key,
		}
		wops := []ddt.WriteOp{}
		// obj := s.lookupObjByKey(keyWithDag).(*handle.ImportHandle)
		obj := s.lookupKeyedImport(keyWithDag).(*handle.ImportHandle)
		for _, wopBytes := range req.Wops {
			wop, err := utils.Unmarshal(wopBytes, ddt.WriteOpMapping)
			if err != nil {
				var dictSetOp pdict.DictSetOp
				err1 := proto.Unmarshal(wopBytes, &dictSetOp)
				if err1 != nil {
					log.Fatalf("Both json unmarshal error: %v, and protobuf unmarshal: %v", err, err1)
				}
				wop = &ddt.DictSetOp{
					K: dictSetOp.Key,
					V: dictSetOp.Value,
				}
			}
			wops = append(wops, wop)
		}
		obj.UpdateRemote(wops, from, to)
		return []byte(""), nil
	})
}

func handleOtherEnd(s *FunchanServer) func(w http.ResponseWriter, r *http.Request) {
	return makeHandler(func(req protocol.EndReq) ([]byte, error) {
		key := req.Key
		dagId := req.DagId
		keyWithDag := KeyWithDag{
			DagId: dagId,
			Key:   key,
		}
		// obj := s.lookupObjByKey(keyWithDag).(*handle.ImportHandle)
		obj := s.lookupKeyedImport(keyWithDag).(*handle.ImportHandle)
		obj.End(handle.EXPORT_DEFAULT_RANK)
		return []byte(""), nil
	})
}

func handleUserRequest(s *FunchanServer, usercode Usercode) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var args []byte

		if r.Body != nil {
			defer r.Body.Close()

			bodyBytes, err := io.ReadAll(r.Body)

			if err != nil {
				log.Fatalf("Read body error: %v", err)
			}

			args = bodyBytes
		}

		dagId := utils.NewUId()
		agent := s.forkAgent(dagId)
		res, err := usercode.Handle(agent, args)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			w.Write(res)
		}
	}
}

func (s *FunchanServer) handleWrite(dWrite ddt.DWriteMsg) {
	obj := s.lookupObj(dWrite.Id)
	if obj.IsWriter() {
		writer := obj.(*handle.WriterHandle)
		writer.Update(dWrite.Wop)
		dWrite.Res <- ddt.DWriteRes{}
	} else {
		panic("User error: write on a Reader")
	}
}

func (s *FunchanServer) handleCons(cons ddt.DConsMsg) {
	objId := utils.NewUId()
	val := ddt.DVAL_CONS[cons.DType]()
	s.objs[objId] = handle.NewWriterHandle(val)
	cons.Res <- ddt.DConsRes{
		Id: objId,
	}
}

func (s *FunchanServer) handleAwait(await ddt.DAwaitMsg) {
	obj := s.lookupObj(await.Id)
	obj.Await()
	obj.ApplyWopBuffer()
	await.Res <- ddt.DAwaitRes{
		Val: obj.GetVal(),
	}
}

func (s *FunchanServer) handleUserEnd(end ddt.DEndMsg) {
	obj := s.lookupObj(end.Id)
	if obj.IsWriter() {
		obj.End(handle.EXPORT_DEFAULT_RANK)
	} else {
		panic("User error: end on a reader")
	}
}

func (s *FunchanServer) handleRead(read ddt.DReadMsg) {
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

func (s *FunchanServer) handleExport(export ddt.ExportMsg) {
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

func (s *FunchanServer) handleImport(importMsg ddt.ImportMsg) {
	// if the import is already present, return the id
	keyWithDag := KeyWithDag{
		Key:   importMsg.Key,
		DagId: importMsg.DagId,
	}
	if id := s.keyedImports[keyWithDag]; id != "" {
		importMsg.Res <- ddt.ImportRes{
			Id: id,
		}
		return
	}

	id := utils.NewUId()
	val := ddt.DVAL_CONS[importMsg.DType]()
	importHandle := handle.NewImportHandle(val, importMsg.Key)
	s.objs[id] = importHandle
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

func (s *FunchanServer) handleTrigger(trigger ddt.TriggerMsg) {
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

func (s *FunchanServer) lookupObj(id string) handle.ObjectHandle {
	obj, ok := s.objs[id]
	if !ok {
		panic("write with unknown id")
	}
	return obj
}

func (s *FunchanServer) lookupKeyedExport(key KeyWithDag) handle.ObjectHandle {
	id, ok := s.keyedExports[key]
	if !ok {
		panic(fmt.Sprintf("looking for export with unknown key: %s", key.Key))
	}
	return s.lookupObj(id)
}

func (s *FunchanServer) lookupKeyedImport(key KeyWithDag) handle.ObjectHandle {
	id, ok := s.keyedImports[key]
	if !ok {
		panic(fmt.Sprintf("looking for import with unknown key: %s", key.Key))
	}
	return s.lookupObj(id)
}
