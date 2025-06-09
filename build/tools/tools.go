package tools

import (
	"context"

	"github.com/CoreumFoundation/crust/build/tools"
	"github.com/CoreumFoundation/crust/build/types"
)

const (
	// Cosmovisor is a process manager for Cosmos SDK application binaries
	// that automates application binary switch at chain upgrades.
	Cosmovisor tools.Name = "cosmovisor"
	// MuslCC static cross- and native- musl-based toolchains.
	MuslCC tools.Name = "muslcc"
	// LibWASM is the WASM VM library.
	LibWASM tools.Name = "libwasmvm"
	// Gaia is Cosmos Hub chain.
	Gaia tools.Name = "gaia"
	// Osmosis is Osmosis chain.
	Osmosis tools.Name = "osmosis"
	// Hermes is an Inter-Blockchain Communication (IBC) relayer.
	Hermes tools.Name = "hermes"
	// CoredV500 is an older version of cored used for testing chain upgrades.
	CoredV500 tools.Name = "cored-v5.0.0"
	// Buf is a tool for working with Protocol Buffers.
	Buf tools.Name = "buf"
	// Protoc is the Protocol Buffers compiler.
	Protoc tools.Name = "protoc"
	// ProtocGenDoc is a documentation generator plugin for Google Protocol Buffers.
	ProtocGenDoc tools.Name = "protoc-gen-doc"
	// ProtocGenGRPCGateway is gRPC to JSON proxy generator.
	ProtocGenGRPCGateway tools.Name = "protoc-gen-grpc-gateway"
	// ProtocGenOpenAPIV2 is a tool to generate OpenAPI definitions.
	ProtocGenOpenAPIV2 tools.Name = "protoc-gen-openapiv2"
	// ProtocGenGoCosmos is Protocol Buffers for Go with Gadgets for Cosmos.
	ProtocGenGoCosmos tools.Name = "protoc-gen-gocosmos"
	// ProtocGenBufLint lints Protobuf files.
	ProtocGenBufLint tools.Name = "protoc-gen-buf-lint"
	// ProtocGenBufBreaking detects breaking changes in Protobuf files.
	ProtocGenBufBreaking tools.Name = "protoc-gen-buf-breaking"
)

