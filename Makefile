CoreumBuilder = ./bin/coreum-builder
CrustBuilder = ../crust/bin/crust

.PHONY: lint
lint:
	$(CoreumBuilder) lint

.PHONY: test
test:
	$(CoreumBuilder) test

.PHONY: build
build:
	$(CoreumBuilder) build

.PHONY: images
images:
	$(CoreumBuilder) images

.PHONY: wasm
wasm:
	$(CoreumBuilder) wasm

.PHONY: generate
generate:
	$(CoreumBuilder) generate

.PHONY: release
release:
	$(CoreumBuilder) release

.PHONY: release-images
release-images:
	$(CoreumBuilder) release/images

.PHONY: dependencies
dependencies: check-crust-builder
	$(CrustBuilder) download
	$(CoreumBuilder) download

.PHONY: integration-tests-modules
integration-tests-modules:
	$(CoreumBuilder) integration-tests-unsafe/modules

.PHONY: integration-tests-ibc
integration-tests-ibc:
	$(CoreumBuilder) integration-tests-unsafe/ibc

.PHONY: integration-tests-upgrade
integration-tests-upgrade:
	$(CoreumBuilder) integration-tests-unsafe/upgrade

### Helpers go below this line

.PHONY: check-crust-builder
check-crust-builder:
	test -f $(CrustBuilder) || (echo "You need to checkout crust repository" && exit 1)
