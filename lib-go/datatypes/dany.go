package datatypes

import "fmt"

type DAny struct {
	DBase
}

func (*DAny) GetDType() DType {
	return DANY
}

func (v *DAny) Await() any {
	return v.Proxy.DAwait(v.Id).(*DAnyVal).Val
}

type DAnyVal struct {
	Val any
}

func NewDAnyVal() DVal {
	return &DAnyVal{}
}

func (*DAnyVal) GetDType() DType {
	return DANY
}

func (v *DAnyVal) ToWop() WriteOp {
	return &ReplaceOp{
		Elem: v.Val,
	}
}

func (v *DAnyVal) Apply(op WriteOp) {
	switch op := op.(type) {
	case *NonOp:
		// do nothing
	case *ReplaceOp:
		v.Val = op.Elem
	default:
		panic(fmt.Sprintf("unsupported apply operation on DAny: %s", op.Kind()))
	}
}

// wop Replace
type ReplaceOp struct {
	Elem any
}

func (*ReplaceOp) Kind() string {
	return "Replace"
}

func (wop *ReplaceOp) Diff(rop ReadOp, rank int, val DVal) WriteOp {
	panic("No read operation on DAny.Replace implemented")
}

func (a *DAny) Replace(e any) {
	wop := &ReplaceOp{
		Elem: e,
	}
	a.Proxy.DWrite(a.Id, wop)
}
