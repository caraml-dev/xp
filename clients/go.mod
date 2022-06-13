module github.com/gojek/turing-experiments/clients

go 1.16

require (
	github.com/deepmap/oapi-codegen v1.8.2
	github.com/gojek/turing-experiments/common v0.0.0
	github.com/pkg/errors v0.9.1
)

replace github.com/gojek/turing-experiments/common => ../common
