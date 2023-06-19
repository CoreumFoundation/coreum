# Overview

The folder contains the source code for the wasm contracts used for the tests.
For the tests simplification we use the prebuilt wasm artifacts.

# For a production-ready (compressed) build:

- Make sure you have `docker` (for optimized builds)
- Rebuild the optimized contracts for tests
- Open the folder with the contract to rebuild
- Execute the command:

```
docker run --rm -v "$(pwd)":/code \
  --mount type=volume,source="$(basename "$(pwd)")_cache",target=/code/target \
  --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
  cosmwasm/rust-optimizer:0.12.13
```

The optimized contracts are generated in the artifacts/ directory.
Docs can be generated using `cargo doc --no-deps`

## Contracts

| Name                             | Description                        |
| -------------------------------- | ---------------------------------- |
| [`bank-send`](./bank-send)       | Simple bank transfer               |
| [`simple-state`](./simple-state) | Simple state interaction           |
| [`ft`](./ft)                     | AssetFT custom module interaction  |
| [`nft`](./nft)                   | AssetNFT custom module interaction |

Bank-send and simple-state are here to to showcase and test the simple functionality of a CosmWasm contract. On the other side, the ft and nft contracts are useful to understand how to interact with Coreum Custom Modules. The messages and queries that we use to interact with the modules are defined in the [Coreum WASM SDK](https://github.com/CoreumFoundation/coreum-wasm-sdk) and is also available as a crate. Inside the FT and NFT contracts we detail more about these interactions.
