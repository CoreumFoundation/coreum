# Set connection config

The doc describes the common connection configuration for any type of node.

 ```bash
  CORED_NODE_CONFIG=$CORED_HOME/config/config.toml
  ```

  ```bash
  CORED_EXTERNAL_IP=$(hostname -i)
  echo "External IP: $CORED_EXTERNAL_IP"
  ```

**Attention!** *The "CORED_EXTERNAL_IP" is the address that should be accessible for other nodes to be connected to it.
It depends on the network configuration, in that example, we use the simplest option where the "hostname" is the address.*

  ```bash
  crudini --set $CORED_NODE_CONFIG p2p addr_book_strict false
  crudini --set $CORED_NODE_CONFIG p2p external_address "\"tcp://$CORED_EXTERNAL_IP:26656\""
  crudini --set $CORED_NODE_CONFIG rpc laddr "\"tcp://0.0.0.0:26657\""
  ```
