package testutil

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdknetwork "github.com/cosmos/cosmos-sdk/testutil/network"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/CoreumFoundation/coreum/x/nft/client/cli"
)

func ExecSend(val *sdknetwork.Validator, args []string) (testutil.BufferWriter, error) { //nolint:revive // test helper
	cmd := cli.NewCmdSend()
	return clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
}

func ExecQueryClass(val *sdknetwork.Validator, classID string) (testutil.BufferWriter, error) { //nolint:revive // test helper
	cmd := cli.GetCmdQueryClass()
	var args []string
	args = append(args, classID)
	args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
	return clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
}

func ExecQueryClasses(val *sdknetwork.Validator) (testutil.BufferWriter, error) { //nolint:revive // test helper
	cmd := cli.GetCmdQueryClasses()
	var args []string
	args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
	return clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
}

func ExecQueryNFT(val *sdknetwork.Validator, classID, nftID string) (testutil.BufferWriter, error) { //nolint:revive // test helper
	cmd := cli.GetCmdQueryNFT()
	var args []string
	args = append(args, classID)
	args = append(args, nftID)
	args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
	return clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
}

func ExecQueryNFTs(val *sdknetwork.Validator, classID, owner string) (testutil.BufferWriter, error) { //nolint:revive // test helper
	cmd := cli.GetCmdQueryNFTs()
	var args []string
	args = append(args, fmt.Sprintf("--%s=%s", cli.FlagClassID, classID))
	args = append(args, fmt.Sprintf("--%s=%s", cli.FlagOwner, owner))
	args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
	return clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
}

func ExecQueryOwner(val *sdknetwork.Validator, classID, nftID string) (testutil.BufferWriter, error) { //nolint:revive // test helper
	cmd := cli.GetCmdQueryOwner()
	var args []string
	args = append(args, classID)
	args = append(args, nftID)
	args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
	return clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
}

func ExecQueryBalance(val *sdknetwork.Validator, classID, owner string) (testutil.BufferWriter, error) { //nolint:revive // test helper
	cmd := cli.GetCmdQueryBalance()
	var args []string
	args = append(args, owner)
	args = append(args, classID)
	args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
	return clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
}

func ExecQuerySupply(val *sdknetwork.Validator, classID string) (testutil.BufferWriter, error) { //nolint:revive // test helper
	cmd := cli.GetCmdQuerySupply()
	var args []string
	args = append(args, classID)
	args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
	return clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
}
