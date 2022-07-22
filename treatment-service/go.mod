module github.com/gojek/xp/treatment-service

go 1.16

require (
	cloud.google.com/go/bigquery v1.32.0
	cloud.google.com/go/pubsub v1.21.1
	github.com/confluentinc/confluent-kafka-go v1.8.2
	github.com/deepmap/oapi-codegen v1.11.0
	github.com/getkin/kin-openapi v0.94.0
	github.com/go-chi/chi/v5 v5.0.7
	github.com/gojek/mlp v1.5.3
	github.com/gojek/xp/clients v0.0.0-00010101000000-000000000000
	github.com/gojek/xp/common v0.0.0
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551
	github.com/google/go-cmp v0.5.8
	github.com/google/uuid v1.3.0
	github.com/heptiolabs/healthcheck v0.0.0-20211123025425-613501dd5deb
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.2
	github.com/rs/cors v1.8.2
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.4.0
	github.com/stretchr/testify v1.7.1
	github.com/testcontainers/testcontainers-go v0.13.0
	go.einride.tech/protobuf-bigquery v0.19.0
	go.uber.org/automaxprocs v1.5.1
	golang.org/x/oauth2 v0.0.0-20220411215720-9780585627b5
	google.golang.org/api v0.80.0
	google.golang.org/protobuf v1.28.0
)

replace (
	github.com/gojek/xp/clients => ../clients
	github.com/gojek/xp/common => ../common
	github.com/gojek/xp/management-service => ../management-service
)
