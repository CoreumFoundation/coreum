package build

// Commands is a definition of commands available in build system
var Commands = map[string]interface{}{
	"build":             buildAll,
	"build/cored":       buildCored,
	"build/coreznet":    buildCoreZNet,
	"build/corezstress": buildCoreZStress,
	"lint":              goLint,
	"setup":             installTools,
	"test":              goTest,
	"tidy":              goModTidy,
}
