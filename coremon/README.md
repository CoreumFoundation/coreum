## CoreMon

CoreMon is a block processing service currently set up to monitor various chain metrics with enhanced accuracy (per-block measurements, etc). This provides an alternative pipeline to Prometheus and used to gather more info on benchmark result, also use for chain debugging purposes.


## Usage

Options may be set as flags, env vars, also via `.env` file (dotenv format).

```
> coremon -h

Usage: coremon [OPTIONS] COMMAND [arg...]

Daemon for cosmos chain monitoring and accurate stats exporting.

Options:
  -e, --env          The environment name this app runs in. Used for metrics and error reporting. (env $COREMON_ENV) (default "local")
      --log-format   Format of log output: console | json (env $COREMON_LOG_FORMAT) (default "json")
  -v, --verbose      Turns on verbose logging. (env $COREMON_LOG_VERBOSE) (default "false")

Commands:
  process            Start chain blocks processing

Run 'coremon COMMAND --help' for more information on a command.
```

Processing subcommand:
```
> coremon process -h

Usage: coremon process [OPTIONS]

Start chain blocks processing

Options:
      --chain-id              Specify Chain ID of the network. (env $COREMON_CHAIN_ID) (default "coredev")
      --tendermint-rpc        Tendermint RPC endpoint (env $COREMON_TENDERMINT_RPC) (default "http://localhost:26657")
      --parallel-fetch-jobs   Number of sumultaneous jobs fetching the blocks from the RPC. Change only if need to access historical. (env $COREMON_BLOCK_FETCH_JOBS) (default 1)
```

## Building

```
> go install ./cmd/coremon
```

Later option to build using `core build` will be added.

## Docker

```
> docker build -t coremon .

> docker run -it --rm coremon process -h

Usage: coremon process [OPTIONS]

...
```

## Example with local node

In the background, a `corezstress` instance is running (started on the second minute).

```
 > coremon process

2022-06-06 13:00:41.055 INFO CoreMon Block Watching routine starts
2022-06-06 13:00:41.063 INFO Block Sync: At block height 0 while chain is at 12012 (2022-06-06T13:00:39Z)
2022-06-06 13:00:41.063 INFO Block Sync: Initial sync done. Continuing to poll TmRPC for the new blocks.
2022-06-06 13:01:41.055 INFO blocks synced: 53/m in 1m0s
2022-06-06 13:02:41.056 INFO blocks synced: 53/m in 1m0s
2022-06-06 13:02:41.056 INFO tx seen: 10/m in 1m0s
2022-06-06 13:03:41.057 INFO blocks synced: 50/m in 1m0s
2022-06-06 13:03:41.057 INFO tx seen: 13413/m in 1m0s
2022-06-06 13:04:41.058 INFO blocks synced: 54/m in 1m0s
2022-06-06 13:04:41.058 INFO tx seen: 16205/m in 1m0s
2022-06-06 13:05:41.058 INFO blocks synced: 51/m in 1m0s
2022-06-06 13:05:41.058 INFO tx seen: 13909/m in 1m0s
2022-06-06 13:06:41.059 INFO blocks synced: 53/m in 1m0s
2022-06-06 13:06:41.060 INFO tx seen: 14516/m in 1m0s
2022-06-06 13:07:41.063 INFO tx seen: 14498/m in 1m0s
2022-06-06 13:07:41.063 INFO blocks synced: 50/m in 1m0s
2022-06-06 13:08:41.064 INFO blocks synced: 53/m in 1m0s
2022-06-06 13:08:41.064 INFO tx seen: 15497/m in 1m0s
2022-06-06 13:09:41.065 INFO tx seen: 14843/m in 1m0s
2022-06-06 13:09:41.065 INFO blocks synced: 52/m in 1m0s
2022-06-06 13:10:41.066 INFO blocks synced: 51/m in 1m0s
2022-06-06 13:10:41.066 INFO tx seen: 13507/m in 1m0s
2022-06-06 13:11:41.067 INFO tx seen: 15066/m in 1m0s
2022-06-06 13:11:41.067 INFO blocks synced: 51/m in 1m0s
```