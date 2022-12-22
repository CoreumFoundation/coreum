use cosmwasm_std::{CosmosMsg, CustomMsg, CustomQuery, Uint128};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum Messages {
    AssetFTMsgIssue {
        symbol: String,
        subunit: String,
        precision: u32,
        initial_amount: Uint128,
    },
}

impl Into<CosmosMsg<Messages>> for Messages {
    fn into(self) -> CosmosMsg<Messages> {
        CosmosMsg::Custom(self)
    }
}

impl CustomMsg for Messages {}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum Queries {
    AssetFTGetToken { denom: String },
}

impl CustomQuery for Queries {}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct FungibleTokenResponse {
    pub issuer: String,
}
