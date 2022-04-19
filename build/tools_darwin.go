package build

var tools = map[string]tool{
	"go": {
		Version: "1.18.1",
		URL:     "https://go.dev/dl/go1.18.1.darwin-amd64.tar.gz",
		Hash:    "sha256:63e5035312a9906c98032d9c73d036b6ce54f8632b194228bd08fe3b9fe4ab01",
		Binaries: []string{
			"go/bin/go",
			"go/bin/gofmt",
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
