package stub

import (
	ddt "differentiable/datatypes"
)

type FunchanStub struct {
	StandaloneStub

	DagId string

	exportChan  chan<- ddt.ExportMsg
	importChan  chan<- ddt.ImportMsg
	triggerChan chan<- ddt.TriggerMsg
}

func NewFunchanStub(
	invId string,
	dagId string,
	writeChan chan<- ddt.DWriteMsg,
	endChan chan<- ddt.DEndMsg,
	awaitChan chan<- ddt.DAwaitMsg,
	consChan chan<- ddt.DConsMsg,
	readChan chan<- ddt.DReadMsg,

	exportChan chan<- ddt.ExportMsg,
	importChan chan<- ddt.ImportMsg,
	triggerChan chan<- ddt.TriggerMsg) AgentStub {
	return &FunchanStub{
		StandaloneStub: StandaloneStub{
			InvId:     invId,
			writeChan: writeChan,
			endChan:   endChan,
			awaitChan: awaitChan,
			consChan:  consChan,
			readChan:  readChan,
		},
		DagId:       dagId,
		exportChan:  exportChan,
		importChan:  importChan,
		triggerChan: triggerChan,
	}
}

func (stub *FunchanStub) Import(key string, dtype ddt.DType) ddt.DObj {
	resChan := make(chan ddt.ImportRes, 1)
	msg := ddt.ImportMsg{
		Key:   key,
		DType: dtype,
		InvId: stub.InvId,
		DagId: stub.DagId,
		Res:   resChan,
	}
	stub.importChan <- msg
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

func (stub *FunchanStub) Export(key string, dobj ddt.DObj) {
	msg := ddt.ExportMsg{
		Key:   key,
		InvId: stub.InvId,
		DagId: stub.DagId,
		Id:    dobj.GetId(),
	}
	stub.exportChan <- msg
}

func (stub *FunchanStub) Trigger(fname string, args string) {
	// resChan := make(chan ddt.TriggerRes, 1)
	msg := ddt.TriggerMsg{
		Fname: fname,
		Args:  args,
		DagId: stub.DagId,
		// Res:   resChan,
	}
	stub.triggerChan <- msg
	// res := <-resChan
	// return res.Val
}

func (stub *FunchanStub) NewDObj(dtype ddt.DType) ddt.DObj {
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
