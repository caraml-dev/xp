export

MANAGEMENT_SVC_PATH=management-service
TREATMENT_SVC_PATH=treatment-service
MANAGEMENT_SVC_BIN_NAME=$(if $(MANAGEMENT_APP_NAME),$(MANAGEMENT_APP_NAME),xp-management)
TREATMENT_SVC_BIN_NAME=$(if $(TREATMENT_APP_NAME),$(TREATMENT_APP_NAME),xp-treatment)
VERSION_NUMBER=$(if $(VERSION),$(VERSION),$(shell ./scripts/vertagen/vertagen.sh -f docker))

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	PLATFROM=linux-x86_64
endif
ifeq ($(UNAME_S),Darwin)
	PLATFROM=osx-x86_64
endif

PB_URL="https://github.com/protocolbuffers/protobuf/releases"
PB_VERSION=3.15.8
PROTOC_VERSION=1.5.2
protoc_dir=${PWD}/.protoc

OPENAPI_VERSION=1.8.1

# ==================================
# General
# ==================================

.PHONY: format
format: format-python format-go

.PHONY: lint
lint: lint-python lint-go

.PHONY: vendor
vendor:
	@echo "Fetching dependencies..."
	go mod vendor

.PHONY: version
version:
	$(eval VERSION=$(if $(OVERWRITE_VERSION),$(OVERWRITE_VERSION),v$(shell scripts/vertagen/vertagen.sh)))
	@echo "xp-api version:" $(VERSION)

generate-api:
	test -x ${GOPATH}/bin/oapi-codegen || go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v${OPENAPI_VERSION}
	oapi-codegen -config api/common/schema.conf api/schema.yaml
	oapi-codegen -config api/clients/management.conf api/experiments.yaml
	oapi-codegen -config api/clients/treatment.conf api/treatment.yaml
	oapi-codegen -templates api/templates/chi -config api/management/server.conf api/experiments.yaml
	oapi-codegen -config api/mockmanagement/server.conf api/experiments.yaml
	oapi-codegen -config api/treatment/server.conf api/treatment.yaml
	cd clients/management/ && mockery --name=ClientInterface --output=../../clients/testutils/mocks --filename=ManagementClientInterface.go
	cd clients/treatment/ && mockery --name=TreatmentClientInterface --output=../../clients/testutils/mocks --filename=TreatmentClientInterface.go

# ==================================
# Setup Management & Treatment Services
# ==================================

local-authz-server:
	@docker-compose up -d postgres-auth && docker-compose run keto-server migrate sql -e
	@docker-compose up -d keto-server
	@docker-compose run keto-server-bootstrap-policies engines acp ory policies import glob /policies/example_policy.json

local-db:
	@docker-compose up -d postgres

local-pubsub:
	@docker-compose up -d pubsub

.PHONY: mgmt-svc
mgmt-svc: local-authz-server local-db local-pubsub
	cd management-service && PUBSUB_EMULATOR_HOST="localhost:8085" go run main.go serve --config="config/example.yaml"

.PHONY: treatment-svc
treatment-svc: local-pubsub
	source .env.development
	cd treatment-service && go run main.go serve

swagger-ui:
	@docker-compose up -d swagger-ui
	@xdg-open 2>/dev/null http://localhost:8081 || open http://localhost:8081

$(protoc_dir):
	curl -LO ${PB_URL}/download/v${PB_VERSION}/protoc-${PB_VERSION}-${PLATFROM}.zip
	unzip protoc-${PB_VERSION}-${PLATFROM}.zip -d ${protoc_dir}

compile-protos: | $(protoc_dir)
	go get github.com/golang/protobuf/protoc-gen-go@v${PROTOC_VERSION}
	${protoc_dir}/bin/protoc --proto_path=. -I=api/proto/ --go_out=treatment-service api/proto/logs.proto
	${protoc_dir}/bin/protoc --proto_path=. -I=api/proto/ --go_out=common/segmenters --go_opt=module=github.com/gojek/xp/common/segmenters api/proto/segmenters.proto
	${protoc_dir}/bin/protoc --proto_path=. -I=api/proto/ --go_out=common api/proto/message.proto
	${protoc_dir}/bin/protoc --proto_path=. -I=api/proto/ --go_out=common api/proto/experiment.proto
	${protoc_dir}/bin/protoc --proto_path=. -I=api/proto/ --go_out=common api/proto/settings.proto

