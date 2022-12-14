# Run genesis validator

The doc describes the procedure of creating and running the validator node which.

* The [install binaries](../install-cored.md) doc describes the installation process.

* Set up the CLI environment following the [doc](../cli-env.md).

* Set up [node prerequisites](../node/node-prerequisites.md).

* Init the node using [instruction](run-full.md).

  *Keep the $CORED_HOME/config/node_key.json and $CORED_HOME/config/priv_validator_key.json in a safe place, since it
  can be used to recover the validator node!*

* Set the common connection config using the [doc](../node/set-connection-config.md).

* (Optional) Run sentry nodes using the [doc](../node/run-sentry.md).

* Set the moniker variable to reuse it in the following instructions.
  ```bash
  export MONIKER="validator"
  ```

* Init new validator account (if you don't have existing).

  ```
  cored keys add $MONIKER --keyring-backend os
  ```

  You will be asked to set the keyring passphrase, set it, and remember/save it, since you will need it to access your
  private key.

  The output example:

  ```bash
  - name: validator
    type: local
    address: devcore15zehkq504xqgha8cx0k6qqhems58sjysklr8p3
    pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AwzsffiidUiFtmNng5pLTH6cj1hv4Ufa+zKZpmRVGfNk"}'
    mnemonic: ""
  

  **Important** write this mnemonic phrase in a safe place.
  It is the only way to recover your account if you ever forget your password.
  
  nice equal sample cabbage demise online winner lady theory invest clarify organ divorce wheel patient gap group endless security price smoke insane link position
  ```

**Attention!** *Keep the mnemonic phrase in a safe place, since it can be used to recover the key!*

* If you have the mnemonic you can import it.

  ```bash
  cored keys add $MONIKER --keyring-backend os --recover
  ```

  You will be asked to "Enter keyring passphrase" and "Enter your bip39 mnemonic".

* Get the validator account.

  ```bash
  cored keys show $MONIKER --bech val --address --keyring-backend os
  ```

  The output example:
  ```bash
  devcore1wsc2en4yedhrqsgu7phvfgpp43jmsctwxm7r4r
  ```

* Fund the account to be able to create the validator, and check that you have enough to start.

  ```bash
  cored q bank balances  $(cored keys show $MONIKER --address --keyring-backend os) --denom $CORED_DENOM
  ``` 

* Check that node is fully synced

  ```bash
  echo "catching_up: $(echo  $(cored status) | jq -r '.SyncInfo.catching_up')"
  ``` 
  If the output `catching_up: false` the node is fully synced.

* Create validator
  ** set up validator configuration
  ```bash
   export CORED_VALIDATOR_DELEGATION_AMOUNT=20000000000 # (Required) default 20k, must be grater or equal CORED_MIN_DELEGATION_AMOUNT
   export CORED_VALIDATOR_NAME="" # update it with the name which is visible on the explorer
   export CORED_VALIDATOR_WEB_SITE="" # (Optional) update with the site
   export CORED_VALIDATOR_IDENTITY="" # (Optional) update with identity id, which can generated on the site https://keybase.io/
   export CORED_VALIDATOR_COMMISSION_RATE="0.10" # (Required) Update with commission rate
   export CORED_VALIDATOR_COMMISSION_MAX_RATE="0.20" # (Required) Update with commission max rate
   export CORED_VALIDATOR_COMMISSION_MAX_CHANGE_RATE="0.01" # (Required) Update with commission max change rate
   export CORED_MIN_DELEGATION_AMOUNT=20000000000 # (Required) default 20k, must be grater or equal min_self_delegation parameter on the current chain
  ```

  ```bash
  # create validator
  cored tx staking create-validator \
  --amount=$CORED_VALIDATOR_DELEGATION_AMOUNT$CORED_DENOM \
  --pubkey="$(cored tendermint show-validator)" \
  --moniker="$CORED_VALIDATOR_NAME" \
  --website="$CORED_VALIDATOR_WEB_SITE" \
  --identity="$CORED_VALIDATOR_IDENTITY" \
  --commission-rate="$CORED_VALIDATOR_COMMISSION_RATE" \
  --commission-max-rate="$CORED_VALIDATOR_COMMISSION_MAX_RATE" \
  --commission-max-change-rate="$CORED_VALIDATOR_COMMISSION_MAX_CHANGE_RATE" \
  --min-self-delegation=$CORED_MIN_DELEGATION_AMOUNT \
  --gas-prices 1500$CORED_DENOM \
  --gas auto \
  --gas-adjustment 1.3 \
  --chain-id="$CHAINID" \
  --from=$MONIKER \
  --gas-prices 1500$CORED_DENOM \
  --keyring-backend os -y -b block $CORED_CHAIN_ID_ARGS
  ``` 

* Check the validator status.

    ```
    cored q staking validator "$(cored keys show $MONIKER --bech val --address $CORED_CHAIN_ID_ARGS)"
    ```

  If in the output `status: BOND_STATUS_BONDED` - the validator is validating.
