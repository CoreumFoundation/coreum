package build

var tools = map[string]tool{
	"go": {
		Version: "1.18.1",
		URL:     "https://go.dev/dl/go1.18.1.linux-amd64.tar.gz",
		Hash:    "sha256:b3b815f47ababac13810fc6021eb73d65478e0b2db4b09d348eefad9581a2334",
		Binaries: []string{
			"go/bin/go",
			"go/bin/gofmt",
		},
	},
	"golangci": {
		Version: "1.45.2",
		URL:     "https://github.com/golangci/golangci-lint/releases/download/v1.45.2/golangci-lint-1.45.2-linux-amd64.tar.gz",
		Hash:    "sha256:595ad6c6dade4c064351bc309f411703e457f8ffbb7a1806b3d8ee713333427f",
		Binaries: []string{
			"golangci-lint-1.45.2-linux-amd64/golangci-lint",
		},
	},
	"ignite": {
		Version: "v0.20.4",
		URL:     "https://github.com/ignite-hq/cli/releases/download/v0.20.4/ignite_0.20.4_linux_amd64.tar.gz",
		Hash:    "sha256:6291e0e3571cfc81caa691932024519cabade44c061d4214f5f4090badb06ab2",
		Binaries: []string{
			"ignite",
		},
	},
}
