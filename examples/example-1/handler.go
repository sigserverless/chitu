package function

import (
	ddt "differentiable/datatypes"
	"differentiable/stub"
	"fmt"
	"strconv"
)

func Handle(stub stub.AgentStub, req []byte) ([]byte, error) {
	n, _ := strconv.Atoi(string(req))
	vec1 := stub.NewDObj(ddt.DVEC).(*ddt.DVec)
	stub.Export("vec1", vec1)
	stub.Trigger("example-2", "")

	vec2 := stub.Import("vec2", ddt.DVEC).(*ddt.DVec)

	for i := 0; i < n; i++ {
		vec1.Push(float64(i))
	}
	vec1.End()

	res := vec2.Await()
	return []byte(fmt.Sprintln(res)), nil
}
