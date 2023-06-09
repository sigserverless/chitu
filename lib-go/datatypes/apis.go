package datatypes

import "differentiable/utils"

type AgentProxy interface {
	DWrite(id string, wop WriteOp)
	DEnd(id string)
	DAwait(id string) any
	DRead(ids []string, rop ReadOp) string
}

type WriteOp interface {
	utils.Marshalable
	// rank indicates the position of the wop as
	// an argument of the rop. For example, rops
	// such as (DVec)Push are unary, with rank
	// passed as 0 by default.
	// The rop Add(a DInt, b DInt) is binary.
	// rank 0 refers to argument a and rank 1
	// refers to b.
	//
	// val is used when the Diff method is
	// stateful. For example, rop (DVec)Reduce
	// relies on the current value of the rop's result
	// with type DAny.
	// When val is not used, the computing is
	// stateless.
	//
	// TODO: Diff should be a method of ReadOp interface
	// with type (ReadOp) (WriteOp -> int -> DVal) -> WriteOp
	// so that it can be extended by usercode
	Diff(rop ReadOp, rank int, val DVal) WriteOp
}

var WriteOpMapping = utils.MakeMarshalMapping(
	[]WriteOp{
		&NonOp{},
		&PlusOp{},
		&PushOp{}, &PopOp{}, &VecSetOp{}, &AppendOp{},
		&DictSetOp{}, &SetsOp{},
		&ReplaceOp{},
	},
)

type ReadOp interface {
	GetRetDType() DType
}

type DWriteMsg struct {
	Id  string
	Wop WriteOp
	Res chan<- DWriteRes
}

type DWriteRes struct {
}

type DEndMsg struct {
	Id string
}

type DAwaitMsg struct {
	Id  string
	Res chan<- DAwaitRes
}

type DAwaitRes struct {
	Val any
}

type DReadMsg struct {
	Ids []string
	Rop ReadOp
	Res chan<- DReadRes
}

type DReadRes struct {
	ReaderId string
}

type DConsMsg struct {
	DType DType
	Res   chan<- DConsRes
}

type DConsRes struct {
	Id string
}

type ExportMsg struct {
	Id    string
	InvId string
	DagId string
	Key   string
}

type ImportMsg struct {
	Key   string
	DType DType
	InvId string
	DagId string

	Res chan<- ImportRes
}

type ImportRes struct {
	Id string
}

type TriggerMsg struct {
	Fname string
	DagId string
	Args  string
	// Res   chan<- TriggerRes
}
