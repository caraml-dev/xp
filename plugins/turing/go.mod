module github.com/gojek/xp/plugins/turing

go 1.16

require (
	bou.ke/monkey v1.0.2
	github.com/go-playground/validator/v10 v10.4.1
	github.com/gojek/turing/engines/experiment v0.0.0-20220719121551-2a7eb84e97b4
	github.com/gojek/xp/clients v0.0.0-00010101000000-000000000000
	github.com/gojek/xp/common v0.0.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.19.1
	golang.org/x/oauth2 v0.0.0-20201208152858-08078c50e5b5
)

replace (
	github.com/gojek/xp/clients => ../../clients
	github.com/gojek/xp/common => ../../common
)
