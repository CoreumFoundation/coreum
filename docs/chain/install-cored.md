# Install binaries

The doc provides the instruction on how to install the release binaries of the `cored`.

1. [Install `cored` and `cosmovisor`](#install-cored-and-cosmovisor)
2. [Install `cored`](#install-cored)
3. [Build from sources](#build-from-sources)

## Install `cored` and `cosmovisor`

    In case you want to run any type of node, it's strongly recommended to run it with the `cosmovisor`.
    It allows automatically upgrading the `cored` on the "chain upgrade".

* Set up the CLI environment following the [doc](cli-env.md).

* Create proper folder structure for the `cosmovisor` and `cored`.

    ```bash
    mkdir -p $CORED_HOME/bin
    mkdir -p $CORED_HOME/cosmovisor/genesis/bin
    mkdir -p $CORED_HOME/cosmovisor/upgrades/bin
    mkdir -p $CORED_HOME/data
    ```

* Install the required utils: `curl` and `tar`.

* Download the binaries and put in the required folders.

    ```bash
    curl -LOf https://github.com/CoreumFoundation/coreum/releases/download/$CORED_VERSION/cored-linux-amd64
    mv cored-linux-amd64 $CORED_HOME/cosmovisor/genesis/bin/cored
    
    curl -LOf https://github.com/cosmos/cosmos-sdk/releases/download/cosmovisor%2F$CORED_COSMOVISOR_VERSION/cosmovisor-$CORED_COSMOVISOR_VERSION-linux-amd64.tar.gz
    mkdir cosmovisor-binaries
    tar -xvf "cosmovisor-$CORED_COSMOVISOR_VERSION-linux-amd64.tar.gz" -C cosmovisor-binaries
    mv "cosmovisor-binaries/cosmovisor" $CORED_HOME/bin/cosmovisor
    rm "cosmovisor-$CORED_COSMOVISOR_VERSION-linux-amd64.tar.gz"
    rm -r cosmovisor-binaries
    ```

* Set the binaries PATH and required environment variables.

    ```bash
    export PATH=$PATH:$CORED_HOME/bin
    chmod +x $CORED_HOME/cosmovisor/genesis/bin/*
    export PATH=$PATH:$CORED_HOME/cosmovisor/genesis/bin
    export DAEMON_HOME="$CORED_HOME/"
    export DAEMON_NAME="cored"
    ```
  
  **Attention!** *Set those variables globally to be automatically set after starting a new terminal session.*

* Test the binaries

    ```bash
    cored version
    ```

    ```bash
    cosmovisor version
    ```

## Install `cored`

    This option should be used in case you interact with the chain with the CLI only.

* Set up the CLI environment following the [doc](cli-env.md).

* Create a proper folder structure for the `cored`.

    ```bash
    mkdir -p $CORED_HOME/bin
    ```

* Install the required util: `curl`.

* Download the `cored` and put in the required folder.

    ```bash
    curl -LO https://github.com/CoreumFoundation/coreum/releases/download/$CORED_VERSION/cored-linux-amd64
    mv cored-linux-amd64 $CORED_HOME/bin/cored
    ```

* Add the `cored` to PATH and make it executable.

    ```bash
  
    chmod +x $CORED_HOME/bin/*
    export PATH=$PATH:$CORED_HOME/bin
    ```

  **Attention!** *Set it variable globally to be automatically set after starting a new terminal session.*

* Test the `cored`.

    ```bash
    cored version
    ```

## Build from sources

The [Build and Play](https://github.com/CoreumFoundation/coreum/blob/master/README.md#build-and-play) doc describes the
process of the `cored` binary building and installation from sources.
