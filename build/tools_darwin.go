package build

var tools = map[string]tool{
	"go": {
		Version: "1.17.9",
		URL:     "https://go.dev/dl/go1.17.9.darwin-amd64.tar.gz",
		Hash:    "sha256:af9f9deabd6e8a9d219b695b81c19276e2cd5bbc1215741e3bb1b18e88147990",
		Binaries: []string{
			"go/bin/go",
		},
	},
	"golangci": {
		Version: "1.45.2",
		URL:     "https://github.com/golangci/golangci-lint/releases/download/v1.45.2/golangci-lint-1.45.2-darwin-amd64.tar.gz",
		Hash:    "sha256:995e509e895ca6a64ffc7395ac884d5961bdec98423cb896b17f345a9b4a19cf",
		Binaries: []string{
			"golangci-lint-1.45.2-darwin-amd64/golangci-lint",
		},
	},
	"ignite": {
		Version: "v0.20.4",
		URL:     "https://github.com/ignite-hq/cli/releases/download/v0.20.4/ignite_0.20.4_darwin_amd64.tar.gz",
		Hash:    "sha256:2e9366168de8b8dbf743ec0de21c93430eca79c76d947c6de4d7c728c757f05e",
		Binaries: []string{
			"ignite",
		},
	},
}