// Tools list of required binaries and libraries.
var Tools = []tools.Tool{
	// https://github.com/cosmos/cosmos-sdk/releases
	tools.BinaryTool{
		Name:    Cosmovisor,
		Version: "1.6.0",
		Sources: tools.Sources{
			tools.TargetPlatformLinuxAMD64InDocker: {
				URL:  "https://github.com/cosmos/cosmos-sdk/releases/download/cosmovisor%2Fv1.6.0/cosmovisor-v1.6.0-linux-amd64.tar.gz", //nolint:lll // breaking down urls is not beneficial
				Hash: "sha256:844ac6de7aeccb9a05e46fbb5a6c107e5ba896a488ec19e59febb959d6f6a43e",
			},
			tools.TargetPlatformLinuxARM64InDocker: {
				URL:  "https://github.com/cosmos/cosmos-sdk/releases/download/cosmovisor%2Fv1.6.0/cosmovisor-v1.6.0-linux-arm64.tar.gz", //nolint:lll // breaking down urls is not beneficial
				Hash: "sha256:b425ef02ea22f10753b293270ced49d1f158c5f6a8707a51eb20788570a65d63",
			},
		},
		Binaries: map[string]string{
			"bin/cosmovisor": "cosmovisor",
		},
	},

	// http://musl.cc/#binaries
	tools.BinaryTool{
		Name: MuslCC,
		// update GCP bin source when update the version
		Version: "11.2.1",
		Sources: tools.Sources{
			tools.TargetPlatformLinuxAMD64InDocker: {
				URL:  "https://storage.googleapis.com/cored-build-process-binaries/muslcc/11.2.1/x86_64-linux-musl-cross.tgz", //nolint:lll // breaking down urls is not beneficial
				Hash: "sha256:c5d410d9f82a4f24c549fe5d24f988f85b2679b452413a9f7e5f7b956f2fe7ea",
				Binaries: map[string]string{
					"bin/x86_64-linux-musl-gcc": "x86_64-linux-musl-cross/bin/x86_64-linux-musl-gcc",
				},
			},
			tools.TargetPlatformLinuxARM64InDocker: {
				URL:  "https://storage.googleapis.com/cored-build-process-binaries/muslcc/11.2.1/aarch64-linux-musl-cross.tgz", //nolint:lll // breaking down urls is not beneficial
				Hash: "sha256:c909817856d6ceda86aa510894fa3527eac7989f0ef6e87b5721c58737a06c38",
				Binaries: map[string]string{
					"bin/aarch64-linux-musl-gcc": "aarch64-linux-musl-cross/bin/aarch64-linux-musl-gcc",
				},
			},
		},
	},

	// https://github.com/CosmWasm/wasmvm/releases
	// Check compatibility with wasmd before upgrading: https://github.com/CosmWasm/wasmd
	tools.BinaryTool{
		Name:    LibWASM,
		Version: "v2.2.4",
		Sources: tools.Sources{
			tools.TargetPlatformLinuxAMD64InDocker: {
				URL:  "https://github.com/CosmWasm/wasmvm/releases/download/v2.2.4/libwasmvm_muslc.x86_64.a",
				Hash: "sha256:70c989684d2b48ca17bbd55bb694bbb136d75c393c067ef3bdbca31d2b23b578",
				Binaries: map[string]string{
					"lib/libwasmvm_muslc.x86_64.a": "libwasmvm_muslc.x86_64.a",
				},
			},
			tools.TargetPlatformLinuxARM64InDocker: {
				URL:  "https://github.com/CosmWasm/wasmvm/releases/download/v2.2.4/libwasmvm_muslc.aarch64.a",
				Hash: "sha256:27fb13821dbc519119f4f98c30a42cb32429b111b0fdc883686c34a41777488f",
				Binaries: map[string]string{
					"lib/libwasmvm_muslc.aarch64.a": "libwasmvm_muslc.aarch64.a",
				},
			},
			tools.TargetPlatformDarwinAMD64InDocker: {
				URL:  "https://github.com/CosmWasm/wasmvm/releases/download/v2.2.4/libwasmvmstatic_darwin.a",
				Hash: "sha256:43f1341015143c626b634a709872efe848e45ad24444c091496f9c648fd71a67",
				Binaries: map[string]string{
					"lib/libwasmvmstatic_darwin.a": "libwasmvmstatic_darwin.a",
				},
			},
			tools.TargetPlatformDarwinARM64InDocker: {
				URL:  "https://github.com/CosmWasm/wasmvm/releases/download/v2.2.4/libwasmvmstatic_darwin.a",
				Hash: "sha256:43f1341015143c626b634a709872efe848e45ad24444c091496f9c648fd71a67",
				Binaries: map[string]string{
					"lib/libwasmvmstatic_darwin.a": "libwasmvmstatic_darwin.a",
				},
			},
		},
	},

	// https://github.com/cosmos/gaia/releases
	// Before upgrading verify in go.mod that they use the same version of IBC
	tools.BinaryTool{
		Name:    Gaia,
		Version: "v24.0.0",
		Sources: tools.Sources{
			tools.TargetPlatformLinuxAMD64InDocker: {
				URL:  "https://github.com/cosmos/gaia/releases/download/v24.0.0/gaiad-v24.0.0-linux-amd64",
				Hash: "sha256:9c50ed3188d4f7519bfa5a06b8a8311b8ac480e86cbba57722cd62b9c3baeb3a",
				Binaries: map[string]string{
					"bin/gaiad": "gaiad-v24.0.0-linux-amd64",
				},
			},
			tools.TargetPlatformLinuxAMD64: {
				URL:  "https://github.com/cosmos/gaia/releases/download/v24.0.0/gaiad-v24.0.0-linux-amd64",
				Hash: "sha256:9c50ed3188d4f7519bfa5a06b8a8311b8ac480e86cbba57722cd62b9c3baeb3a",
				Binaries: map[string]string{
					"bin/gaiad": "gaiad-v24.0.0-linux-amd64",
				},
			},
			tools.TargetPlatformDarwinAMD64: {
				URL:  "https://github.com/cosmos/gaia/releases/download/v24.0.0/gaiad-v24.0.0-darwin-amd64",
				Hash: "sha256:e1273a56fbb9f75b0b50a6a386e9c1172c9c38c57813c99974d9b2e1c02571fa",
				Binaries: map[string]string{
					"bin/gaiad": "gaiad-v24.0.0-darwin-amd64",
				},
			},
			tools.TargetPlatformDarwinARM64: {
				URL:  "https://github.com/cosmos/gaia/releases/download/v24.0.0/gaiad-v24.0.0-darwin-arm64",
				Hash: "sha256:d92ecc2a873b1e49bc7bb4dd712d91c0be8fa3ac71d93cdd91bd5df8267356f0",
				Binaries: map[string]string{
					"bin/gaiad": "gaiad-v24.0.0-darwin-arm64",
				},
			},
		},
	},

	// https://github.com/osmosis-labs/osmosis/releases
	tools.BinaryTool{
		Name:    Osmosis,
		Version: "29.0.2",
		Sources: tools.Sources{
			tools.TargetPlatformLinuxAMD64InDocker: {
				URL:  "https://github.com/osmosis-labs/osmosis/releases/download/v29.0.2/osmosisd-29.0.2-linux-amd64",
				Hash: "sha256:9276c11c814c8b5731ef7b904c96530c6933e71b02e6eb11f99b4be2b9968c92",
				Binaries: map[string]string{
					"bin/osmosisd": "osmosisd-29.0.2-linux-amd64",
				},
			},
			tools.TargetPlatformLinuxARM64InDocker: {
				URL:  "https://github.com/osmosis-labs/osmosis/releases/download/v29.0.2/osmosisd-29.0.2-linux-arm64",
				Hash: "sha256:e6a3c81ba5ba9da6598582d6c430618a4cb083c7552302412def141f846098d6",
				Binaries: map[string]string{
					"bin/osmosisd": "osmosisd-29.0.2-linux-arm64",
				},
			},
		},
	},

	// https://github.com/informalsystems/hermes/releases
	tools.BinaryTool{
		Name:    Hermes,
		Version: "v1.13.1",
		Sources: tools.Sources{
			tools.TargetPlatformLinuxAMD64InDocker: {
				URL:  "https://github.com/informalsystems/hermes/releases/download/v1.13.1/hermes-v1.13.1-x86_64-unknown-linux-gnu.tar.gz", //nolint:lll // breaking down urls is not beneficial
				Hash: "sha256:effd6b5d698f9e3d78a4e064a407a44987ac941f3faa10411a34e72506b3aa28",
			},
			tools.TargetPlatformLinuxARM64InDocker: {
				// TODO: Change it when hermes CI/CD bug is fixed
				URL:  "https://github.com/informalsystems/hermes/releases/download/v1.13.1/hermes-v1.13.1-x86_64-unknown-linux-gnu.tar.gz", //nolint:lll // breaking down urls is not beneficial
				Hash: "sha256:effd6b5d698f9e3d78a4e064a407a44987ac941f3faa10411a34e72506b3aa28",
			},
		},
		Binaries: map[string]string{
			"bin/hermes": "hermes",
		},
	},

	// https://github.com/CoreumFoundation/coreum/releases
	tools.BinaryTool{
		Name:    CoredV500,
		Version: "v5.0.0",
		Sources: tools.Sources{
			tools.TargetPlatformLinuxAMD64InDocker: {
				URL:  "https://github.com/CoreumFoundation/coreum/releases/download/v5.0.0/cored-linux-amd64",
				Hash: "sha256:6bdbd15f7159e9d0aef62369cb822acf57bad51c4c664f6736b95ecfb8250702",
				Binaries: map[string]string{
					"bin/cored-v5.0.0": "cored-linux-amd64",
				},
			},
			tools.TargetPlatformLinuxARM64InDocker: {
				URL:  "https://github.com/CoreumFoundation/coreum/releases/download/v5.0.0/cored-linux-arm64",
				Hash: "sha256:cee6aeb8043529cce43713f9213c8866f3c4745198a5b9ef44318d6b6728a380",
				Binaries: map[string]string{
					"bin/cored-v5.0.0": "cored-linux-arm64",
				},
			},
			tools.TargetPlatformLinuxAMD64: {
				URL:  "https://github.com/CoreumFoundation/coreum/releases/download/v5.0.0/cored-linux-amd64",
				Hash: "sha256:6bdbd15f7159e9d0aef62369cb822acf57bad51c4c664f6736b95ecfb8250702",
				Binaries: map[string]string{
					"bin/cored-v5.0.0": "cored-linux-amd64",
				},
			},
			tools.TargetPlatformLinuxARM64: {
				URL:  "https://github.com/CoreumFoundation/coreum/releases/download/v5.0.0/cored-linux-arm64",
				Hash: "sha256:cee6aeb8043529cce43713f9213c8866f3c4745198a5b9ef44318d6b6728a380",
				Binaries: map[string]string{
					"bin/cored-v5.0.0": "cored-linux-arm64",
				},
			},
			tools.TargetPlatformDarwinAMD64: {
				URL:  "https://github.com/CoreumFoundation/coreum/releases/download/v5.0.0/cored-darwin-amd64",
				Hash: "sha256:bb32768a1114733dc9a90db70a32dd8cd25828a725a57fc9831301ceb648c0f9",
				Binaries: map[string]string{
					"bin/cored-v5.0.0": "cored-darwin-amd64",
				},
			},
			tools.TargetPlatformDarwinARM64: {
				URL:  "https://github.com/CoreumFoundation/coreum/releases/download/v5.0.0/cored-darwin-arm64",
				Hash: "sha256:8776d26dcd02694c927858da183e81a5cb51ec25aa6fa22d382d8b082ca57cc1",
				Binaries: map[string]string{
					"bin/cored-v5.0.0": "cored-darwin-arm64",
				},
			},
		},
	},

	// https://github.com/bufbuild/buf/releases
	tools.BinaryTool{
		Name:    Buf,
		Version: "v1.28.0",
		Local:   true,
		Sources: tools.Sources{
			tools.TargetPlatformLinuxAMD64: {
				URL:  "https://github.com/bufbuild/buf/releases/download/v1.28.0/buf-Linux-x86_64",
				Hash: "sha256:97dc21ba30be34e2d4d11ee5fa4454453f635c8f5476bfe4cbca58420eb20299",
				Binaries: map[string]string{
					"bin/buf": "buf-Linux-x86_64",
				},
			},
			tools.TargetPlatformDarwinAMD64: {
				URL:  "https://github.com/bufbuild/buf/releases/download/v1.28.0/buf-Darwin-x86_64",
				Hash: "sha256:577fd9fe2e38693b690c88837f70503640897763376195404651f7071493a21a",
				Binaries: map[string]string{
					"bin/buf": "buf-Darwin-x86_64",
				},
			},
			tools.TargetPlatformDarwinARM64: {
				URL:  "https://github.com/bufbuild/buf/releases/download/v1.28.0/buf-Darwin-arm64",
				Hash: "sha256:8e51a9c3e09def469969002c15245cfbf1e7d8f878ddc5205125b8107a22cfbf",
				Binaries: map[string]string{
					"bin/buf": "buf-Darwin-arm64",
				},
			},
		},
	},

	// https://github.com/protocolbuffers/protobuf/releases
	tools.BinaryTool{
		Name:    Protoc,
		Version: "v25.0",
		Local:   true,
		Sources: tools.Sources{
			tools.TargetPlatformLinuxAMD64: {
				URL:  "https://github.com/protocolbuffers/protobuf/releases/download/v25.0/protoc-25.0-linux-x86_64.zip",
				Hash: "sha256:d26c4efe0eae3066bb560625b33b8fc427f55bd35b16f246b7932dc851554e67",
			},
			tools.TargetPlatformDarwinAMD64: {
				URL:  "https://github.com/protocolbuffers/protobuf/releases/download/v25.0/protoc-25.0-osx-x86_64.zip",
				Hash: "sha256:15eefb30ba913e8dc4dd21d2ccb34ce04a2b33124f7d9460e5fd815a5d6459e3",
			},
			tools.TargetPlatformDarwinARM64: {
				URL:  "https://github.com/protocolbuffers/protobuf/releases/download/v25.0/protoc-25.0-osx-aarch_64.zip",
				Hash: "sha256:76a997df5dacc0608e880a8e9069acaec961828a47bde16c06116ed2e570588b",
			},
		},
		Binaries: map[string]string{
			"bin/protoc": "bin/protoc",
		},
	},

	// https://github.com/pseudomuto/protoc-gen-doc/releases/
	tools.BinaryTool{
		Name:    ProtocGenDoc,
		Version: "v1.5.1",
		Local:   true,
		Sources: tools.Sources{
			tools.TargetPlatformLinuxAMD64: {
				URL:  "https://github.com/pseudomuto/protoc-gen-doc/releases/download/v1.5.1/protoc-gen-doc_1.5.1_linux_amd64.tar.gz", //nolint:lll // breaking down urls is not beneficial
				Hash: "sha256:47cd72b07e6dab3408d686a65d37d3a6ab616da7d8b564b2bd2a2963a72b72fd",
			},
			tools.TargetPlatformDarwinAMD64: {
				URL:  "https://github.com/pseudomuto/protoc-gen-doc/releases/download/v1.5.1/protoc-gen-doc_1.5.1_darwin_amd64.tar.gz", //nolint:lll // breaking down urls is not beneficial
				Hash: "sha256:f429e5a5ddd886bfb68265f2f92c1c6a509780b7adcaf7a8b3be943f28e144ba",
			},
			tools.TargetPlatformDarwinARM64: {
				URL:  "https://github.com/pseudomuto/protoc-gen-doc/releases/download/v1.5.1/protoc-gen-doc_1.5.1_darwin_arm64.tar.gz", //nolint:lll // breaking down urls is not beneficial
				Hash: "sha256:6e8c737d9a67a6a873a3f1d37ed8bb2a0a9996f6dcf6701aa1048c7bd798aaf9",
			},
		},
		Binaries: map[string]string{
			"bin/protoc-gen-doc": "protoc-gen-doc",
		},
	},

	// https://github.com/grpc-ecosystem/grpc-gateway/releases
	tools.GoPackageTool{
		Name:    ProtocGenGRPCGateway,
		Version: "v1.16.0",
		Package: "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway",
	},

	// https://github.com/grpc-ecosystem/grpc-gateway/releases
	tools.GoPackageTool{
		Name:    ProtocGenOpenAPIV2,
		Version: "v2.17.0",
		Package: "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2",
	},

	// https://github.com/cosmos/gogoproto/releases
	tools.GoPackageTool{
		Name:    ProtocGenGoCosmos,
		Version: "v1.5.0",
		Package: "github.com/cosmos/gogoproto/protoc-gen-gocosmos",
	},

	// https://github.com/bufbuild/buf/releases
	tools.GoPackageTool{
		Name:    ProtocGenBufLint,
		Version: "v1.26.1",
		Package: "github.com/bufbuild/buf/cmd/protoc-gen-buf-lint",
	},

	// https://github.com/bufbuild/buf/releases
	tools.GoPackageTool{
		Name:    ProtocGenBufBreaking,
		Version: "v1.26.1",
		Package: "github.com/bufbuild/buf/cmd/protoc-gen-buf-breaking",
	},
}

