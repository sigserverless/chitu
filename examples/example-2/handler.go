package function

import (
	ddt "differentiable/datatypes"
	"differentiable/stub"
)

// Handle a function invocation
func Handle(stub stub.AgentStub, req []byte) ([]byte, error) {
	vec1 := stub.Import("vec1", ddt.DVEC).(*ddt.DVec)
	vec2 := vec1.Map(func(x any) any { return x.(float64) + 5 })
	stub.Export("vec2", vec2)

	_ = vec2.Await()
	return []byte(""), nil
}