# ==================================
# Code dependencies recipes
# ==================================
.PHONY: setup
setup:
	@echo "> Initializing dependencies ..."
	@test -x ${GOPATH}/bin/golangci-lint || go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.40.1

tidy-management-service:
	cd ${MANAGEMENT_SVC_PATH} && go mod tidy

tidy-treatment-service:
	cd ${TREATMENT_SVC_PATH} && go mod tidy

tidy: tidy-management-service tidy-treatment-service

format-go:
	@echo "> Formatting code"
	gofmt -s -w ${MANAGEMENT_SVC_PATH}
	gofmt -s -w ${TREATMENT_SVC_PATH}

# ==================================
# Linting recipes
# ==================================
lint-management-service:
	@echo "> Linting Management Service code..."
	cd ${MANAGEMENT_SVC_PATH} && golangci-lint run --timeout 5m

lint-treatment-service:
	@echo "> Linting Treatment Service code..."
	cd ${TREATMENT_SVC_PATH} && golangci-lint run --timeout 5m

lint-go: lint-management-service lint-treatment-service

# ==================================
# Test recipes
# ==================================
test-management-service: tidy-management-service compile-protos
	@cd ${MANAGEMENT_SVC_PATH} && go mod vendor
	@echo "> Running Management Service tests ..."
	# Set -gcflags=-l to disable inlining for the tests, to have Monkey patching work reliably
	@cd ${MANAGEMENT_SVC_PATH} && go test -v ./... -coverpkg ./... -gcflags=-l -race -coverprofile cover.out.tmp -tags unit,integration ${API_ALL_PACKAGES}
	@cd ${MANAGEMENT_SVC_PATH} && cat cover.out.tmp | grep -v "api/api.go\|cmd\|.pb.go\|mock\|testutils\|server" > cover.out
	@cd ${MANAGEMENT_SVC_PATH} && go tool cover -func cover.out

test-treatment-service: tidy-treatment-service compile-protos
	@cd ${TREATMENT_SVC_PATH} && go mod vendor
	@echo "> Running Fetch Treatment Service tests..."
	cd ${TREATMENT_SVC_PATH} && go test -v ./... -coverpkg ./... -race -coverprofile cover.out.tmp -tags unit,integration ${API_ALL_PACKAGES}
	cd ${TREATMENT_SVC_PATH} && cat cover.out.tmp | grep -v "api/api.go\|cmd\|.pb.go\|testhelper\|server\|internal"  > cover.out
	cd ${TREATMENT_SVC_PATH} && go tool cover -func cover.out

test: test-management-service test-treatment-service

# ==================================
# Build recipes
# ==================================

build-management-service: version
	@echo "Building binary..."
	@cd ${MANAGEMENT_SVC_PATH} && go build -o ./bin/${MANAGEMENT_SVC_BIN_NAME}
	@echo "Copying OpenAPI specs..."
	@cp api/experiments.yaml ${MANAGEMENT_SVC_PATH}/bin
	@cp api/schema.yaml ${MANAGEMENT_SVC_PATH}/bin

build-treatment-service: version
	@echo "Building binary..."
	@cd ${TREATMENT_SVC_PATH} && go build -o ./bin/${TREATMENT_SVC_BIN_NAME}
	@cp api/treatment.yaml ${TREATMENT_SVC_PATH}/bin
	@cp api/schema.yaml ${TREATMENT_SVC_PATH}/bin

build: build-management-service build-treatment-service

.PHONY: build-image
build-image: version
	@$(eval IMAGE_TAG = $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)${BIN_NAME}:${VERSION})
	@echo "Building docker image: ${IMAGE_TAG}"
	docker build --tag ${IMAGE_TAG} .

# ==================================
# Python E2E tests
# ==================================

install-python-ci-dependencies:
	pip install -r tests/requirements.txt

e2e: build
	docker-compose down
	docker-compose up -d postgres pubsub
	cd tests/e2e; python -m pytest -s -v

e2e-ci:
	cd tests/e2e; python -m pytest -s -v

format-python:
	cd tests ; isort e2e/
	cd tests ; flake8 e2e/
	cd tests ; black e2e/

lint-python:
	cd tests ; isort e2e/ --check-only
	cd tests ; flake8 e2e/
	cd tests ; black e2e/ --check
