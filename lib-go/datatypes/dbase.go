package datatypes

type DBase struct {
	Id    string
	Proxy AgentProxy
}

func (d *DBase) End() {
	d.Proxy.DEnd(d.Id)
}

func (d *DBase) GetId() string {
	return d.Id
}

type EndOp struct {
}

// wop Non
type NonOp struct{}

func (wop *NonOp) Diff(rop ReadOp, rank int, val DVal) WriteOp {
	return &NonOp{}
}

func (wop *NonOp) Kind() string {
	return "Non"
}
