# ==================================
# General
# ==================================

.PHONY: version
version:
	$(eval VERSION=$(if $(OVERWRITE_VERSION),$(OVERWRITE_VERSION),v$(shell ../../scripts/vertagen/vertagen.sh -f docker)))
	@echo "version:" $(VERSION)

.PHONY: vendor
vendor:
	@echo "Fetching dependencies..."
	go mod vendor

# ==================================
# Code dependencies recipes
# ==================================

tidy-plugin:
	go mod tidy

# ==================================
# Build recipes
# ==================================

.PHONY: build
build:
	@echo "Building binary..."
	go build -o ./bin/plugin

.PHONY: build-image
build-image: vendor version
	@$(eval IMAGE_TAG = $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)xp-engine-plugin:${VERSION})
	@echo "Building docker image: ${IMAGE_TAG}"
	docker build . --tag ${IMAGE_TAG}

# ==================================
# Test recipes
# ==================================

test-plugin: tidy-plugin vendor
	@echo "> Running Turing Experiment Engine XP Plugin tests ..."
	# Set -gcflags=-l to disable inlining for the tests, to have Monkey patching work reliably
	go test -v ./... -coverpkg ./... -gcflags=-l -race -coverprofile cover.out.tmp -tags unit,integration
	cat cover.out.tmp | grep -v "cmd\|testutils" > cover.out
	go tool cover -func cover.out