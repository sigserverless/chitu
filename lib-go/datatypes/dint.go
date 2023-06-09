package datatypes

import "fmt"

type DInt struct {
	DBase
}

func (*DInt) GetDType() DType {
	return DINT
}

type DIntVal struct {
	Val int
}

func NewDIntVal() DVal {
	return &DIntVal{
		Val: 0,
	}
}

func (v *DIntVal) ToWop() WriteOp {
	return &PlusOp{
		Operand: v.Val,
	}
}

func (*DIntVal) GetDType() DType {
	return DINT
}

func (d *DInt) Await() int {
	return d.Proxy.DAwait(d.Id).(*DIntVal).Val
}

// wop Plus
type PlusOp struct {
	Operand int
}

func (*PlusOp) Kind() string {
	return "Plus"
}

func (d *DInt) Plus(operand int) {
	wop := &PlusOp{
		Operand: operand,
	}
	d.Proxy.DWrite(d.Id, wop)
}

func (d *DIntVal) Apply(op WriteOp) {
	switch op := op.(type) {
	case *NonOp:
		// do nothing
	// case *ReplaceOp:
	// 	d.val = op.Elem.(int)
	case *PlusOp:
		d.Val = d.Val + op.Operand
	default:
		panic(fmt.Sprintf("unsupported apply operation on DInt: %s", op.Kind()))
	}
}

func (wop *PlusOp) Diff(rop ReadOp, rank int, val DVal) WriteOp {
	switch rop.(type) {
	case *AddOp:
		return &PlusOp{wop.Operand}
	default:
		panic("Some read operation on Plus not implemented. ")
	}
}

// rop Add
type AddOp struct{}

func (*AddOp) GetRetDType() DType {
	return DINT
}

func Add(a, b *DInt) *DInt {
	rop := &AddOp{}
	ids := []string{a.Id, b.Id}
	id := a.Proxy.DRead(ids, rop)
	return &DInt{
		DBase{
			Id:    id,
			Proxy: a.Proxy,
		},
	}
}
