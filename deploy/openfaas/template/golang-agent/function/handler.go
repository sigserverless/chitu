package function

import (
	"differentiable/stub"
)

// Handle a function invocation
func Handle(stub stub.AgentStub, req []byte) ([]byte, error) {
	return []byte(string(req) + "hello"), nil
}
