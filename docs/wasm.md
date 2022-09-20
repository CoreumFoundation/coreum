# WASM

The doc provides the brief overview of the integrated WASM module with the tutorial how to build and deploy simple
WASM contract.

# Overview

WASM module is a Cosmos SDK based module which is plugged into the coreum chain.
This module allows to build, deploy and use the WASM contracts on the coreum chain.
Currently, it supports the Rust language as a language of the contracts.

The detailed module architecture overview can be
found [here](https://docs.cosmwasm.com/docs/1.0/architecture/multichain).

The contract semantics is [here](https://docs.cosmwasm.com/docs/1.0/smart-contracts/contract-semantics).

# Write and deploy first contract

The section provides the tutorial how to build and deploy the WASM contract to the coreum chain.

## Set up environment

### Build the cored

The [Build and Play](https://github.com/CoreumFoundation/coreum/blob/master/README.md#build-and-play) doc describes the
process of the cored binary building and installation.

### Set up the CLI environment

Set up the CLI environment following the [doc](cli-env.md).

### Install the CLI utils

* Mac OS.

```bash
brew install jq curl
```

* Linux (Ubuntu and Debian).

```bash
apt install jq curl
```

## Write, build and deploy WASM contract

* Clone the smart contract template.

```bash
git clone https://github.com/CoreumFoundation/cw-contracts.git
```

* Open the template folder.

```bash
cd cw-contracts
git checkout main
cd contracts/nameservice
```

* Generate a new wallet for testing.

```bash
cored keys add wallet $CORED_CHAIN_ID_ARGS
```

* Fund the wallet from the faucet and check the balance.

Using the CLI:

```bash
fund_cored_account $(cored keys show --address wallet $CORED_CHAIN_ID_ARGS)
cored q bank balances $(cored keys show --address wallet $CORED_CHAIN_ID_ARGS) $CORED_NODE_ARGS
```

Or use the [faucet doc](https://docs.coreum.dev/faucet/).

* Build optimized WASM smart contract.

```bash
docker run --rm -v "$(pwd)":/code \
  --mount type=volume,source="$(basename "$(pwd)")_cache",target=/code/target \
  --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
  cosmwasm/rust-optimizer:0.12.6
```

This operation might take the significant amount of time.

* List the already deployed contract codes.

```bash
cored q wasm list-code $CORED_NODE_ARGS
```

* Deploy the built artifact.

```bash
RES=$(cored tx wasm store artifacts/cw_nameservice.wasm \
    --from wallet --gas-prices 1500$CORED_DENOM --gas auto --gas-adjustment 1.3 -y -b block --output json $CORED_NODE_ARGS)
echo $RES    
CODE_ID=$(echo $RES | jq -r '.logs[0].events[-1].attributes[0].value')
echo $CODE_ID
```

* Check the deployed code.

```bash
cored q wasm code-info $CODE_ID $CORED_NODE_ARGS
```

* Instantiating the contract.

```bash
INIT="{\"purchase_price\":{\"amount\":\"100\",\"denom\":\"$CORED_DENOM\"},\"transfer_price\":{\"amount\":\"999\",\"denom\":\"$CORED_DENOM\"}}"
cored tx wasm instantiate $CODE_ID "$INIT" --from wallet --gas-prices 1500$CORED_DENOM --label "name service" -b block -y --no-admin $CORED_NODE_ARGS
```

* Check the contract details and account balance.

```bash
cored query wasm list-contract-by-code $CODE_ID --output json $CORED_NODE_ARGS
CONTRACT=$(cored query wasm list-contract-by-code $CODE_ID --output json $CORED_NODE_ARGS | jq -r '.contracts[-1]')
echo $CONTRACT
```

* Register a name for the wallet address on the contract.

```bash
REGISTER='{"register":{"name":"fred"}}'
cored tx wasm execute $CONTRACT "$REGISTER" --amount 100$CORED_DENOM --from wallet --gas-prices 1500$CORED_DENOM -b block -y $CORED_NODE_ARGS
```

* Query the owner of the name record.

```bash
NAME_QUERY='{"resolve_record": {"name": "fred"}}'
cored query wasm contract-state smart $CONTRACT "$NAME_QUERY" --output json $CORED_NODE_ARGS
```

The owner is the "wallet" now.

* Transfer the ownership of the name record to "new-owner" wallet.

```bash
cored keys add new-owner $CORED_CHAIN_ID_ARGS
RECIPIENT_ADDRESS=$(cored keys show --address new-owner $CORED_CHAIN_ID_ARGS)
TRANSFER="{\"transfer\":{\"name\":\"fred\",\"to\":\"$RECIPIENT_ADDRESS\"}}"
cored tx wasm execute $CONTRACT "$TRANSFER" --amount 999$CORED_DENOM --from wallet --gas-prices 1500$CORED_DENOM -b block -y $CORED_NODE_ARGS
``` 

* Query the record owner again to see the new owner address.

```bash
echo "Recipient address: $RECIPIENT_ADDRESS"
NAME_QUERY='{"resolve_record": {"name": "fred"}}'
cored query wasm contract-state smart $CONTRACT "$NAME_QUERY" --output json $CORED_NODE_ARGS
```
