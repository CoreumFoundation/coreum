# Counter contract from [IBC Apps](https://github.com/cosmos/ibc-apps/tree/26f3ad8f58e4ffc7769c6766cb42b954181dc100/modules/ibc-hooks)

This contract is a modification of the standard cosmwasm `counter` contract.
Namely, it tracks a counter, _by sender_.
This is a better way to test wasmhooks.

This contract tracks any funds sent to it by adding it to the state under the `sender` key.

This way we can verify that, independently of the sender, the funds will end up under the 
`WasmHooksModuleAccount` address when the contract is executed via an IBC send that goes 
through the wasmhooks module.
