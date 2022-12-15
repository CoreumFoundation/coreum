use cosmwasm_std::{CosmosMsg, CustomMsg, CustomQuery, Uint128};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum FungibleTokenMsg {
    MsgIssueFungibleToken {
        symbol: String,
        subunit: String,
        precision: u32,
        initial_amount: Uint128,
    },
}

impl Into<CosmosMsg<FungibleTokenMsg>> for FungibleTokenMsg {
    fn into(self) -> CosmosMsg<FungibleTokenMsg> {
        CosmosMsg::Custom(self)
    }
}

impl CustomMsg for FungibleTokenMsg {}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum FungibleTokenQuery {
    FungibleToken { denom: String },
}

impl CustomQuery for FungibleTokenQuery {}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct FungibleTokenResponse {
    pub issuer: String,
}
