module github.com/caraml-dev/xp/clients

go 1.18

require (
	github.com/caraml-dev/xp/common v0.0.0
	github.com/deepmap/oapi-codegen v1.11.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.8.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/getkin/kin-openapi v0.94.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.4.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/caraml-dev/xp/common => ../common
