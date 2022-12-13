# Run full

The doc describes the procedure of creating and running the full node.

* The [install binaries](../install-cored.md) doc describes the installation process.

* Set up the CLI environment following the [doc](../cli-env.md).

* Set up [node prerequisites](node-prerequisites.md)

* Set the moniker variable to reuse it in the following instructions.
  ```bash
  export MONIKER="full"
  ```

* Init the node.

  ```bash
  cored init $MONIKER $CORED_CHAIN_ID_ARGS
  ```
  The command will create a default node configuration

* Install the required utils: `crudini, curl, jq`.

* Set the common connection config using the [doc](set-connection-config.md).

* Set the config path variables.

  ```bash
  CORED_NODE_CONFIG=$CORED_HOME/config/config.toml
  CORED_APP_CONFIG=$CORED_HOME/config/app.toml
  ```

* (Optional) Enable REST APIs disabled by default.
  ```bash
  crudini --set $CORED_APP_CONFIG api enable true # enable API
  crudini --set $CORED_APP_CONFIG api swagger true # enable swagger UI for the API
  ```

* (Optional) Enable prometheus monitoring.
  ```bash
  crudini --set $CORED_NODE_CONFIG instrumentation prometheus true
  ```

* (Optional) Enable state-sync snapshotting (for state-sync servers).
  You can read [Using State Sync](https://docs.tendermint.com/v0.34/tendermint-core/state-sync.html) document to get
  more details.
  ```bash
  crudini --set $CORED_APP_CONFIG state-sync snapshot-interval 500
  crudini --set $CORED_APP_CONFIG state-sync snapshot-keep-recent 3
  ```
  That configuration is required for the state state-sync servers, used as a snapshot provided for the nodes.


* (Optional) Enable state-sync.
  You can read [Using State Sync](https://docs.tendermint.com/v0.34/tendermint-core/state-sync.html) document to get
  more details.

  ** Open the [doc](../cli-env.md) and set the `State sync servers` variable.
  ```bash
  export COREUM_STATE_SYNC_SERVERS="{State sync servers}" # example "foo.net:26657,bar.com:26657"
  ```

  ** Get the trusted block hash and height
  ```bash
  # Get block details from one of the state sync servers
  TRUSTED_BLOCK_DETAILS=$(curl http://${COREUM_STATE_SYNC_SERVERS#*,}/block | jq -r '.result.block.header.height + "\n" + .result.block_id.hash')
  TRUSTED_BLOCK_HEIGHT=$(echo $TRUSTED_BLOCK_DETAILS | cut -d$' ' -f1)
  TRUSTED_BLOCK_HASH=$(echo $TRUSTED_BLOCK_DETAILS | cut -d$' ' -f2)
  echo "height:$TRUSTED_BLOCK_HEIGHT, hash:$TRUSTED_BLOCK_HASH"
  # Enable state sync
      # Change statesync settings
  crudini --set $CORED_NODE_CONFIG statesync enable true
  crudini --set $CORED_NODE_CONFIG statesync rpc_servers "\"$COREUM_STATE_SYNC_SERVERS\""
  crudini --set $CORED_NODE_CONFIG statesync trust_height $TRUSTED_BLOCK_HEIGHT
  crudini --set $CORED_NODE_CONFIG statesync trust_hash "\"$TRUSTED_BLOCK_HASH\""
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

  **Attention!** *Be sure that the node will be automatically started after starting a new terminal session. Add it as
  an OS "service",
  or schedule the start using the tools you prefer.*
