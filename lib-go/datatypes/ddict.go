package datatypes

import (
	"differentiable/utils"
	"fmt"
)

type DDict struct {
	DBase
}

func (*DDict) GetDType() DType {
	return DDICT
}

func (v *DDict) Await() map[string]any {
	return v.Proxy.DAwait(v.Id).(*DDictVal).Val
}

type DDictVal struct {
	Val map[string]any
	// mergeOperands map[string]map[int]any
}

func NewDDictVal() DVal {
	return &DDictVal{
		Val: map[string]any{},
		// mergeOperands: map[string]map[int]any{},
	}
}

func (v *DDictVal) ToWop() WriteOp {
	return &SetsOp{
		KVs: v.Val,
	}
}

func (*DDictVal) GetDType() DType {
	return DDICT
}

func (v *DDictVal) Apply(op WriteOp) {
	switch op := op.(type) {
	case *NonOp:
		// do nothing
	// case *ReplaceOp:
	// 	v.val = op.Elem.(map[string]any)
	case *DictSetOp:
		v.Val[op.K] = op.V
	case *SetsOp:
		for k1, v1 := range op.KVs {
			v.Val[k1] = v1
		}
	default:
		panic(fmt.Sprintf("unsupported apply operation on DDict: %s", op.Kind()))
	}
}

// wop Set
type DictSetOp struct {
	K string
	V any
}

func (*DictSetOp) Kind() string {
	return "DictSet"
}

func (d *DDict) Set(k string, v any) {
	wop := &DictSetOp{
		K: k,
		V: v,
	}
	d.Proxy.DWrite(d.Id, wop)
}

func (wop *DictSetOp) Diff(rop ReadOp, rank int, val DVal) WriteOp {
	switch rop := rop.(type) {
	case *DictGetOp:
		if wop.K == rop.K {
			return &ReplaceOp{wop.V}
		} else {
			return &NonOp{}
		}
	case *MergeMeanOp:
		cur := val.(*DDictVal).Val
		k := wop.K
		// v := wop.V.([]float64)
		v := utils.ConvertToSlice[float64](wop.V)
		old, ok := cur[k]
		// old := utils.ConvertToSlice[float64](wop.V.([]any))
		if ok {
			return &DictSetOp{
				K: k,
				V: utils.Mean(v, old.([]float64)),
			}
		} else {
			return &DictSetOp{
				K: k,
				V: v,
			}
		}
	case *MergeMeansOp:
		cur := val.(*DDictVal).Val
		k := wop.K
		// v := wop.V.([]float64)
		v := utils.ConvertToSlice[float64](wop.V)
		old, ok := cur[k]
		// old := utils.ConvertToSlice[float64](wop.V.([]any))
		if ok {
			return &DictSetOp{
				K: k,
				V: utils.Sum(old.([]float64), utils.Map(v, func(e float64) float64 { return e / float64(rop.Number) })),
			}
		} else {
			return &DictSetOp{
				K: k,
				V: utils.Map(v, func(e float64) float64 { return e / float64(rop.Number) }),
			}
		}

	// case *MergeOp:
	// 	cur := val.(*DDictVal).val
	// 	operands := val.(*DDictVal).mergeOperands
	// 	return setDiffOnMerge(cur, operands, wop.K, wop.V, rank, rop)
	default:
		panic("Some read operation on Set not implemented. ")
	}
}

// func setDiffOnMerge(cur map[string]any, operands map[string]map[int]any, k string, curV any, rank int, rop *MergeOp) WriteOp {
// 	v, ok := cur[k]
// 	if ok {
// 		oldVal, ok := operands[k][rank]
// 		operands[k][rank] = curV
// 		if ok {
// 			newVal := rop.F(curV, rop.Rev(v, oldVal))
// 			return &DictSetOp{k, newVal}
// 		} else {
// 			return &DictSetOp{k, rop.F(v, curV)}
// 		}
// 	} else {
// 		operands[k] = map[int]any{}
// 		operands[k][rank] = curV
// 		return &DictSetOp{k, curV}
// 	}
// }

// wop Sets
// wop Append
type SetsOp struct {
	KVs map[string]any
}

func (*SetsOp) Kind() string {
	return "Sets"
}

func (v *DDict) Sets(kvs map[string]any) {
	wop := &SetsOp{
		KVs: kvs,
	}
	v.Proxy.DWrite(v.Id, wop)
}

func (wop *SetsOp) Diff(rop ReadOp, rank int, val DVal) WriteOp {
	switch rop := rop.(type) {
	case *DictGetOp:
		newVal, ok := wop.KVs[rop.K]
		if ok {
			return &ReplaceOp{newVal}
		} else {
			return &NonOp{}
		}
	case *MergeMeanOp:
		cur := val.(*DDictVal).Val
		newKVs := map[string]any{}
		for k, v := range wop.KVs {
			old, ok := cur[k]
			if ok {
				newKVs[k] = utils.Mean(v.([]float64), old.([]float64))
			} else {
				newKVs[k] = v
			}
		}
		return &SetsOp{newKVs}
	case *MergeMeansOp:
		cur := val.(*DDictVal).Val
		newKVs := map[string]any{}
		for k, v := range wop.KVs {
			old, ok := cur[k]
			if ok {
				newKVs[k] = utils.Sum(old.([]float64), utils.Map(v.([]float64), func(e float64) float64 { return e / float64(rop.Number) }))
			} else {
				newKVs[k] = utils.Map(v.([]float64), func(e float64) float64 { return e / float64(rop.Number) })
			}
		}
		return &SetsOp{newKVs}
	// TODO:
	// case *MergeOp:
	// 	cur := val.(*DDictVal).val
	// 	newMap
	// 	for k, v := range wop.KVs {
	// 		old, ok := cur[k]
	// 		if ok {
	// 			return &DictSetOp{wop.K, rop.F(v, wop.V)}
	// 		} else {
	// 			return &DictSetOp{wop.K, wop.V}
	// 		}
	// 	}
	default:
		panic("Some read operation on Sets not implemented. ")
	}
}

// rop Get
type DictGetOp struct {
	K string
}

func (op *DictGetOp) GetRetDType() DType {
	return DANY
}

func (v *DDict) Get(key string) *DAny {
	rop := &DictGetOp{
		K: key,
	}
	id := v.Proxy.DRead([]string{v.Id}, rop)
	return &DAny{
		DBase{
			Id:    id,
			Proxy: v.Proxy,
		},
	}
}

// rop MergeMean
type MergeMeanOp struct {
}

func (*MergeMeanOp) GetRetDType() DType {
	return DDICT
}

func MergeMean(a, b *DDict) *DDict {
	rop := &MergeMeanOp{}
	ids := []string{a.Id, b.Id}
	id := a.Proxy.DRead(ids, rop)
	return &DDict{
		DBase{
			Id:    id,
			Proxy: a.Proxy,
		},
	}
}

// rop MergeMeans
type MergeMeansOp struct {
	Number int
}

func (*MergeMeansOp) GetRetDType() DType {
	return DDICT
}

func MergeMeans(dicts []*DDict) *DDict {
	rop := &MergeMeansOp{Number: len(dicts)}
	ids := []string{}
	for _, d := range dicts {
		ids = append(ids, d.Id)
	}
	id := dicts[0].Proxy.DRead(ids, rop)
	return &DDict{
		DBase{
			Id:    id,
			Proxy: dicts[0].Proxy,
		},
	}
}
