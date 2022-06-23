# crustznet
`crustznet` helps you run all the applications needed for development and testing.

## Prerequisites
To use `crustznet` you need:
- `tmux`
- `docker`

Install them manually before continuing.

## Building

Build `crustznet` using our [building system](../../../build).
To build `crustznet` use this command:

```
$ crust build/crustznet
```

`cored` binary is also required. If you hasn't built it earlier do it by running:

```
$ crust build/cored
```

You may build all the binaries at the same time by executing:

```
$ crust build
```

After doing this `crustznet` binary is available in [bin](../../../bin).

## Executing `crustznet`

`crustznet` may be executed using two methods.
First is direct where you execute command directly:

```
$ crustznet <command> [flags]
```

Second one is by entering the `environment`:

```
$ crustznet [flags]
(<environment name>) [logs] $ <command> 
```

The second method saves some typing.

## Flags

All the flags are optional. Execute

```
$ crustznet <command> --help
```

to see what the default values are.

You may enter the environment like this:

```
$ crustznet --env=crustznet --mode=dev --target=tmux
(crustznet) [logs] $
```

### --env

Defines name of the environment, it is visible in brackets on the left.
Each environment is independent, you may create many of them and work with them in parallel.

### --mode

Defines the list of applications to run. You may see their definitions in [pkg/znet/mode.go](../../pkg/znet/mode.go).

### --target

Defines where applications are deployed. Possible values:
- `tmux` - applications are started as docker containers ant their logs are presented in tmux console
- `docker` - applications are started as docker containers

## Logs

After entering environment the current directory in console is set to the one
containing logs produced by all the applications. The real path is `~/.cache/crustznet/<env-name>/logs`.

No matter what `--target` is used, logs are always dumped here, so you may analyze them using any method you like (`grep`, `cut`, `tail` etc.)

After entering and starting environment:

```
$ crustznet --env=crustznet --mode=dev --target=tmux
(crustznet) [logs] $ start
```

it is possible to use `logs` wrapper to tail logs from an application:

```
(crustznet) [logs] $ logs coredev-00
```

## Commands

In the environment some wrapper scripts for `crustznet` are generated automatically to make your life easier.
Each such `<command>` calls `crustznet <command>`.

Available commands are:
- `start` - starts applications
- `stop` - stops applications
- `remove` - stops applications and removes all the resources used by the environment
- `spec` - prints specification of the environment
- `ping-pong` - sends transactions to generate traffic on blockchain
- `stress` - tests the benchmarking logic of `crustzstress`

## Example

Basic workflow may look like this:

```
# Enter the environment:
$ crustznet --env=crustznet --mode=dev --target=tmux
(crustznet) [logs] $

# Start applications
(crustznet) [logs] $ start

# Print spec
(crustznet) [logs] $ spec

# Stop applications
(crustznet) [logs] $ stop

# Start applications again
(crustznet) [logs] $ start

# Stop everything and clean resources
(crustznet) [logs] $ remove
$
```

## Playing with the blockchain manually

For each `cored` instance started by `crustznet` wrapper script named after the name of the node is created so you may call the client manually.
There are also three standard keys: `alice`, `bob` and `charlie` added to the keystore of each instance.

If you started `crustznet` using `--mode=dev` there is one `cored` application called `cored-node`.
To use the client you may use `cored-node` wrapper:

```
(crustznet) [logs] $ coredev-00 keys list
(crustznet) [logs] $ coredev-00 query bank balances cosmos1rd8wynz2987ey6pwmkuwfg9q8hf04xdyjqy2f4
(crustznet) [logs] $ coredev-00 tx bank send bob cosmos1rd8wynz2987ey6pwmkuwfg9q8hf04xdyjqy2f4 10core
(crustznet) [logs] $ coredev-00 query bank balances cosmos1rd8wynz2987ey6pwmkuwfg9q8hf04xdyjqy2f
```

Different `cored` instances might available in another `--mode`. Run `spec` command to list them.

## Integration tests

Tests are defined in [tests/index.go](../../tests/index.go)

You may run tests directly:

```
$crustznet test
```

Tests run on top `--mode=test` and by default use `--target=tmux`

It's also possible to enter the environment first, and run tests from there:

```
$ crustznet --env=crustznet --mode=test --target=tmux
(crustznet) [logs] $ tests

# Remember to clean everything
(crustznet) [logs] $ remove
```

You may run tests using any `--target` you like so running it on top of applications deployed to `docker` is possible:

```
$ crustznet --env=crustznet --mode=test --target=docker
(crustznet) [logs] $ tests

# Remember to clean everything
(crustznet) [logs] $ remove
```

After tests complete environment is still running so if something went wrong you may inspect it manually.
Especially if you run them using `--target=tmux` it is possible to enter tmux console after tests completed:

```
$ crustznet --env=crustznet --mode=test --target=tmux
(crustznet) [logs] $ tests
(crustznet) [logs] $ start
```

and again, all the logs are available inside current directory.

## Ping-pong

There is `ping-pong` command available in `crustznet` sending transactions to generate some traffic on blockchain.
To start it runs these commands:

```
$ crustznet --target=docker
(crustznet) [logs] $ start
(crustznet) [logs] $ ping-pong
```

You will see logs reporting that tokens are constantly transferred.

## Hard reset

If you want to manually remove all the data created by `crustznet` do this:
- use `docker ps -a`, `docker stop <container-id>` and `docker rm <container-id>` to delete related running containers
- run `rm -rf ~/.cache/crustznet` to remove all the files created by `crustznet`
