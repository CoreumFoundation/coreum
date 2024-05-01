.PHONY: lint
lint:
	./bin/coreum-builder lint

.PHONY: test
test:
	./bin/coreum-builder test

.PHONY: build
build:
	./bin/coreum-builder build

.PHONY: images
images:
	./bin/coreum-builder images

.PHONY: wasm
wasm:
	./bin/coreum-builder wasm

.PHONY: generate
generate:
	./bin/coreum-builder generate

.PHONY: release
release:
	./bin/coreum-builder release

.PHONY: release-images
release-images:
	./bin/coreum-builder release/images

.PHONY: dependencies
dependencies: check-crust-builder check-faucet-builder
	../crust/bin/crust download
	../faucet/bin/faucet-builder download
	./bin/coreum-builder download

.PHONY: integration-tests-modules
integration-tests-modules: check-crust-builder
	../crust/bin/crust znet remove
	./bin/coreum-builder wasm build images
	../crust/bin/crust znet start --profiles=3cored --timeout-commit 0.5s
	./bin/coreum-builder integration-tests-unsafe/modules
	../crust/bin/crust znet stop
	../crust/bin/crust znet coverage-convert
	../crust/bin/crust znet remove

.PHONY: integration-tests-ibc
integration-tests-ibc: check-crust-builder
	../crust/bin/crust znet remove
	./bin/coreum-builder build images
	../crust/bin/crust znet start --profiles=3cored,ibc --timeout-commit 1s
	./bin/coreum-builder integration-tests-unsafe/ibc
	../crust/bin/crust znet remove

.PHONY: integration-tests-faucet
integration-tests-faucet: check-crust-builder check-faucet-builder
	../crust/bin/crust znet remove
	./bin/coreum-builder build images
	../faucet/bin/faucet-builder images
	../crust/bin/crust znet start --profiles=faucet --timeout-commit 0.5s
	../faucet/bin/faucet-builder integration-tests
	../crust/bin/crust znet remove

.PHONY: integration-tests-upgrade
integration-tests-upgrade: check-crust-builder
	../crust/bin/crust znet remove
	./bin/coreum-builder wasm build images
	../crust/bin/crust znet start --cored-version=v3.0.3 --profiles=3cored,ibc --timeout-commit 1s
	./bin/coreum-builder integration-tests-unsafe/upgrade integration-tests-unsafe/ibc integration-tests-unsafe/modules
	../crust/bin/crust znet remove

### Helpers go below this line

.PHONY: check-crust-builder
check-crust-builder:
	test -f ../crust/bin/crust || (echo "You need to checkout crust repository" && exit 1)

.PHONY: check-faucet-builder
check-faucet-builder:
	test -f ../faucet/bin/faucet-builder || (echo "You need to checkout faucet repository" && exit 1)
