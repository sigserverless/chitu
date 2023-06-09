package tests

import (
	ddt "differentiable/datatypes"
	"differentiable/server"
	"reflect"
	"testing"
)

func TestDInt(t *testing.T) {
	server, stop := server.NewStandaloneServer()
	defer stop()
	agent := server.ForkAgent()
	i := agent.NewDObj(ddt.DINT).(*ddt.DInt)
	i.Plus(10)
	i.Plus(20)
	i.End()
	x := i.Await()

	expected := 30
	if !reflect.DeepEqual(x, expected) {
		t.Fail()
	}
}

func TestAdd(t *testing.T) {
	server, stop := server.NewStandaloneServer()
	defer stop()
	agent := server.ForkAgent()
	a := agent.NewDObj(ddt.DINT).(*ddt.DInt)
	b := agent.NewDObj(ddt.DINT).(*ddt.DInt)
	c := ddt.Add(a, b)
	a.Plus(10)
	b.Plus(20)
	a.Plus(1)
	b.Plus(2)
	a.End()
	b.End()
	x := c.Await()

	expected := 33
	if !reflect.DeepEqual(x, expected) {
		t.Fail()
	}
}
