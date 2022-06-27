package build

// Commands is a definition of commands available in build system
var Commands = map[string]interface{}{
	"build":         buildAll,
	"build/crust":   buildCrust,
	"build/cored":   buildCored,
	"build/znet":    buildZNet,
	"build/zstress": buildZStress,
	"lint":          goLint,
	"setup":         installTools,
	"test":          goTest,
	"tidy":          goModTidy,
}
