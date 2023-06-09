package datatypes

import (
	"differentiable/utils"
	"fmt"
)

// DVec
type DVec struct {
	DBase
}

func (*DVec) GetDType() DType {
	return DVEC
}

func (v *DVec) Await() []any {
	return v.Proxy.DAwait(v.Id).(*DVecVal).Val
}

// DVecVal
type DVecVal struct {
	Val []any
}

func NewDVecVal() DVal {
	return &DVecVal{
		Val: []any{},
	}
}

func (v *DVecVal) ToWop() WriteOp {
	return &AppendOp{
		Elems: v.Val,
	}
}

func (*DVecVal) GetDType() DType {
	return DVEC
}

func (v *DVecVal) Apply(op WriteOp) {
	switch op := op.(type) {
	case *NonOp:
		// do nothing
	// case *ReplaceOp:
	// 	v.val = op.Elem.([]any)
	case *PushOp:
		v.Val = append(v.Val, op.Elem)
	case *PopOp:
		l := len(v.Val)
		if l != 0 {
			v.Val = v.Val[:l-1]
		}
	case *VecSetOp:
		v.Val[op.Index] = op.Elem
	case *AppendOp:
		v.Val = append(v.Val, op.Elems...)
	default:
		panic(fmt.Sprintf("unsupported apply operation on DVec: %s", op.Kind()))
	}
}

// wop Push
type PushOp struct {
	Elem any
}

func (*PushOp) Kind() string {
	return "Push"
}

func (v *DVec) Push(e any) {
	wop := &PushOp{
		Elem: e,
	}
	v.Proxy.DWrite(v.Id, wop)
}

func (wop *PushOp) Diff(rop ReadOp, rank int, val DVal) WriteOp {
	switch rop := rop.(type) {
	case *LengthOp:
		return &PlusOp{1}
	case *MapOp:
		return &PushOp{
			Elem: rop.F(wop.Elem),
		}
	case *FilterOp:
		if rop.F(wop.Elem) {
			return &PushOp{
				Elem: wop.Elem,
			}
		} else {
			return &NonOp{}
		}
	case *SumByOp:
		kvs := val.(*DDictVal).Val
		k, v := rop.F(wop.Elem)
		old, ok := kvs[k]
		if ok {
			return &DictSetOp{
				K: k,
				V: old.(float64) + v,
			}
		} else {
			return &DictSetOp{
				K: k,
				V: v,
			}
		}
	case *CountOp:
		kvs := val.(*DDictVal).Val
		k := wop.Elem.(string)
		old, ok := kvs[k]
		if ok {
			return &DictSetOp{
				K: k,
				V: old.(int) + 1,
			}
		} else {
			return &DictSetOp{
				K: k,
				V: 1,
			}
		}
	default:
		panic("Some read operation on Push not implemented. ")
	}
}

// wop Append
type AppendOp struct {
	Elems []any
}

func (*AppendOp) Kind() string {
	return "Append"
}

func (v *DVec) Append(es []any) {
	wop := &AppendOp{
		Elems: es,
	}
	v.Proxy.DWrite(v.Id, wop)
}

func (wop *AppendOp) Diff(rop ReadOp, rank int, val DVal) WriteOp {
	switch rop := rop.(type) {
	case *LengthOp:
		return &PlusOp{len(wop.Elems)}
	case *MapOp:
		return &AppendOp{
			Elems: utils.Map(wop.Elems, rop.F),
		}
	case *FilterOp:
		newVec := []any{}
		for _, e := range wop.Elems {
			if rop.F(e) {
				newVec = append(newVec, e)
			}
		}
		return &AppendOp{
			Elems: newVec,
		}
	// case *GetOp:
	// TODO:
	case *SumByOp:
		kvs := val.(*DDictVal).Val
		newDict := map[string]any{}
		for k, v := range kvs {
			newDict[k] = v.(float64)
		}
		for _, e := range wop.Elems {
			k, v := rop.F(e)
			old, ok := newDict[k]
			if ok {
				newDict[k] = old.(float64) + v
			} else {
				newDict[k] = v
			}
		}
		return &SetsOp{
			newDict,
		}
	case *CountOp:
		kvs := val.(*DDictVal).Val
		newDict := map[string]any{}
		for k, v := range kvs {
			newDict[k] = v
		}
		for _, e := range wop.Elems {
			k := e.(string)
			old, ok := newDict[k]
			if ok {
				newDict[k] = old.(int) + 1
			} else {
				newDict[k] = 1
			}
		}
		return &SetsOp{
			newDict,
		}
	default:
		panic("Some read operation on Append not implemented. ")
	}
}

