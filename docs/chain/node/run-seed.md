# Run seed

The doc describes the procedure of creating and running the seed node.

[Here](https://docs.tendermint.com/v0.34/tendermint-core/using-tendermint.html#seed) you can find additional information about the node type.

* The [install binaries](../install-cored.md) doc describes the installation process.

* Set up the CLI environment following the [doc](../cli-env.md).

* Set up [node prerequisites](node-prerequisites.md)

* Set the moniker variable to reuse it in the following instructions.
  ```bash
  export MONIKER="seed1"
  ```

* Init the node.

  ```bash
  cored init $MONIKER $CORED_CHAIN_ID_ARGS
  ```
  The command will create a default node configuration

* Install the required util: `crudini`.

* Set the common connection config using the [doc](set-connection-config.md).

* Update the seed node config.

  ```bash
  CORED_NODE_CONFIG=$CORED_HOME/config/config.toml
  ```

  ```bash
  crudini --set $CORED_NODE_CONFIG p2p seed_mode true
  ```

  In case you don't want to connect your seed to default seeds or in case it's the first launch - reset the seeds:
  ```bash
  crudini --set $CORED_NODE_CONFIG p2p seeds "\"\"" 
  ```

* Capture the seed peer to be used for connection to it.
  ```bash
  echo "$(cored tendermint show-node-id)@$CORED_EXTERNAL_IP:26656"
  ```

* Start the node.

  * Start with `cosmovisor` (recommended)
  ```bash
  cosmovisor run start $CORED_CHAIN_ID_ARGS
  ```

  * Start with `cored`
   ```bash
  cored start $CORED_CHAIN_ID_ARGS
  ```

  **Attention!** *Be sure that the node will be automatically started after starting a new terminal session. Add it as an OS "service",
  or schedule the start using the tools you prefer.*
