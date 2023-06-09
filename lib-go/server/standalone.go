package server

import (
	"context"
	ddt "differentiable/datatypes"
	"differentiable/stub"
	"differentiable/stub/handle"
	"differentiable/utils"
)

type StandaloneServer struct {
	objs      map[string]handle.ObjectHandle
	writeChan chan ddt.DWriteMsg
	endChan   chan ddt.DEndMsg
	consChan  chan ddt.DConsMsg
	awaitChan chan ddt.DAwaitMsg
	readChan  chan ddt.DReadMsg
}

func NewStandaloneServer() (*StandaloneServer, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	s := &StandaloneServer{
		objs:      map[string]handle.ObjectHandle{},
		consChan:  make(chan ddt.DConsMsg, CHANNEL_BUFFER_SIZE),
		writeChan: make(chan ddt.DWriteMsg, CHANNEL_BUFFER_SIZE),
		endChan:   make(chan ddt.DEndMsg, CHANNEL_BUFFER_SIZE),
		awaitChan: make(chan ddt.DAwaitMsg, CHANNEL_BUFFER_SIZE),
		readChan:  make(chan ddt.DReadMsg, CHANNEL_BUFFER_SIZE),
	}
	go s.mainLoop(ctx)
	return s, cancel
}

func (s *StandaloneServer) mainLoop(ctx context.Context) {
	for {
		select {
		case write := <-s.writeChan:
			s.handleWrite(write)
		case cons := <-s.consChan:
			s.handleCons(cons)
		case end := <-s.endChan:
			s.handleEnd(end)
		case await := <-s.awaitChan:
			go s.handleAwait(await)
		case read := <-s.readChan:
			s.handleRead(read)
		case <-ctx.Done():
			return
		}
	}
}

func (s *StandaloneServer) handleWrite(dWrite ddt.DWriteMsg) {
	obj := s.lookupObj(dWrite.Id)
	if obj.IsWriter() {
		writer := obj.(*handle.WriterHandle)
		dWrite.Res <- ddt.DWriteRes{}
		writer.Update(dWrite.Wop)
	} else {
		panic("User error: write on a Reader")
	}
}

func (s *StandaloneServer) handleCons(cons ddt.DConsMsg) {
	objId := utils.NewUId()
	val := ddt.DVAL_CONS[cons.DType]()
	s.objs[objId] = handle.NewWriterHandle(val)
	cons.Res <- ddt.DConsRes{
		Id: objId,
	}
}

func (s *StandaloneServer) handleAwait(await ddt.DAwaitMsg) {
	obj := s.lookupObj(await.Id)
	obj.Await()
	obj.ApplyWopBuffer()
	await.Res <- ddt.DAwaitRes{
		Val: obj.GetVal(),
	}
}

func (s *StandaloneServer) handleEnd(end ddt.DEndMsg) {
	obj := s.lookupObj(end.Id)
	if obj.IsWriter() {
		obj.End(handle.EXPORT_DEFAULT_RANK)
	} else {
		panic("User error: end on a reader")
	}
}

func (s *StandaloneServer) handleRead(read ddt.DReadMsg) {
	readerId := utils.NewUId()
	val := ddt.DVAL_CONS[read.Rop.GetRetDType()]()
	writersCount := len(read.Ids)
	readerObj := handle.NewReaderHandle(val, read.Rop, writersCount)
	s.objs[readerId] = readerObj
	for rank, id := range read.Ids {
		writer := s.lookupObj(id)
		writer.AddReader(readerObj, rank)
	}
	read.Res <- ddt.DReadRes{
		ReaderId: readerId,
	}
}

func (s *StandaloneServer) ForkAgent() stub.ObjectCreator {
	invId := utils.NewUId()
	return stub.NewStandaloneStub(
		invId,
		s.writeChan,
		s.endChan,
		s.awaitChan,
		s.consChan,
		s.readChan,
	)
}

func (s *StandaloneServer) lookupObj(id string) handle.ObjectHandle {
	obj, ok := s.objs[id]
	if !ok {
		panic("write with unknown id")
	}
	return obj
}
