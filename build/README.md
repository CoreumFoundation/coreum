# Build system

## Motivation

We need a tool to automate common tasks:
- building binaries
- linting
- testing
- releasing
- installing development tools

Most common approach is to use `Makefile`. This approach would introduce a lot of bash code to our project.
The other option is to write everything in go to keep our code as consistent as possible.
Moreover, we are much better go developers than bash ones.

So here is the simple tool written in go which helps us in our daily work.

## Configuration

Assuming you cloned `coreum` repository to `~/coreum` this is the configuration to put in
your `~/.bashrc`:

```
PATH="$HOME/coreum/bin:$PATH"
complete -o nospace -C core core
```

then run:

```
$ core setup
```

to install all the tools we use.

Whenever tool downloads or builds binaries it puts them inside [bin](../bin) directory so they are
easily callable from console.

After doing this and restarting `bash` session you may call `core` command.

## `core` command

`core` command is used to execute operations. you may pass one or more operations to it:

`core <op-1> <op-2> ... <op-n>`

Here is the list of operations supported at the moment:

- `setup` - install all the tools required to develop our software
- `lint` - runs code linter
- `test` - runs unit tests
- `build` - builds `cored`

If you want to inspect source code of operations, go to [build/index.go](index.go). 

You may run operations one by one:

```
$ core lint
$ core test
```

or together:

```
$ core lint test
```

Running operations together is better because if they internally have common dependencies, each of them will
be executed once. Moreover, each execution of `core` may compile code. By running more operations at once
you just save your time. In all the cases operations are executed sequentially.

## Common environment

The build tool is also responsible for installing external binaries required by our environment.
The goal is to keep our environment consistent across all the computers used by our team members.

So whenever `go` binary or anything else is required to complete the operation, the build tool ensures
that correct version is used. If the version hasn't been installed yet, it is downloaded automatically for you.