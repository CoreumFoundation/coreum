# Run sentry

The doc describes the procedure of creating and running the sentry node.

[Here](https://docs.tendermint.com/v0.34/tendermint-core/validators.html) you can find additional information about the node type.

* The [install binaries](../install-cored.md) doc describes the installation process.

* Set up the CLI environment following the [doc](../cli-env.md).

* Set up [node prerequisites](node-prerequisites.md)

* Set the moniker variable to reuse it in the following instructions.
  ```bash
  export MONIKER="sentry1"
  ```

* Init the node.

  ```bash
  cored init $MONIKER $CORED_CHAIN_ID_ARGS
  ```
  The command will create a default node configuration

* Install the required util: `crudini`.

* Set the common connection config using the [doc](set-connection-config.md).

* Capture the validator peer and ip to be used for connection to it.

  **Attention!** *That command must be executed on the validator node.
  The "$CORED_EXTERNAL_IP" is configured in the [doc](set-connection-config.md)",
  If it isn't set, set it for the node.*

  ```bash
  echo "CORED_VALIDATOR_PEER=$(cored tendermint show-node-id)@$CORED_EXTERNAL_IP:26656"
  echo "CORED_VALIDATOR_ID=$(cored tendermint show-node-id)"
  ```



* Set the validator peer to variable
  ```
  CORED_VALIDATOR_PEER={Validator peer from prev step}
  CORED_VALIDATOR_ID=={Validator id from prev step}
  ```

* Update the sentry node config.

  ```bash
  CORED_NODE_CONFIG=$CORED_HOME/config/config.toml
  ```

  ```bash
  crudini --set $CORED_NODE_CONFIG p2p pex true
  crudini --set $CORED_NODE_CONFIG p2p persistent_peers "\"$CORED_VALIDATOR_PEER\""
  crudini --set $CORED_NODE_CONFIG p2p private_peer_ids "\"$CORED_VALIDATOR_ID\""
  crudini --set $CORED_NODE_CONFIG p2p unconditional_peer_ids "\"$CORED_VALIDATOR_ID\""
  ```

* Capture the sentry peer to be used for connection to it.

  **Attention!** *That command must be executed on the sentry node.
  The "$CORED_EXTERNAL_IP" is configured in the [doc](set-connection-config.md)",
  If it isn't set, set it for the node.*

  ```bash
  echo "$(cored tendermint show-node-id)@$CORED_EXTERNAL_IP:26656"
  ```

* Capture the sentry ID to be used for connection to it.
  ```bash
  echo "$(cored tendermint show-node-id)"
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

* Repeat the operation for all senty nodes and capture the peers of them.

  Example of peers:
  ```
  86c5be788da1ebd1c5a7f52d5e2f159039ee218c@172.19.0.6:26656,095f7e0a462cf749027ee22913d77619fe1c2267@172.29.0.8:26656
  ```

  Example of ids:
  ```
  86c5be788da1ebd1c5a7f52d5e2f159039ee218c,095f7e0a462cf749027ee22913d77619fe1c2267
  ```

* Go to the validator node and connect it to the sentries.
  ```bash
  CORED_SENTRY_PEERS="{Sentry peers}"
  CORED_SENTRY_IDS="{Sentry ids}"
  ```

  ```bash
  crudini --set $CORED_NODE_CONFIG p2p pex false
  crudini --set $CORED_NODE_CONFIG p2p persistent_peers "\"$CORED_SENTRY_PEERS\""
  crudini --set $CORED_NODE_CONFIG p2p private_peer_ids "\"$CORED_SENTRY_IDS"\"
  crudini --set $CORED_NODE_CONFIG p2p unconditional_peer_ids "\"$CORED_SENTRY_IDS"\"
  crudini --set $CORED_NODE_CONFIG p2p addr_book_strict false
  ```
