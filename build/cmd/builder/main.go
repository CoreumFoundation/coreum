package main

import (
	selfBuild "github.com/CoreumFoundation/coreum/build"
	selfTools "github.com/CoreumFoundation/coreum/build/tools"
	"github.com/CoreumFoundation/crust/build"
	"github.com/CoreumFoundation/crust/build/tools"
)

func init() {
	tools.AddTools(selfTools.Tools...)
}

func main() {
	build.Main(selfBuild.Commands)
}
