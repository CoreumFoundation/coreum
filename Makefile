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

.PHONY: build-znet
build-znet: check-crust-builder
	$(CoreumBuilder) wasm build images

.PHONY: restart-znet
restart-znet: check-crust-builder
	$(CrustBuilder) znet remove
	$(CrustBuilder) znet start --profiles=3cored,ibc --timeout-commit 1s

.PHONY: integration-tests-modules
integration-tests-modules: check-crust-builder
	$(CrustBuilder) znet remove
	$(CoreumBuilder) wasm build images
	$(CrustBuilder) znet start --profiles=3cored --timeout-commit 0.5s
	$(CoreumBuilder) integration-tests-unsafe/modules
	$(CrustBuilder) znet stop
	$(CrustBuilder) znet coverage-convert
	$(CrustBuilder) znet remove

.PHONY: integration-tests-ibc
integration-tests-ibc: check-crust-builder
	$(CrustBuilder) znet remove
	$(CoreumBuilder) build images
	$(CrustBuilder) znet start --profiles=3cored,ibc --timeout-commit 1s
	$(CoreumBuilder) integration-tests-unsafe/ibc
	$(CrustBuilder) znet remove

.PHONY: integration-tests-upgrade
integration-tests-upgrade: check-crust-builder
	$(CrustBuilder) znet remove
	$(CoreumBuilder) wasm build images
	$(CrustBuilder) znet start --cored-version=v3.0.3 --profiles=3cored,ibc --timeout-commit 1s
	$(CoreumBuilder) integration-tests-unsafe/upgrade integration-tests-unsafe/ibc integration-tests-unsafe/modules
	$(CrustBuilder) znet remove

### Helpers go below this line

.PHONY: check-crust-builder
check-crust-builder:
	test -f $(CrustBuilder) || (echo "You need to checkout crust repository" && exit 1)
