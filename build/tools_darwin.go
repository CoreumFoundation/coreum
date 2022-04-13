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
}
