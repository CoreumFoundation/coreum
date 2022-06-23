# Coreum Blockchain

Coreum addresses the existing limitations of the current blockchains and empowers a solid foundation for future decentralized projects.
Coreum’s unique approach is to provide built-in, on-chain solutions to process transactions in a deterministic way to ensure fast, secure, cheap and a green network for a variety of use-cases.

The chain is designed to solve real-world problems at scale by providing native token management systems, Decentralized Exchange (DEX), while being fully decentralized. In addition to the built-on-chain solutions, Coreum uses WebAssembly (WASM) to process smart contracts, and utilizes the Tendermint Byzantine Fault Tolerance (BFT) consensus mechanism and Cosmos SDK’s proven Bonded Proof of Stake (BPoS).    

Read more on [our website](https://www.coreum.com).

## Build and Play

Coreum blockchain is under development and all the features are going to be added progressively over time.
No official devnet exists at the moment but everyone is encouraged to run a chain locally for development and testing purposes.

Entire process of running local chain is automated by our tooling. The only prerequisites are:
- `docker` and `tmux` installed from your favorite package manager
- `go 1.16` or newer installed and available in your `PATH`

### Build binaries

Steps to build required binaries:
1. Clone our [crust repository](https://github.com/CoreumFoundation/crust) to your `$HOME` directory:
```
$ cd $HOME
$ git clone https://github.com/CoreumFoundation/crust
```
2. Not required but recommended: Add `$HOME/crust/bin` to your `PATH` environment variable:
```
$ export PATH="$HOME/crust/bin:$PATH"
```
3. Compile all the required binaries:
```
$ $HOME/crust/bin/crust build
```

After the command completes you may find new executables in the `$HOME/crust/bin` directory:
- `cored`: blockchain node and client binary
- `crustznet`: tool used to spin up local blockchain and tools
- `crustzstress`: tool used to benchmark the blockchain network

### Start local chain

To start local Coreum blockchain execute:

```
$ $HOME/crust/bin/crustznet
(crustznet) start
```

After a while applications will be deployed to your docker:
- `coredev-00`: single `cored` validator
- `explorer-postgres`, `explorer-hasura` and `explorer-bdjuno`: components of the block explorer (work in progress)

To stop and purge the testing environment run:

```
$ $HOME/crust/bin/crustznet remove
```

To get all the details on how `crustznet` tool might be used, go to the [crust repository](https://github.com/CoreumFoundation/crust).

### Interact with the chain

After entering `crustznet` console by executing:

```
$ $HOME/crust/bin/crustznet
(crustznet) $ start
```
you may use client to interact with the chain:
1. List pregenerated wallets:
```
(crustznet) $ coredev-00 keys list
```
You may use those wallets to issue transactions and queries

2. Query balances:
```
(crustznet) $ coredev-00 q bank balances core1x645ym2yz4gckqjtpwr8yddqzkkzdpkt4dfrcc
```
Remember to replace address with the one existing in your keystore.

3. Send tokens from one account to another:
```
(crustznet) $ coredev-00 tx bank send alice core1cjs7qela0trw2qyyfxw5e5e7cvwzprkj0643h8 10core
```
