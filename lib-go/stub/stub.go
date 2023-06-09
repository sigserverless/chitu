package stub

import ddt "differentiable/datatypes"

type ObjectCreator interface {
	NewDObj(ddt.DType) ddt.DObj
}

type AgentStub interface {
	ObjectCreator
	Import(key string, dtype ddt.DType) ddt.DObj
	Export(key string, obj ddt.DObj)
	Trigger(fname string, args string)
}