// EnsureBuf ensures that buf is available.
func EnsureBuf(ctx context.Context, deps types.DepsFunc) error {
	return tools.Ensure(ctx, Buf, tools.TargetPlatformLocal)
}

// EnsureProtoc ensures that protoc is available.
func EnsureProtoc(ctx context.Context, deps types.DepsFunc) error {
	return tools.Ensure(ctx, Protoc, tools.TargetPlatformLocal)
}

// EnsureProtocGenDoc ensures that protoc-gen-doc is available.
func EnsureProtocGenDoc(ctx context.Context, deps types.DepsFunc) error {
	return tools.Ensure(ctx, ProtocGenDoc, tools.TargetPlatformLocal)
}

// EnsureProtocGenGRPCGateway ensures that protoc-gen-grpc-gateway is available.
func EnsureProtocGenGRPCGateway(ctx context.Context, deps types.DepsFunc) error {
	return tools.Ensure(ctx, ProtocGenGRPCGateway, tools.TargetPlatformLocal)
}

// EnsureProtocGenGoCosmos ensures that protoc-gen-gocosmos is available.
func EnsureProtocGenGoCosmos(ctx context.Context, deps types.DepsFunc) error {
	return tools.Ensure(ctx, ProtocGenGoCosmos, tools.TargetPlatformLocal)
}

// EnsureProtocGenOpenAPIV2 ensures that protoc-gen-openapiv2 is available.
func EnsureProtocGenOpenAPIV2(ctx context.Context, deps types.DepsFunc) error {
	return tools.Ensure(ctx, ProtocGenOpenAPIV2, tools.TargetPlatformLocal)
}

// EnsureProtocGenBufLint ensures that protoc-gen-buf-lint is available.
func EnsureProtocGenBufLint(ctx context.Context, deps types.DepsFunc) error {
	return tools.Ensure(ctx, ProtocGenBufLint, tools.TargetPlatformLocal)
}

// EnsureProtocGenBufBreaking ensures that protoc-gen-buf-breaking is available.
func EnsureProtocGenBufBreaking(ctx context.Context, deps types.DepsFunc) error {
	return tools.Ensure(ctx, ProtocGenBufBreaking, tools.TargetPlatformLocal)
}

// EnsureBinary installs gaiad binary to crust cache.
func EnsureBinary(ctx context.Context, deps types.DepsFunc) error {
	return tools.Ensure(ctx, Gaia, tools.TargetPlatformLocal)
}
