# CLI environment setup

This doc describes the command to set up the cored environment depending on the type of the network you want to use.

# Network variables

\- | znet (localnet) | devnet
----|-----------------| ----
**Chain ID**   | coreum-devnet-1 | coreum-devnet-1
**Denom** | dmcore | dmcore
**Node URL**   | http://localhost:26657 | http://104.197.42.0:26657
**Faucet URL** | http://localhost:8090 | https://api.devnet-1.coreum.dev

* Set the chain env variables with the "network" corresponding values

```
export CORED_CHAIN_ID="{Chain ID}"
export CORED_DENOM="{Denom}"
export CORED_NODE="{Node URL}"
export CORED_FAUCET_URL="{Faucet URL}"

export CORED_CHAIN_ID_ARGS=(--chain-id=$CORED_CHAIN_ID)
export CORED_NODE_ARGS=(--node=$CORED_NODE $CORED_CHAIN_ID_ARGS)
```

* Check that set-up works

```
cored query bank total $CORED_NODE_ARGS
```

* Init the fund account function to use later

```bash
fund_cored_account(){ 
  echo Funding account: $1
  curl --location --request POST "$CORED_FAUCET_URL/api/faucet/v1/send-money" \
--header 'Content-Type: application/json' \
--data-raw "{ \"address\": \"$1\"}"
}
```
