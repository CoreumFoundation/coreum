package coreum

import "github.com/CoreumFoundation/coreum/build/tools"

// CoredUpgrades returns the mapping from upgrade name to the upgraded version.
func CoredUpgrades() map[string]string {
	return map[string]string{
		"v5": "cored",
		"v4": string(tools.CoredV401),
	}
}
