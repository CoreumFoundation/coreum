package app

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/types/module"
)

//const appName = "CoreumDexApp"

var (
	ModuleBasics = module.NewBasicManager()
)

type CoreumApp struct {
	*baseapp.BaseApp
}
