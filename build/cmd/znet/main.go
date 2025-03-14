package main

import (
	coreumbuild "github.com/CoreumFoundation/coreum/build/coreum"
	"github.com/CoreumFoundation/crust/znet/infra"
	"github.com/CoreumFoundation/crust/znet/pkg/znet"
)

func main() {
	znet.Main(infra.ConfigFactoryWithCoredUpgrades(coreumbuild.CoredUpgrades()))
}
