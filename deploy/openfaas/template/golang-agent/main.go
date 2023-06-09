package main

import (
	server "differentiable/server"
	handler "handler/function"
	"time"
)

func main() {
	server.NewFunchanServer(time.Millisecond*20, handler.Handle)
}
