module github.com/gojek/xp/management-service

go 1.16

require (
	bou.ke/monkey v1.0.2
	cloud.google.com/go/pubsub v1.3.1
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/deepmap/oapi-codegen v1.8.2
	github.com/getkin/kin-openapi v0.75.0
	github.com/go-chi/chi/v5 v5.0.3
	github.com/go-playground/validator/v10 v10.9.0
	github.com/gojek/mlp v1.4.9
	github.com/gojek/xp/common v0.0.0
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/golang-migrate/migrate/v4 v4.14.1
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551
	github.com/google/go-cmp v0.5.6
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/jinzhu/gorm v1.9.16
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/rs/cors v1.7.0
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/testcontainers/testcontainers-go v0.11.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/api v0.30.0
	google.golang.org/protobuf v1.26.0-rc.1
)

replace (
	github.com/gojek/xp/clients => ../clients
	github.com/gojek/xp/common => ../common
	github.com/gojek/xp/management-service => ./
)
