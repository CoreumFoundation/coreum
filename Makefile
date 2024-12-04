BUILDER = ./bin/coreum-builder

.PHONY: znet
znet:
	$(BUILDER) znet

.PHONY: znet-start
znet-start:
	$(BUILDER) znet start --profiles=3cored

.PHONY: znet-start-ibc
znet-start-ibc:
	$(BUILDER) znet start --profiles=3cored,ibc

.PHONY: znet-start-stress
znet-start-stress:
	$(BUILDER) znet start --profiles=3cored,dex

.PHONY: znet-remove
znet-remove:
	$(BUILDER) znet remove

.PHONY: lint
lint:
	$(BUILDER) lint

.PHONY: test
test:
	$(BUILDER) test

.PHONY: test-fuzz
test-fuzz:
	$(BUILDER) test-fuzz

.PHONY: build
build:
	$(BUILDER) build

.PHONY: images
images:
	$(BUILDER) images

.PHONY: wasm
wasm:
	$(BUILDER) wasm

.PHONY: generate
generate:
	$(BUILDER) generate

.PHONY: release
release:
	$(BUILDER) release

.PHONY: release-images
release-images:
	$(BUILDER) release/images

.PHONY: integration-tests-modules
integration-tests-modules:
	$(BUILDER) integration-tests-unsafe/modules

.PHONY: integration-tests-stress
integration-tests-stress:
	$(BUILDER) integration-tests-unsafe/stress

.PHONY: integration-tests-ibc
integration-tests-ibc:
	$(BUILDER) integration-tests-unsafe/ibc

.PHONY: integration-tests-upgrade
integration-tests-upgrade:
	$(BUILDER) integration-tests-unsafe/upgrade
 