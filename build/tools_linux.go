package build

var tools = map[string]tool{
	"go": {
		Version: "1.17.9",
		URL:     "https://go.dev/dl/go1.17.9.linux-amd64.tar.gz",
		Hash:    "sha256:9dacf782028fdfc79120576c872dee488b81257b1c48e9032d122cfdb379cca6",
		Binaries: []string{
			"go/bin/go",
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
}
