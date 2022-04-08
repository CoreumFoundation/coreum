# Token Issuance
##Background
Token issuance is the process of **creating new tokens** that are then **added to the total supply** of the cryptocurrency

Token issuance may also refer to the process of tokenization, in which an asset outside of the cryptocurrency ecosystem is added to the blockchain via a specific crypto token. In such cases, token issuance becomes the process of creating a token, yet not one that belongs to a cryptocurrency, but rather a token that represents an outside asset.

There are 3 moments in which tokens can be issued
- Upon blockchain genesis
- Inflation, at the beginning each block fees are paid to validators. These fees come out of newly minted tokens.
- Out of thin air, through the bank keeper minting function calls.

##Cosmos SDK bank module
The bank module is responsible for handling coin transfers between accounts, minting and handling supply. 

The total Supply of the network is equal to the sum of all coins from all accounts and it is updated every time a Coin is minted or burned.

###Module accounts
The supply functionality introduces a new type of auth.Account which can be used by modules to allocate tokens and in special cases mint or burn tokens.

At a base level these module accounts are capable of sending/receiving tokens to and from auth.Accounts and other module accounts

```go
type ModuleAccount interface {
  auth.Account               // same methods as the Account interface

  GetName() string           // name of the module; used to obtain the address
  GetPermissions() []string  // permissions of module account
  HasPermission(string) bool
}
```

### State
The bank module keeps state of three primary objects:
- Account balances, 
- Denom metadata 
- Total supply of all balances.

### Keepers
The x/bank module accepts a map of addresses that are considered blocklisted from directly and 
explicitly receiving funds through means such as MsgSend and MsgMultiSend and direct API calls like SendCoinsFromModuleToAccount.

Typically, these addresses are module accounts. If these addresses receive funds outside the expected rules of the state machine, 
invariants are likely to be broken and could result in a halted network.

By providing the x/bank module with a blocklisted set of addresses, an error occurs for the operation if a user or 
client attempts to directly or indirectly send funds to a blocklisted account, for example, by using IBC

```go
// Input models transaction input.
message Input {
  string   address                        = 1;
  repeated cosmos.base.v1beta1.Coin coins = 2;
}

// Output models transaction outputs.
message Output {
string   address                        = 1;
repeated cosmos.base.v1beta1.Coin coins = 2;
}
```

###BaseKeeper
The base keeper provides full-permission access: the ability to arbitrary modify any account's balance and mint or burn coins.

Restricted permission to mint per module could be achieved by using baseKeeper with WithMintCoinsRestriction to give specific restrictions to mint 
(e.g. only minting certain denom).

```go
// Keeper defines a module interface that facilitates the transfer of coins
// between accounts.
type Keeper interface {
    SendKeeper

    BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
    DelegateCoins(ctx sdk.Context, delegatorAddr, moduleAccAddr sdk.AccAddress, amt sdk.Coins) error
    DelegateCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
    ExportGenesis(sdk.Context) *types.GenesisState
    GetDenomMetaData(ctx sdk.Context, denom string) (types.Metadata, bool)
    GetPaginatedTotalSupply(ctx sdk.Context, pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error)
    GetSupply(ctx sdk.Context, denom string) sdk.Coin
    InitGenesis(sdk.Context, *types.GenesisState)
    IterateAllDenomMetaData(ctx sdk.Context, cb func(types.Metadata) bool)
    IterateTotalSupply(ctx sdk.Context, cb func(sdk.Coin) bool)
    MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
    SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
    SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
    SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
    SetDenomMetaData(ctx sdk.Context, denomMetaData types.Metadata)
    UndelegateCoins(ctx sdk.Context, moduleAccAddr, delegatorAddr sdk.AccAddress, amt sdk.Coins) error
    UndelegateCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error

    types.QueryServer
}
```

###SendKeeper
The send keeper provides access to account balances and the ability to transfer coins between accounts. 

**The send keeper does not alter the total supply (mint or burn coins).**
```go
// SendKeeper defines a module interface that facilitates the transfer of coins
// between accounts without the possibility of creating coins.
type SendKeeper interface {
    ViewKeeper

    BlockedAddr(addr sdk.AccAddress) bool
    GetParams(ctx sdk.Context) types.Params
    InputOutputCoins(ctx sdk.Context, inputs []types.Input, outputs []types.Output) error
    IsSendEnabledCoin(ctx sdk.Context, coin sdk.Coin) bool
    IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error
    SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
    SetParams(ctx sdk.Context, params types.Params)
}
```

###ViewKeeper
The view keeper provides read-only access to account balances. It does not have balance alteration functionality. 

