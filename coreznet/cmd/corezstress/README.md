# corezstress
`corezstress` is used to generate infrastructure for and run benchmarks to test the performance of the network of validators and sentry nodes.

## Building

Build `corezstress` using our [building system](../../../build).
To build `corezstress` use this command:

```
$ core build/corezstress
```

Also `cored` is required. Build it using this command:

```
$ core build/cored
```

You may also build everything at once:

```
$ core build
```

After doing this binaries are available in [bin](../../../bin).
If you work on macOS, linux binaries are also created in [bin/linux](../../../bin/linux) which might be used to build
docker images.

## Infrastructure for benchmarking

This is how the infrastructure for testing is going to look like:
- number of validators in different data centers, running `cored` docker containers
- number of sentry nodes in different data centers, running `cored` docker containers
- number of instances used to broadcast tons of transactions to sentry nodes in parallel, running `corezstress` docker containers

## Generation

`corezstress generate` command prepares all the files required by our DevOps department to create the infrastructure
described above, including:
- genesis configuration for validators and sentry nodes
- private keys and configuration for validators and sentry nodes
- files containing node IDs of validators and sentry nodes
- private keys of wallets used to generate transactions on each instance
- files to generate docker image of `cored`
- files to generate docker image of `corezstress`

These are the CLI flags accepted by `corezstress generate` command:

- `--out` - path to the directory where generated files are stored
- `--chain-id` - ID of the chain to generate
- `--accounts` - maximum number of funded accounts per each instance used in the future during benchmarking
- `--validators` - number of validators present on the blockchain
- `--sentry-nodes` - number of sentry nodes to generate config for
- `--instances` - maximum number of application instances used in the future during benchmarking

After running the command directory `corezstress-deployment` is created under the path defined by `--out` option.

This is how its content looks like:
- `validators` - contains configuration for each validator
- `validators/x/config` - should be mounted to `/config` inside `cored` docker container
- `validators/x/data` - should be mounted to `/data` inside `cored` docker container
- `validators/ids.json` - contains node IDs of validators
- `sentry-nodes` - contains configuration for each sentry node
- `sentry-nodes/x/config` - should be mounted to `/config` inside `cored` docker container
- `sentry-nodes/ids.json` - contains node IDs of sentry-nodes
- `instances/x/accounts.json` - contains private keys of wallets funded in genesis block to be used by the instance broadcasting transactions, two instances must not (!!!) use the same file, file must be mounted inside `corezstress` docker container
- `docker-cored` - files used to build `cored` docker container
- `docker-corezstress` - files used to build `corezstress` docker container

## Benchmarking

During benchmarks `corezstress` command deployed to instances is used to broadcast transactions.

Accepted CLI config options are:
- `--chain-id` - ID of the chain to connect to, this must match the chain ID used during generation
- `--node-addr` - address of a `cored` node RPC endpoint, in the form of host:port, to connect to
- `--account-file` - path to a JSON file containing private keys of accounts funded on blockchain, this is one of `accounts.json` files under `instances` directory generated in previous step
- `--accounts` - number of accounts used to benchmark the node in parallel, must not be greater than the number of keys available in account file
- `--transactions` - number of transactions to send from each account

Keep in mind that number of transactions specified by `--transactions` is executed concurrently by each account, so if you
execute `corezstress --accounts=1000 --transactions=1000` then the total number of 1 million transactions is executed.
If you use 16 instances to run the benchmark then 16 millions of transactions is broadcasted to the blockchain all together.

Accounts send transactions in parallel, so if you run an instance with 16 accounts then there are 16 goroutines,
each broadcasting one transaction at a time. The same account broadcasts the next transaction only after the previous one is included in a block.
This means that each goroutine spends most of the time on waiting for transaction to be included in a block.
That's why it safe to run many times more accounts on each instance than the number of CPU cores available on the server.
The exact number must be determined by practice.

The process of signing the transaction is the most time-consuming element of broadcasting. To generate maximum throughput
during the test, all the transactions are generated and presigned first and only then all the concurrent accounts start
broadcasting them.

## Rising up the limits

If you're on Linux, you should have the ability to rise the system-wide limits for open files and whole TCP stack. This is important as the stresser opens a lot of HTTP connections, and your cored node accepts a lot of HTTP connections, and if limits are not properly set for both sides, the node may crash, and stresser client will degrade in performance.

We suggest that you set these limits (inspired by [Sample config for 2 million web socket connection](https://gist.github.com/joennlae/7c822f641d78117eedcae6a68c2c3964)): 

```bash
#!/bin/sh

sysctl -w fs.file-max=12000500
sysctl -w fs.nr_open=20000500

# Set the maximum number of open file descriptors
ulimit -n 20000000

# Set the memory size for TCP with minimum, default and maximum thresholds
sysctl -w net.ipv4.tcp_mem='10000000 10000000 10000000'

# Set the receive buffer for each TCP connection with minumum, default and maximum thresholds
sysctl -w net.ipv4.tcp_rmem='1024 4096 16384'

# Set the TCP send buffer space with minumum, default and maximum thresholds
sysctl -w net.ipv4.tcp_wmem='1024 4096 16384'

# The maximum socket receive buffer sizemem_max=16384
sysctl -w net.core.rmem_max=16384

# The maximum socket send buffer size
sysctl -w net.core.wmem_max=16384
```

## Gathering results

Each validator and sentry node exposes prometheus endpoint on port `26660` so it's possible to collect them and inspect easily
in any tool like Grafana.