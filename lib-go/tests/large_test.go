package tests

import (
	"differentiable/datatypes"
	"differentiable/server"
	"reflect"
	"testing"
)

func TestLarge(t *testing.T) {
	n := 1000
	server, stop := server.NewStandaloneServer()
	agent := server.ForkAgent()
	defer stop()

	vec1 := agent.NewDObj(datatypes.DVEC).(*datatypes.DVec)
	vec2 := vec1.Map(func(x any) any { return x.(int) + n })
	for i := 0; i < n; i++ {
		vec1.Push(i)
	}
	vec1.End()
	x := vec2.Await()

	expected := []any{}
	for i := n; i < n+n; i++ {
		expected = append(expected, i)
	}

	if !reflect.DeepEqual(x, expected) {
		t.Errorf("vec looks like: %v", x)
		t.Fail()
	}
}
