# Prepare final binary

In the initial binary we have set the genesis configuration, but the final binary should contain the default peers configuration
to start the peering without additional data.

* Capture the seed peers from the seed.

    If you don't have them already, execute on each seed:
    ```bash
    echo "$(cored tendermint show-node-id)@$CORED_EXTERNAL_IP:26656"
    ```
  
    **Attention!** *The "$CORED_EXTERNAL_IP" is configured in the [doc](../node/set-connection-config.md),
    If it isn't set, set it for the node.*    

    Example of output:
    ```
    095f7e0a462cf749027ee22913d77619fe1c2267@172.29.0.8:26656
    ```
   
* Repeat the action for each seed to get a list of the seed peers.

* Update the "pkg/config/networks/network.go" file with coma separated seed peers. 
  Example:  
  ```go
  ...
  NodeConfig: NodeConfig{
	  SeedPeers: []string{"095f7e0a462cf749027ee22913d77619fe1c2267@172.29.0.8:26656,602df7489bd45626af5c9a4ea7f700ceb2222b19@35.223.81.227:26656"},
  },
  ...
  ```

* Run the "network_test.go" to be sure that configuration is correct.

* Create PR and release with tag after the merge.
