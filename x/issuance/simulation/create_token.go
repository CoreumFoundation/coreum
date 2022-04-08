package simulation

import (
	"math/rand"

	"github.com/coreumfoundation/coreum/coreum/x/issuance/keeper"
	"github.com/coreumfoundation/coreum/coreum/x/issuance/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func SimulateMsgCreateToken(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgCreateToken{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the CreateToken simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "CreateToken simulation not implemented"), nil, nil
	}
}
