package coreum

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/crust/build/tools"
)

type swaggerDoc struct {
	Swagger     string                                           `json:"swagger"`
	Info        swaggerInfo                                      `json:"info"`
	Consumes    []string                                         `json:"consumes"`
	Produces    []string                                         `json:"produces"`
	Paths       map[string]map[string]map[string]json.RawMessage `json:"paths"`
	Definitions map[string]json.RawMessage                       `json:"definitions"`
}

type swaggerInfo struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

func generateProtoOpenAPI(ctx context.Context, deps build.DepsFunc) error {
	deps(Tidy)

	moduleDirs, includeDirs, err := protoCDirectories(ctx, repoPath, deps)
	if err != nil {
		return err
	}

	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return errors.WithStack(err)
	}

	coreumPath := filepath.Join(absPath, "proto", "coreum")
	cosmosPath := filepath.Join(moduleDirs[cosmosSDKModule], "proto", "cosmos")
	ibcPath := filepath.Join(moduleDirs[cosmosIBCModule], "proto", "ibc")
	generateDirs := []string{
		filepath.Join(coreumPath, "asset", "ft", "v1"),
		filepath.Join(coreumPath, "asset", "nft", "v1"),
		filepath.Join(coreumPath, "customparams", "v1"),
		filepath.Join(coreumPath, "feemodel", "v1"),
		filepath.Join(coreumPath, "nft", "v1beta1"),
		filepath.Join(cosmosPath, "base", "node", "v1beta1"),
		filepath.Join(cosmosPath, "base", "tendermint", "v1beta1"),
		filepath.Join(cosmosPath, "tx", "v1beta1"),
		filepath.Join(cosmosPath, "auth", "v1beta1"),
		filepath.Join(cosmosPath, "authz", "v1beta1"),
		filepath.Join(cosmosPath, "bank", "v1beta1"),
		filepath.Join(cosmosPath, "consensus", "v1"),
		filepath.Join(cosmosPath, "distribution", "v1beta1"),
		filepath.Join(cosmosPath, "evidence", "v1beta1"),
		filepath.Join(cosmosPath, "feegrant", "v1beta1"),
		filepath.Join(cosmosPath, "gov", "v1beta1"),
		filepath.Join(cosmosPath, "gov", "v1"),
		filepath.Join(cosmosPath, "group", "v1"),
		filepath.Join(cosmosPath, "mint", "v1beta1"),
		filepath.Join(cosmosPath, "nft", "v1beta1"),
		filepath.Join(cosmosPath, "slashing", "v1beta1"),
		filepath.Join(cosmosPath, "staking", "v1beta1"),
		filepath.Join(cosmosPath, "upgrade", "v1beta1"),
		filepath.Join(ibcPath, "core", "channel", "v1"),
		filepath.Join(ibcPath, "core", "client", "v1"),
		filepath.Join(ibcPath, "core", "connection", "v1"),
		filepath.Join(ibcPath, "applications", "transfer", "v1"),
		filepath.Join(moduleDirs[cosmWASMModule], "proto", "cosmwasm", "wasm", "v1"),
	}

	return executeOpenAPIProtocCommand(ctx, deps, includeDirs, generateDirs)
}

// executeGoProtocCommand generates go code from proto files.
func executeOpenAPIProtocCommand(ctx context.Context, deps build.DepsFunc, includeDirs, generateDirs []string) error {
	deps(tools.EnsureProtoc, tools.EnsureProtocGenOpenAPIV2)

	outDir, err := os.MkdirTemp("", "")
	if err != nil {
		return errors.WithStack(err)
	}

	defer os.RemoveAll(outDir) //nolint:errcheck // we don't care

	args := []string{
		"--openapiv2_out=logtostderr=true,allow_merge=true,json_names_for_fields=false,fqn_for_openapi_name=true,simple_operation_ids=true,Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:.", //nolint:lll // breaking down this string will make it more complicated.
		"--plugin", must.String(filepath.Abs("bin/protoc-gen-openapiv2")),
	}

	for _, path := range includeDirs {
		args = append(args, "--proto_path", path)
	}

	finalDoc := swaggerDoc{
		Swagger: "2.0",
		Info: swaggerInfo{
			Title:   "title goes here",
			Version: "version goes here",
		},
		Consumes:    []string{"application/json"},
		Produces:    []string{"application/json"},
		Paths:       map[string]map[string]map[string]json.RawMessage{},
		Definitions: map[string]json.RawMessage{},
	}

	for _, dir := range generateDirs {
		var processed bool
		for _, protoFile := range []string{"query.proto", "service.proto"} {
			pf := filepath.Join(dir, protoFile)
			pkg, err := goPackage(pf)
			switch {
			case err == nil:
				processed = true
			case errors.Is(err, os.ErrNotExist):
				continue
			default:
				return err
			}

			dir := filepath.Join(outDir, pkg)
			if err := os.MkdirAll(dir, 0o700); err != nil {
				return err
			}
			args := append([]string{}, args...)
			args = append(args, pf)
			cmd := exec.Command(tools.Path("bin/protoc", tools.TargetPlatformLocal), args...)
			cmd.Dir = dir
			if err := libexec.Exec(ctx, cmd); err != nil {
				return err
			}

			if err := mergeSpecFile(filepath.Join(dir, "apidocs.swagger.json"), pkg, finalDoc); err != nil {
				return err
			}
		}

		if !processed {
			return errors.Errorf("rpc proto files not found in %s", dir)
		}
	}

	f, err := os.OpenFile(
		filepath.Join(repoPath, "docs", "static", "openapi.json"),
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
		0o600,
	)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return errors.WithStack(encoder.Encode(finalDoc))
}

func mergeSpecFile(file string, operationPrefix string, finalDoc swaggerDoc) error {
	var sd swaggerDoc
	f, err := os.Open(file)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&sd); err != nil {
		return errors.WithStack(err)
	}

	const operationIDField = "operationId"
	for k, v := range sd.Paths {
		for opK, opV := range v {
			var opID string
			if err := json.Unmarshal(opV[operationIDField], &opID); err != nil {
				return errors.WithStack(err)
			}
			v[opK][operationIDField] =
				json.RawMessage(fmt.Sprintf(`"%s%s"`, strcase.ToCamel(strings.ReplaceAll(operationPrefix, "/", ".")), opID))
		}
		finalDoc.Paths[k] = v
	}
	for k, v := range sd.Definitions {
		finalDoc.Definitions[k] = v
	}

	return nil
}
