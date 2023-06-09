module agent

replace differentiable => ../../../lib-go

go 1.19

require (
	differentiable v0.0.0-00010101000000-000000000000
	github.com/aws/aws-lambda-go v1.39.1
)

require github.com/google/uuid v1.3.0 // indirect
