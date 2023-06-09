package datatypes

type agentProxy struct {
	writeChan chan<- DWriteMsg
	endChan   chan<- DEndMsg
	awaitChan chan<- DAwaitMsg
	readChan  chan<- DReadMsg
}

func NewAgentProxy(
	writeChan chan<- DWriteMsg,
	endChan chan<- DEndMsg,
	awaitChan chan<- DAwaitMsg,
	readChan chan<- DReadMsg,
) AgentProxy {
	return &agentProxy{
		writeChan: writeChan,
		endChan:   endChan,
		awaitChan: awaitChan,
		readChan:  readChan,
	}
}

func (p *agentProxy) DWrite(id string, wop WriteOp) {
	resChan := make(chan DWriteRes, 1)
	p.writeChan <- DWriteMsg{
		id,
		wop,
		resChan,
	}
	<-resChan
}

func (p *agentProxy) DEnd(id string) {
	p.endChan <- DEndMsg{
		Id: id,
	}
}

func (p *agentProxy) DAwait(id string) any {
	resChan := make(chan DAwaitRes, 1)
	p.awaitChan <- DAwaitMsg{
		Id:  id,
		Res: resChan,
	}
	res := <-resChan
	return res.Val
}

func (p *agentProxy) DRead(ids []string, rop ReadOp) string {
	resChan := make(chan DReadRes, 1)
	p.readChan <- DReadMsg{
		Ids: ids,
		Res: resChan,
		Rop: rop,
	}
	res := <-resChan
	return res.ReaderId
}
