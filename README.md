# Coreum Blockchain

Coreum is a 3rd-generation layer-1 enterprise-grade blockchain
built to serve as a core infrastructure for decentralized applications with ISO20022 compatibility,
IBC interoperability, and novel [Smart Tokens](https://www.coreum.com/smart-tokens).

Offering 7,000 TPS, it guarantees elevated throughput, cost-effective fees, and unparalleled scalability.
WASM-based smart contracts enable diverse use cases, while the low-latency, PoS network propels rapid,
secure, and modular applications, expediting decentralized tech adoption in large-scale organizations.

<!-- markdown-link-check-disable --> 
Read more on [our website](https://www.coreum.com) and [documentation portal](https://docs.coreum.dev).
<!-- markdown-link-check-enable -->

## Build and Play

Coreum blockchain is under development and all the features are going to be added progressively over time.
Everyone is encouraged to run a chain locally for development and testing purposes.

Entire process of running local chain is automated by our tooling. The only prerequisites are:
- `docker` and `tmux` installed from your favorite package manager
- `go 1.18` or newer installed and available in your `PATH`

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
3. Compile all the required binaries and docker images:
```
$ $HOME/crust/bin/crust build images
```

After the command completes you may find executable `$HOME/crust/bin/cored`, being both blockchain node and client.

### Start local chain

To start local Coreum blockchain execute:

```
$ $HOME/crust/bin/crust znet
(znet) [znet] $ start
```

After a while applications will be deployed to your docker:
- `coredev-00`: single `cored` validator
- `explorer-postgres`, `explorer-hasura` and `explorer-bdjuno`: components of the block explorer (work in progress)

To stop and purge the testing environment run:

```
$ $HOME/crust/bin/crust znet remove
```

To get all the details on how `znet` tool might be used, go to the [crust repository](https://github.com/CoreumFoundation/crust).

### Interact with the chain

After entering `znet` console by executing:

```
$ $HOME/crust/bin/crust znet
(znet) [znet] $ start
```
you may use client to interact with the chain:
1. List pregenerated wallets:
```
(znet) [znet] $ coredev-00 keys list
```
You may use those wallets to issue transactions and queries

2. Query balances:
```
(znet) [znet] $ coredev-00 q bank balances devcore1x645ym2yz4gckqjtpwr8yddqzkkzdpkt8nypky
```
Remember to replace address with the one existing in your keystore.

3. Send tokens from one account to another:
```
(znet) [znet] $ coredev-00 tx bank send alice devcore1cjs7qela0trw2qyyfxw5e5e7cvwzprkjaycnem 10core
```

## Devnet

You may connect to Coreum devnet network by connecting to host `s-0.devnet-1.coreum.dev` on port `443`:

```
$ cored status --chain-id=coreum-devnet-1 --node=https://s-0.devnet-1.coreum.dev:443
```

Block explorer for devnet is available at [https://explorer.coreum.com](https://explorer.coreum.com/coreum)
