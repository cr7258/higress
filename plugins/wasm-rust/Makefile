PLUGIN_NAME ?= say-hello
REGISTRY ?= higress-registry.cn-hangzhou.cr.aliyuncs.com/plugins/
BUILD_TIME := $(shell date "+%Y%m%d-%H%M%S")
COMMIT_ID := $(shell git rev-parse --short HEAD 2>/dev/null)
IMAGE_TAG = $(if $(strip $(PLUGIN_VERSION)),${PLUGIN_VERSION},${BUILD_TIME}-${COMMIT_ID})
IMG ?= ${REGISTRY}${PLUGIN_NAME}:${IMAGE_TAG}

.DEFAULT:
lint-base:
	cargo fmt --all --check
	cargo clippy --workspace --all-features --all-targets
lint:
	cargo fmt --all --check --manifest-path extensions/${PLUGIN_NAME}/Cargo.toml
	cargo clippy --workspace --all-features --all-targets --manifest-path extensions/${PLUGIN_NAME}/Cargo.toml
build:
	DOCKER_BUILDKIT=1 docker build \
		--build-arg PLUGIN_NAME=${PLUGIN_NAME} \
		-t ${IMG} \
	    --output extensions/${PLUGIN_NAME} \
	    .
	@echo ""
	@echo "output wasm file: extensions/${PLUGIN_NAME}/plugin.wasm"
