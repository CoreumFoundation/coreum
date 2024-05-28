package main

import (
	selfBuild "github.com/CoreumFoundation/coreum/build"
	"github.com/CoreumFoundation/crust/build"
)

func main() {
	build.Main(selfBuild.Commands)
}
