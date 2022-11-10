# Run genesis validator

The doc describes the procedure of creating and running the validator node which was set in the genesis file.

* The [install binaries](../install-cored.md) doc describes the installation process.

* Set up the CLI environment following the [doc](../cli-env.md).

* Set up [node prerequisites](../node/node-prerequisites.md).

* Set the moniker variable to reuse it in the following instructions.
  ```bash
  export MONIKER="validator1"
  ```

  **Attention!** *If you have already installed the validator keys on the node the name should be the same as before.*

* Init the node

  ```bash
  cored init $MONIKER $CORED_CHAIN_ID_ARGS
  ```
  The command will create a default node configuration, keeping the previously generated node keys.

* Set the common connection config using the [doc](../node/set-connection-config.md).

* Run sentry nodes using the [doc](../node/run-sentry.md).

* Start the node.

  * Start with `cosmovisor` (recommended)
  ```bash
  cosmovisor run start $CORED_CHAIN_ID_ARGS
  ```

  * Start with `cored`
   ```bash
  cored start $CORED_CHAIN_ID_ARGS
  ```

  **Attention!** *Be sure that node will be automatically started after starting a new terminal session. Add it as an OS "service" or schedule to start using the tools you prefer.*
