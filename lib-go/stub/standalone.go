package stub

import ddt "differentiable/datatypes"

type StandaloneStub struct {
	InvId     string
	consChan  chan<- ddt.DConsMsg
	writeChan chan<- ddt.DWriteMsg
	readChan  chan<- ddt.DReadMsg
	endChan   chan<- ddt.DEndMsg
	awaitChan chan<- ddt.DAwaitMsg
}

func NewStandaloneStub(
	invId string,
	writeChan chan<- ddt.DWriteMsg,
	endChan chan<- ddt.DEndMsg,
	awaitChan chan<- ddt.DAwaitMsg,
	consChan chan<- ddt.DConsMsg,
	readChan chan<- ddt.DReadMsg,
) ObjectCreator {
	return &StandaloneStub{
		InvId:     invId,
		writeChan: writeChan,
		endChan:   endChan,
		awaitChan: awaitChan,
		consChan:  consChan,
		readChan:  readChan,
	}
}

func (stub *StandaloneStub) NewDObj(dtype ddt.DType) ddt.DObj {
	resChan := make(chan ddt.DConsRes, 1)
	stub.consChan <- ddt.DConsMsg{
		DType: dtype,
		Res:   resChan,
	}
	res := <-resChan
	base := ddt.DBase{
		Id: res.Id,
		Proxy: ddt.NewAgentProxy(
			stub.writeChan,
			stub.endChan,
			stub.awaitChan,
			stub.readChan,
		),
	}
	return ddt.DOBJ_CONS(base)[dtype]
}
