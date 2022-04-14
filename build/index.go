package build

// Commands is a definition of commands available in build system
var Commands = map[string]interface{}{
	"lint":  goLint,
	"test":  goTest,
	"build": buildCored,
}
