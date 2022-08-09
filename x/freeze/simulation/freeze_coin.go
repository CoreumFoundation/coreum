package simulation

import (
	"math/rand"

	"github.com/CoreumFoundation/coreum/x/freeze/keeper"
	"github.com/CoreumFoundation/coreum/x/freeze/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func SimulateMsgFreezeCoin(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgFreezeCoin{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the FreezeCoin simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "FreezeCoin simulation not implemented"), nil, nil
	}
}
