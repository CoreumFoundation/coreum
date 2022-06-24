package build

// Commands is a definition of commands available in build system
var Commands = map[string]interface{}{
	"build":              buildAll,
	"build/cored":        buildCored,
	"build/crustznet":    buildCrustZNet,
	"build/crustzstress": buildCrustZStress,
	"lint":               goLint,
	"setup":              installTools,
	"test":               goTest,
	"tidy":               goModTidy,
}
