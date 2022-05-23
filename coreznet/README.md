# coreznet
`coreznet` helps you run all the applications needed for development and testing.

## Prerequisites
To use `coreznet` you need:
- `tmux`
- `docker`

Install them manually before continuing.

## Building

Build `coreznet` using our [building system](../build).
To build `coreznet` use this command:

```
$ core build/coreznet
```

`cored` binary is also required. If you hasn't built it earlier do it by running:

```
$ core build/cored
```

You may build all the binaries at the same time by executing:

```
$ core build
```

After doing this `coreznet` binary is available in [bin](../bin).

## Executing `coreznet`

`coreznet` may be executed using two methods.
First is direct where you execute command directly:

```
$ coreznet <command> [flags]
```

Second one is by entering the `environment`:

```
$ coreznet [flags]
(<environment name>) [logs] $ <command> 
```

The second method saves some typing.

## Flags

All the flags are optional. Execute

```
$ coreznet <command> --help
```

to see what the default values are.

You may enter the environment like this:

```
$ coreznet --env=coreznet --mode=dev --target=tmux
(coreznet) [logs] $
```

### --env

Defines name of the environment, it is visible in brackets on the left.
Each environment is independent, you may create many of them and work with them in parallel.

### --mode

Defines the list of applications to run. You may see their definitions in [mode.go](mode.go).

### --target

Defines where applications are deployed. Possible values:
- `tmux` - applications are started as OS processes and their logs are presented in tmux console
- `docker` - applications are started as docker containers
- `direct` - applications are started as OS processes

## Logs

After entering environment the current directory in console is set to the one
containing logs produced by all the applications. The real path is `~/.cache/coreznet/<env-name>/logs`.

No matter what `--target` is used, logs are always dumped here, so you may analyze them using any method you like (`grep`, `cut`, `tail` etc.)

After entering and starting environment:

```
$ coreznet --env=coreznet --mode=dev --target=tmux
(coreznet) [logs] $ start
```

it is possible to use `logs` wrapper tot ail logs from an application:

```
(coreznet) [logs] $ logs cored-node
```

## Commands

In the environment some wrapper scripts for `coreznet` are generated automatically to make your life easier.
Each such `<command>` calls `coreznet <command>`.

Available commands are:
- `start` - starts applications
- `stop` - stops applications
- `remove` - stops applications and removes all the resources used by the environment
- `spec` - prints specification of the environment

## Example

Basic workflow may look like this:

```
# Enter the environment:
$ coreznet --env=coreznet --mode=dev --target=tmux
(coreznet) [logs] $

# Start applications
(coreznet) [logs] $ start

# Print spec
(coreznet) [logs] $ spec

# Stop applications
(coreznet) [logs] $ stop

# Start applications again
(coreznet) [logs] $ start

# Stop everything and clean resources
(coreznet) [logs] $ remove
(coreznet) [logs] $ exit
$
```

## Integration tests

Tests are defined in [tests/index.go](tests/index.go)

You may run tests directly:

```
$coreznet test
```

Tests run on top `--mode=test` and by default use `--target=tmux`

It's also possible to enter the environment first, and run tests from there:

```
$ coreznet --env=coreznet --mode=test --target=tmux
(coreznet) [logs] $ test

# Remember to clean everything
(coreznet) [logs] $ remove
```

You may run tests using any `--target` you like so running it on top of applications deployed to `docker` is possible:

```
$ coreznet --env=coreznet --mode=test --target=docker
(coreznet) [logs] $ test

# Remember to clean everything
(coreznet) [logs] $ remove
```

After tests complete environment is still running so if something went wrong you may inspect it manually.
Especially if you run them using `--target=tmux` it is possible to enter tmux console after tests completed:

```
$ coreznet --env=coreznet --mode=test --target=tmux
(coreznet) [logs] $ test
(coreznet) [logs] $ start
```

and again, all the logs are available inside current directory.

## Hard reset

If you want to manually remove all the data created by `coreznet` do this:
- use `ps aux` to find all the related running processes and kill them using `kill -9 <pid>`
- use `docker ps -a`, `docker stop <container-id>` and `docker rm <container-id>` to delete related running containers
- use `docker images` and `docker rmi <image-id>` to remove related docker images 
- run `rm -rf ~/.cache/coreznet` to remove all the files created by `coreznet`