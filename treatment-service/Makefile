export

TREATMENT_SVC_BIN_NAME=$(if $(TREATMENT_APP_NAME),$(TREATMENT_APP_NAME),xp-treatment)

# ==================================
# General
# ==================================

.PHONY: vendor
vendor:
	@echo "Fetching dependencies..."
	go mod vendor

.PHONY: version
version:
	$(eval VERSION=$(if $(OVERWRITE_VERSION),$(OVERWRITE_VERSION),v$(shell ../scripts/vertagen/vertagen.sh)))
	@echo "xp-api version:" $(VERSION)

# ==================================
# Build recipes
# ==================================

build-treatment-service: version
	@echo "Building binary..."
	@go build -o ./bin/${TREATMENT_SVC_BIN_NAME}
	@cp api/treatment.yaml ${TREATMENT_SVC_PATH}/bin
	@cp api/schema.yaml ${TREATMENT_SVC_PATH}/bin

.PHONY: build-image
build-image: vendor version
	@$(eval IMAGE_TAG = $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)${BIN_NAME}:${VERSION})
	@echo "Building docker image: ${IMAGE_TAG}"
	docker build --tag ${IMAGE_TAG} .
