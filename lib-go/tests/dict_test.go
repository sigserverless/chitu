package tests

import (
	ddt "differentiable/datatypes"
	"differentiable/server"
	"differentiable/utils"
	"testing"
)

func TestDDict(t *testing.T) {
	server, stop := server.NewStandaloneServer()
	defer stop()
	agent := server.ForkAgent()
	dict := agent.NewDObj(ddt.DDICT).(*ddt.DDict)
	x := dict.Get("x")
	y := dict.Get("y")
	dict.Set("x", 10)
	dict.Set("x", 20)
	dict.Set("z", 30)

	dict.End()

	xv := x.Await().(int)
	yv := y.Await()

	xe := 20
	var ye any = nil
	if xv != xe || yv != ye {
		t.Fail()
	}
}

func TestMergeMean(t *testing.T) {
	server, stop := server.NewStandaloneServer()
	defer stop()
	agent := server.ForkAgent()
	x := agent.NewDObj(ddt.DDICT).(*ddt.DDict)
	y := agent.NewDObj(ddt.DDICT).(*ddt.DDict)
	z := ddt.MergeMean(x, y)
	x.Set("a", []float64{1.0})
	y.Set("a", []float64{3.0})

	x.Set("b", []float64{1.0, 2.0, 3.0})
	y.Set("b", []float64{3.0, 4.0, 5.0})

	x.End()
	y.End()

	zv := z.Await()

	ze := map[string][]float64{
		"a": {2.0},
		"b": {2.0, 3.0, 4.0},
	}

	if !utils.DeepEqual(zv, ze) {
		t.Errorf("result is %v, expected: %v", zv, ze)
		t.Fail()
	}
}

func TestMergeMeans(t *testing.T) {
	server, stop := server.NewStandaloneServer()
	defer stop()
	agent := server.ForkAgent()
	x := agent.NewDObj(ddt.DDICT).(*ddt.DDict)
	y := agent.NewDObj(ddt.DDICT).(*ddt.DDict)
	z := agent.NewDObj(ddt.DDICT).(*ddt.DDict)
	w := ddt.MergeMeans([]*ddt.DDict{x, y, z})
	x.Set("a", []float64{1.0})
	y.Set("a", []float64{3.0})
	z.Set("a", []float64{5.0})

	x.Set("b", []float64{1.0, 2.0, 3.0})
	y.Set("b", []float64{3.0, 4.0, 5.0})
	z.Set("b", []float64{5.0, 6.0, 7.0})

	x.End()
	y.End()
	z.End()

	wv := w.Await()

	we := map[string][]float64{
		"a": {3.0},
		"b": {3.0, 4.0, 5.0},
	}

	if !utils.DeepEqual(wv, we) {
		t.Errorf("result is %v, expected: %v", wv, we)
		t.Fail()
	}
}

func TestCount(t *testing.T) {
	server, stop := server.NewStandaloneServer()
	defer stop()
	agent := server.ForkAgent()
	w1 := agent.NewDObj(ddt.DVEC).(*ddt.DVec)
	w2 := agent.NewDObj(ddt.DVEC).(*ddt.DVec)
	w3 := agent.NewDObj(ddt.DVEC).(*ddt.DVec)

	c := ddt.Count([]*ddt.DVec{w1, w2, w3})
	w1.Append([]any{"a", "b", "c", "a"})
	w2.Append([]any{"a", "b", "d"})
	w3.Append([]any{"c", "d", "e"})

	w1.Push("a")
	w2.Push("f")
	w3.Push("fk")
	w1.End()
	w2.End()
	w3.End()

	cv := c.Await()
	ce := map[any]int{
		"a":  4,
		"b":  2,
		"c":  2,
		"d":  2,
		"e":  1,
		"f":  1,
		"fk": 1,
	}

	if !utils.DeepEqual(cv, ce) {
		t.Errorf("result is %v, expected: %v", cv, ce)
		t.Fail()
	}
}
