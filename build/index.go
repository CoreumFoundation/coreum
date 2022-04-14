package build

// Commands is a definition of commands available in build system
var Commands = map[string]interface{}{
	"setup": installTools,
	"lint":  goLint,
	"test":  goTest,
	"build": buildCored,
}
