package datatypes

import "differentiable/utils"

type DType int

const (
	DVEC DType = iota
	DINT
	DDICT
	DANY
)

var DVAL_CONS = utils.MakeMapping(
	func(f func() DVal) DType { return f().GetDType() },
	[]func() DVal{NewDVecVal, NewDIntVal, NewDDictVal, NewDAnyVal},
)

var DOBJ_CONS = func(base DBase) map[DType]DObj {
	return utils.MakeMapping(
		func(dobj DObj) DType { return dobj.GetDType() },
		[]DObj{&DVec{base}, &DInt{base}, &DDict{base}, &DAny{base}},
	)
}

type DObj interface {
	End()
	GetDType() DType
	GetId() string
}

// Each DObj type is associated with a concrete
// DVal type. For example, DVec is associated
// with DVecVal. Since golang does not support
// associated types, developers have to ensure
// the association relationship.
type DVal interface {
	GetDType() DType
	Apply(WriteOp)
	ToWop() WriteOp
}
