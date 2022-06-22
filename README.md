# Coreum Blockchain

The advent of blockchain technology has attracted huge interest in modern times. Through various protocols, the blockchain plays an important role in every business vertical. Some blockchains such as Bitcoin are a great store of value, others like Ethereum promise to verify and execute application code. While smart contract engines such as EVM are great for numerous types of applications, they remain non-deterministic and non-scalable.
In this paper, we introduce a new 3rd generation, layer 1 blockchain, Coreum.

Coreum addresses the existing limitations of the current blockchains and empowers a solid foundation for future decentralized projects.
Coreum’s unique approach is to provide built-in, on-chain solutions to process transactions in a deterministic way to ensure fast, secure, cheap and a green network for a variety of use-cases. The Coreum blockchain is distinct in many ways and incentivizes the network participants to conduct more transactions by providing bulk fee discounts.

The chain is designed to solve real-world problems at scale by providing native token management systems, Decentralized Exchange (DEX), while being fully decentralized. In addition to the built-on-chain solutions, Coreum uses WebAssembly (WASM) to process smart contracts, and utilizes the Tendermint Byzantine Fault Tolerance (BFT) consensus mechanism and Cosmos SDK’s proven Bonded Proof of Stake (BPoS).    
Coreum is built for token ecosystems such as digital assets issuing, stablecoins, traditional asset tokenizations, CBDCs, and NFTs.

## Build and Play

Coreum blockchain is under development and all the features are going to be added progressively over time.
No official devnet exists at the moment but everyone is encouraged to run a chain locally for development and testing purposes.

Entire process of running local chain is automated by our tooling. The only prerequisites are:
- `docker` and `tmux` installed from your favorite package manager
- `go 1.16` or newer installed and available in your `PATH`

### Build binaries

Steps to build required binaries:
1. Clone our [crust repository](https://github.com/CoreumFoundation/crust)
2. Not required but recommended: Add absolute path to `crust/bin` to your `PATH` environment variable
3. Run `crust/bin/core build` to compile all the required binaries

After the command completes you may find new executables in the `crust/bin` directory:
- `cored`: blockchain node and client binary
- `coreznet`: tool used to spin up local blockchain and tools
- `corezstress`: tool used to benchmark the blockchain network

### Start local chain

To start local Coreum blockchain execute:

```
$ crust/bin/coreznet
(coreznet) start
```

After a while applications will be deployed to your docker:
- `coredev-00`: single `cored` validator
- `explorer-postgres`, `explorer-hasura` and `explorer-bdjuno`: components of the block explorer (work in progress)

To stop and purge the testing environment run:

```
$ crust/bin/coreznet remove
```

To get all the details on how `coreznet` tool might be used, go to the [crust repository](https://github.com/CoreumFoundation/crust).

### Interact with the chain

After entering `coreznet` console by executing:

```
$ crust/bin/coreznet
(coreznet) $ start
```
you may use client to interact with the chain:
1. List pregenerated wallets:
```
(coreznet) $ coredev-00 keys list
```
You may use those wallets to issue transactions and queries

2. Query balances:
```
(coreznet) $ coredev-00 q bank balances core1x645ym2yz4gckqjtpwr8yddqzkkzdpkt4dfrcc
```
Remember to replace address with the one existing in your keystore.

3. Send tokens from one account to another:
```
(coreznet) $ coredev-00 tx bank send alice core1cjs7qela0trw2qyyfxw5e5e7cvwzprkj0643h8 10core
```


