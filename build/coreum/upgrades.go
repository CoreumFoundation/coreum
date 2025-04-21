package coreum

import "github.com/CoreumFoundation/coreum/build/tools"

// CoredUpgrades returns the mapping from upgrade name to the upgraded version.
func CoredUpgrades() map[string]string {
	return map[string]string{
		"v6": "cored",
		"v5": string(tools.CoredV500),
	}
}
