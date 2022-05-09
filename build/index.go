package build

// Commands is a definition of commands available in build system
var Commands = map[string]interface{}{
	"setup":          installTools,
	"lint":           goLint,
	"tidy":           goModTidy,
	"test":           goTest,
	"build":          buildAll,
	"build/coreznet": buildCoreZNet,
	"build/cored":    buildCored,
}