##Messages
###MsgSend
Send coins from one address to another.
```go
// MsgSend represents a message to send coins from one account to another.
message MsgSend {
  option (gogoproto.equal)           = false;
  option (gogoproto.goproto_getters) = false;

  string   from_address                    = 1 [(gogoproto.moretags) = "yaml:\"from_address\""];
  string   to_address                      = 2 [(gogoproto.moretags) = "yaml:\"to_address\""];
  repeated cosmos.base.v1beta1.Coin amount = 3
      [(gogoproto.nullable) = false, (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins"];
}
```

The message will fail under the following conditions:

- The coins do not have sending enabled
- The to address is restricted

###MsgMultiSend
Send coins from and to a series of different address. If any of the receiving addresses do not correspond to an existing account, a new account is created.
```go
// MsgMultiSend represents an arbitrary multi-in, multi-out send message.
message MsgMultiSend {
  option (gogoproto.equal) = false;

  repeated Input  inputs  = 1 [(gogoproto.nullable) = false];
  repeated Output outputs = 2 [(gogoproto.nullable) = false];
}
```
The message will fail under the following conditions:
- Any of the coins do not have sending enabled
- Any of the to addresses are restricted
- Any of the coins are locked
- The inputs and outputs do not correctly correspond to one another

##Events
The bank module emits the following events:

###Handlers
####MsgSend

| Type     | Attribute Key | Attribute Value |
|----------|---------------|-----------------|
| transfer | recipient     | -               |
| transfer | amount        | -               |
| message  | module        | bank            |
| message  | action        | send            |
| message  | sendedr       |                 |

####MsgMultiSend

| Type     | Attribute Key | Attribute Value |
|----------|---------------|-----------------|
| transfer | recipient     | -               |
| transfer | amount        | -               |
| message  | module        | bank            |
| message  | action        | multisend       |
| message  | sendedr       |                 |

###Keeper events
In addition to handlers events, the bank keeper will produce events when the following methods are called (or any method which ends up calling them)

####MintCoins
```json
{
  "type": "coinbase",
  "attributes": [
    {
      "key": "minter",
      "value": "{{sdk.AccAddress of the module minting coins}}",
      "index": true
    },
    {
      "key": "amount",
      "value": "{{sdk.Coins being minted}}",
      "index": true
    }
  ]
}

```
```json
{
  "type": "coin_received",
  "attributes": [
    {
      "key": "receiver",
      "value": "{{sdk.AccAddress of the module minting coins}}",
      "index": true
    },
    {
      "key": "amount",
      "value": "{{sdk.Coins being received}}",
      "index": true
    }
  ]
}
```

####BurnCoins
```json
{
  "type": "burn",
  "attributes": [
    {
      "key": "burner",
      "value": "{{sdk.AccAddress of the module burning coins}}",
      "index": true
    },
    {
      "key": "amount",
      "value": "{{sdk.Coins being burned}}",
      "index": true
    }
  ]
}
```

```json
{
  "type": "coin_spent",
  "attributes": [
    {
      "key": "spender",
      "value": "{{sdk.AccAddress of the module burning coins}}",
      "index": true
    },
    {
      "key": "amount",
      "value": "{{sdk.Coins being burned}}",
      "index": true
    }
  ]
}
```

####addCoins
```json
{
  "type": "coin_received",
  "attributes": [
    {
      "key": "receiver",
      "value": "{{sdk.AccAddress of the address beneficiary of the coins}}",
      "index": true
    },
    {
      "key": "amount",
      "value": "{{sdk.Coins being received}}",
      "index": true
    }
  ]
}
```

####subUnlockedCoins/DelegateCoins
```json
{
  "type": "coin_spent",
  "attributes": [
    {
      "key": "spender",
      "value": "{{sdk.AccAddress of the address which is spending coins}}",
      "index": true
    },
    {
      "key": "amount",
      "value": "{{sdk.Coins being spent}}",
      "index": true
    }
  ]
}
```
###Parameters

| Key                | Type          | Example                           |
|--------------------|---------------|-----------------------------------|
| SendEnabled        | []SendEnabled | [{denom: "stake", enabled: true}] |
| DefaultSendEnabled | bool          | true                              |

####SendEnabled
The send enabled parameter is an array of SendEnabled entries mapping coin denominations to their send_enabled status. Entries in this list take precedence over the DefaultSendEnabled setting.

####DefaultSendEnabled
The default send enabled value controls send transfer capability for all coin denominations unless specifically included in the array of SendEnabled parameters.