// wop Pop
type PopOp struct{}

func (*PopOp) Kind() string {
	return "Pop"
}

func (v *DVec) Pop() {
	wop := &PopOp{}
	v.Proxy.DWrite(v.Id, wop)
}

func (wop *PopOp) Diff(rop ReadOp, rank int, val DVal) WriteOp {
	switch rop := rop.(type) {
	case *LengthOp:
		l := val.(*DIntVal).Val
		if l == 0 {
			return &NonOp{}
		} else {
			return &PlusOp{-1}
		}
	case *MapOp:
		return &PopOp{}
	case *FilterOp:
		vecVal := val.(*DVecVal)
		if rop.F(vecVal.Val[len(vecVal.Val)-1]) {
			return &PopOp{}
		} else {
			return &NonOp{}
		}
	default:
		panic("Some read operation on Pop not implemented. ")
	}
}

// wop Set
type VecSetOp struct {
	Index int
	Elem  any
}

func (*VecSetOp) Kind() string {
	return "VecSet"
}

func (v *DVec) Set(i int, e any) {
	wop := &VecSetOp{
		Index: i,
		Elem:  e,
	}
	v.Proxy.DWrite(v.Id, wop)
}

func (wop *VecSetOp) Diff(rop ReadOp, rank int, val DVal) WriteOp {
	switch rop := rop.(type) {
	case *LengthOp:
		return &NonOp{}
	case *MapOp:
		return &VecSetOp{
			Index: wop.Index,
			Elem:  rop.F(wop.Elem),
		}
	default:
		panic("Some read operation on Set not implemented. ")
	}
}

// rop Length
type LengthOp struct{}

func (op *LengthOp) GetRetDType() DType {
	return DINT
}

func (v *DVec) Length() *DInt {
	rop := &LengthOp{}
	id := v.Proxy.DRead([]string{v.Id}, rop)
	return &DInt{
		DBase{
			Id:    id,
			Proxy: v.Proxy,
		},
	}
}

// rop Filter
type FilterOp struct {
	F func(any) bool
}

func (op *FilterOp) GetRetDType() DType {
	return DVEC
}

func (v *DVec) Filter(f func(any) bool) *DVec {
	rop := &FilterOp{f}
	id := v.Proxy.DRead([]string{v.Id}, rop)
	return &DVec{
		DBase{
			Id:    id,
			Proxy: v.Proxy,
		},
	}
}

// rop Map
type MapOp struct {
	F func(any) any
}

func (op *MapOp) GetRetDType() DType {
	return DVEC
}

func (v *DVec) Map(f func(any) any) *DVec {
	rop := &MapOp{
		F: f,
	}
	id := v.Proxy.DRead([]string{v.Id}, rop)
	return &DVec{
		DBase{
			Id:    id,
			Proxy: v.Proxy,
		},
	}
}

// rop Count
type CountOp struct {
}

func (op *CountOp) GetRetDType() DType {
	return DDICT
}

func Count(wordlist []*DVec) *DDict {
	rop := &CountOp{}
	v := wordlist[0]
	ids := []string{}
	for _, w := range wordlist {
		ids = append(ids, w.Id)
	}
	id := v.Proxy.DRead(ids, rop)
	return &DDict{
		DBase{
			Id:    id,
			Proxy: v.Proxy,
		},
	}
}

// rop SumBy
type SumByOp struct {
	F func(any) (string, float64)
}

func (op *SumByOp) GetRetDType() DType {
	return DDICT
}

func (v *DVec) SumBy(f func(any) (string, float64)) *DDict {
	rop := &SumByOp{f}
	id := v.Proxy.DRead([]string{v.Id}, rop)
	return &DDict{
		DBase{
			Id:    id,
			Proxy: v.Proxy,
		},
	}
}
