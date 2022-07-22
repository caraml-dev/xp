module github.com/gojek/xp/clients

go 1.16

require (
	github.com/deepmap/oapi-codegen v1.8.2
	github.com/gojek/xp/common v0.0.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
)

replace github.com/gojek/xp/common => ../common
