package main

import (
	"fmt"

	ddt "differentiable/datatypes"
	server "differentiable/server"

	lmd "github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lmd.Start(HandleRequest)
}

func HandleRequest() (string, error) {
	server, stop := server.NewStandaloneServer()
	defer stop()
	agent := server.ForkAgent()
	x := agent.NewDObj(ddt.DDICT).(*ddt.DDict)
	y := agent.NewDObj(ddt.DDICT).(*ddt.DDict)
	z := agent.NewDObj(ddt.DDICT).(*ddt.DDict)
	w := ddt.MergeMeans([]*ddt.DDict{x, y, z})
	x.Set("a", []any{1.0})
	y.Set("a", []any{3.0})
	z.Set("a", []any{5.0})

	x.Set("b", []any{1.0, 2.0, 3.0})
	y.Set("b", []any{3.0, 4.0, 5.0})
	z.Set("b", []any{5.0, 6.0, 7.0})

	x.End()
	y.End()
	z.End()

	wv := w.Await()

	we := map[string]any{
		"a": []any{3.0},
		"b": []any{3.0, 4.0, 5.0},
	}

	return fmt.Sprintf("result is %v, expected: %v", wv, we), nil
}
