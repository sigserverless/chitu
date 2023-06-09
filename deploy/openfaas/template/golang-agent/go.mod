module handler

go 1.19

replace handler/function => ./function

replace differentiable => ./lib-go

require (
	differentiable v0.0.0-00010101000000-000000000000
	handler/function v0.0.0-00010101000000-000000000000
)

require (
	github.com/google/uuid v1.3.0 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
)
