package tests

import (
	"differentiable/datatypes"
	"differentiable/server"
	"differentiable/utils"
	"reflect"
	"testing"
)

func TestDVec(t *testing.T) {
	server, stop := server.NewStandaloneServer()
	defer stop()
	agent := server.ForkAgent()
	vec := agent.NewDObj(datatypes.DVEC).(*datatypes.DVec)
	vec.Push(1)
	vec.Push(2)
	vec.End()
	x := vec.Await()

	expected := []any{1, 2}
	if !reflect.DeepEqual(x, expected) {
		t.Fail()
	}
}

func TestMap(t *testing.T) {
	server, stop := server.NewStandaloneServer()
	agent := server.ForkAgent()
	defer stop()

	vec1 := agent.NewDObj(datatypes.DVEC).(*datatypes.DVec)
	vec2 := vec1.Map(func(x any) any { return x.(int) + 10 })
	vec1.Push(1)
	vec1.Push(2)
	vec1.End()
	x := vec2.Await()

	expected := []any{11, 12}
	if !reflect.DeepEqual(x, expected) {
		t.Fail()
	}
}

func TestLength(t *testing.T) {
	server, stop := server.NewStandaloneServer()
	defer stop()
	agent := server.ForkAgent()
	vec := agent.NewDObj(datatypes.DVEC).(*datatypes.DVec)
	l := vec.Length()
	vec.Push(1)
	vec.Push(2)
	vec.Push(3)
	vec.End()
	x := l.Await()

	expected := 3
	if x != expected {
		t.Fail()
	}
}

func TestLengthLTZ(t *testing.T) {
	server, stop := server.NewStandaloneServer()
	defer stop()
	agent := server.ForkAgent()
	vec := agent.NewDObj(datatypes.DVEC).(*datatypes.DVec)
	l := vec.Length()
	vec.Push(1)
	vec.Pop()
	vec.Pop()
	vec.End()
	x := l.Await()

	expected := 0
	if x != expected {
		t.Errorf("length is %v, expected: %d", x, expected)
		t.Fail()
	}
}

func TestFilter(t *testing.T) {
	server, stop := server.NewStandaloneServer()
	defer stop()
	agent := server.ForkAgent()
	vec := agent.NewDObj(datatypes.DVEC).(*datatypes.DVec)
	vec2 := vec.Filter(func(arg any) bool {
		return arg.(int) > 2
	})

	vec.Push(1)
	vec.Push(2)
	vec.Push(3)
	vec.Push(4)

	vec.End()

	x := vec2.Await()
	xe := []any{3, 4}

	if !reflect.DeepEqual(x, xe) {
		t.Errorf("vec looks like: %v", x)
		t.Fail()
	}
}

func TestSumBy(t *testing.T) {
	server, stop := server.NewStandaloneServer()
	defer stop()
	agent := server.ForkAgent()

	type Account struct {
		Name     string
		GetMoney float64
	}

	vec := agent.NewDObj(datatypes.DVEC).(*datatypes.DVec)
	dict := vec.SumBy(func(arg any) (string, float64) {
		account := arg.(Account)
		return account.Name, account.GetMoney
	})

	vec.Push(Account{"Alice", 100})
	vec.Push(Account{"Bob", 200})
	vec.Push(Account{"Alice", 300})
	vec.Push(Account{"Bob", 400})
	vec.Push(Account{"Charlie", 500})

	vec.End()

	dv := dict.Await()
	de := map[string]float64{
		"Alice":   400,
		"Bob":     600,
		"Charlie": 500,
	}

	if !utils.DeepEqual(dv, de) {
		t.Errorf("dict looks like: %v, expected: %v", dv, de)
		t.Fail()
	}
}
