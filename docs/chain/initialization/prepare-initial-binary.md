# Prepare binary

The coreum chain provides the genesis on the chain initialization with the predefined list of the initial validators.
This step describes how to prepare the initial binary.

*The steps should be repeated for each validator node.*

* Set up the CLI environment following the [doc](../cli-env.md).

* Install the required util: `curl`.

* Download the binaries and put to the required folders.

  Find the latest version on the [releases](https://github.com/CoreumFoundation/coreum/releases) page and set it to the variable.
  ```
  TEMP_CORED_VERSION={Version from release}
  ```    

  ```bash
  curl -LOf https://github.com/CoreumFoundation/coreum/releases/download/$CORED_VERSION/cored-linux-amd64
  mv cored-linux-amd64 cored
  chmod +x cored
  ```
  
  Validate `cored` binary.

  ```
  ./cored version
  ```

* Set the moniker variable to reuse it in the following instructions.

  ```bash
  export MONIKER="validator1" 
  ```

  The "validator" here is the name of the key to store in the keyring, also for simplicity, we use it as a validator name. The doc expects it to be just one word without special symbols. Also, the name should be unique per node.

* If you don't have a mnemonic create it

  ```bash
  ./cored keys add $MONIKER --keyring-backend os
  ```

You will be asked to set the keyring passphrase, set it, and remember/save it, since you will need it to access your private key.

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

* If you have the mnemonic you can import it

  ```bash
  ./cored keys add $MONIKER --keyring-backend os --recover
  ```

  You will be asked to "Enter keyring passphrase" and "Enter your bip39 mnemonic".

* Init the chain to generate the node key
  ```bash
  ./cored init $MONIKER $CORED_CHAIN_ID_ARGS
  ```

  *Keep the $CORED_HOME/config/node_key.json and $CORED_HOME/config/priv_validator_key.json in a safe place, since it can be used to recover the validator node!*

* Generate and sign create-validator tx.

  *Update the parameters in case different parameters are needed. The identity might be registered on that website https://keybase.io/*

NOTE: cored still requires connection to node even though tx generation should be fully done offline. The easiest way is to start znet locally.
  ```bash
  ./cored tx staking create-validator \
    --amount=30000000000$CORED_DENOM \
    --pubkey=$(./cored tendermint show-validator) \
    --moniker=$MONIKER \
    --website="" \
    --identity="" \
    --details="" \
    --commission-rate="0.10" \
    --commission-max-rate="0.20" \
    --commission-max-change-rate="0.01" \
    --min-self-delegation="20000000000" \
    --from=$(./cored keys show --address $MONIKER) \
    --gas=0 \
    --generate-only \
    $CORED_CHAIN_ID_ARGS > create-validator-$MONIKER-unsigned.json
  ```

  ```bash
  ./cored tx sign create-validator-$MONIKER-unsigned.json \
    --from $MONIKER \
    --output-document create-validator-$MONIKER-signed.json \
    --offline \
    --sequence=0 \
    --account-number=0 \
    --keyring-backend os \
    $CORED_CHAIN_ID_ARGS
  ```

  The "create-validator-$MONIKER-signed.json" file contains the transaction which should be put into the binary.

* Remove the temp binary
  ```bash
  rm cored
  ```

* Remove the outdated config
  ```bash
  rm $CORED_HOME/config/app.toml
  rm $CORED_HOME/config/client.toml 
  rm $CORED_HOME/config/config.toml
  rm $CORED_HOME/config/genesis.json                                          
  ```

* Repeat the same operation for each validator, collect the signed transactions, put them to the "
  pkg/config/genesis/gentx/$CORED_CHAIN_ID" folder. All transactions from that folder will be used in the genesis file.

* Update the "pkg/config/networks/network.go" file with the initial balances of the validator staker. Pay attention that the
  balance must be more than the amount in the "create-validator" transaction.
  The additional amount can be used to execute needed transactions, for example for governance.

* Run the "network_test.go" to be sure that the configuration is correct.

* Create PR and release with tag after the merge.

* Update the [cli env](../cli-env.md) doc with the new version.
