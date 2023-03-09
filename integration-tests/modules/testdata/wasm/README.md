# Overview

The folder contains the source code for the wasm contracts used for the tests.
For the tests simplification we use the prebuilt wasm artifacts. 

# Build
* Make sure you have `docker` (for optimized builds)
* Rebuild the optimized contracts for tests
* Open the folder with the contract to rebuild
* Execute the command:
```
docker run --rm -v "$(pwd)"/../sdk:/sdk -v "$(pwd)":/code \
  --mount type=volume,source="$(basename "$(pwd)")_cache",target=/code/target \
  --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
  cosmwasm/rust-optimizer:0.12.6
```
