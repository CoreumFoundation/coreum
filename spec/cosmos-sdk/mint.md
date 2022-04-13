# The Minting Mechanism
##Background
The minting mechanism was designed to:

- Allow for a flexible inflation rate determined by market demand **targeting a particular bonded-stake ratio**
- Effect a **balance between market liquidity and staked supply**

In order to best determine the appropriate market rate for inflation rewards, **a moving change rate is used**. 
The moving change rate mechanism ensures that if the % bonded is either over or under the **goal %-bonded**, the inflation rate will adjust to further incentivize or disincentivize being bonded, respectively. 
Setting the goal %-bonded at less than 100% encourages the network to maintain some non-staked tokens which should **help provide some liquidity**.

It can be broken down in the following way:
- If the inflation **rate is below** the goal %-bonded the **inflation rate will increase** until a maximum value is reached
- If the goal % bonded (67% in Cosmos-Hub) is maintained, then the inflation rate will stay constant
- If the inflation **rate is above** the **goal %-bonded** the **inflation rate will decrease** until a minimum value is reached

##State
###Minter
The minter is a space for holding current inflation information.
- Minter: `0x00 -> ProtocolBuffer(minter)`
```go
// Minter represents the minting state.
message Minter {
  // current annual inflation rate
  string inflation = 1
      [(gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec", (gogoproto.nullable) = false];
  // current annual expected provisions
  string annual_provisions = 2 [
    (gogoproto.moretags)   = "yaml:\"annual_provisions\"",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec",
    (gogoproto.nullable)   = false
  ];
}
```

## Params
Minting params are held in the global params store.
- Params: `mint/params -> legacy_amino(params)`

```go
// Params holds parameters for the mint module.
message Params {
  option (gogoproto.goproto_stringer) = false;

  // type of coin to mint
  string mint_denom = 1;
  // maximum annual change in inflation rate
  string inflation_rate_change = 2 [
    (gogoproto.moretags)   = "yaml:\"inflation_rate_change\"",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec",
    (gogoproto.nullable)   = false
  ];
  // maximum inflation rate
  string inflation_max = 3 [
    (gogoproto.moretags)   = "yaml:\"inflation_max\"",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec",
    (gogoproto.nullable)   = false
  ];
  // minimum inflation rate
  string inflation_min = 4 [
    (gogoproto.moretags)   = "yaml:\"inflation_min\"",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec",
    (gogoproto.nullable)   = false
  ];
  // goal of percent bonded atoms
  string goal_bonded = 5 [
    (gogoproto.moretags)   = "yaml:\"goal_bonded\"",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec",
    (gogoproto.nullable)   = false
  ];
  // expected blocks per year
  uint64 blocks_per_year = 6 [(gogoproto.moretags) = "yaml:\"blocks_per_year\""];
}
```

##Begin-Block
Minting parameters are recalculated and inflation paid at the beginning of each block.
###NextInflationRate
The target annual inflation rate is recalculated each block. 
The inflation is also subject to a rate change (positive or negative) depending on the distance from the desired ratio (67%). 

The maximum rate change possible is defined to be 13% per year, however the annual inflation is capped as between 7% and 20%.
```go
NextInflationRate(params Params, bondedRatio sdk.Dec) (inflation sdk.Dec) {
	inflationRateChangePerYear = (1 - bondedRatio/params.GoalBonded) * params.InflationRateChange
	inflationRateChange = inflationRateChangePerYear/blocksPerYr

	// increase the new annual inflation for this next cycle
	inflation += inflationRateChange
	if inflation > params.InflationMax {
		inflation = params.InflationMax
	}
	if inflation < params.InflationMin {
		inflation = params.InflationMin
	}

	return inflation
}
```

###NextAnnualProvisions
Calculate the annual provisions based on current total supply and inflation rate. This parameter is calculated once per block.
```go
NextAnnualProvisions(params Params, totalSupply sdk.Dec) (provisions sdk.Dec) {
    return Inflation * totalSupply
}
```

###BlockProvision
Calculate the provisions generated for each block based on current annual provisions. 
The provisions are then minted by the mint module's **ModuleMinterAccount** and then transferred to the auth's **FeeCollector ModuleAccount**.

```go
BlockProvision(params Params) sdk.Coin {
    provisionAmt = AnnualProvisions/ params.BlocksPerYear
    return sdk.NewCoin(params.MintDenom, provisionAmt.Truncate())
}
```

##Params
The minting module contains the following parameters:

| Key                 | Type        | Example      |
|---------------------|-------------|--------------|
| MintDenom           | String      | "core"       |
| InflationRateChange | string(dec) | "0.13000000" |
| InflationMax        | string(dec) | "0.2000000"  |
| InflationMin        | string(dec) | "0.07000000" |
| GoalBonded          | string(dec)|   "0.670000000000000000"|
| BlocksPerYear       | string(uint64)| "6311520"             |

##Events
###BeginBlocker

| Type    | Attribute Key | Attribute Value |
|---------|---------------|-----------------|
| mint    | bonded_ratio     | -               |
| mint    | inflation        | -               |
| mint    | annual_provisions| -               |
| mint    | amount        | -               |

##Client
###CLI
A user can query and interact with the mint module using the CLI.

####Query
The query commands allow users to query mint state.
```shell
cored query mint --help
```

#####annual-provisions
The annual-provisions command allow users to query the current minting annual provisions value
```shell
cored query mint annual-provisions [flags]
```

#####inflation
```shell
cored query mint inflation [flags]
```

#####inflation
```shell
cored query mint params [flags]
```


