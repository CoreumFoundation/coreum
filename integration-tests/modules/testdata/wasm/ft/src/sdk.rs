use cosmwasm_std::{Coin, CosmosMsg, CustomMsg, CustomQuery, Uint128};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

// ********** Transactions **********

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum AssetFTMessages {
    AssetFTMsgIssue {
        symbol: String,
        subunit: String,
        precision: u32,
        initial_amount: Uint128,
        description: Option<String>,
        features: Option<Vec<u32>>,
        burn_rate: Option<String>,
        send_commission_rate: Option<String>,
    },
    AssetFTMsgMint {
        coin: Coin,
    },
    AssetFTMsgBurn {
        coin: Coin,
    },
    AssetFTMsgFreeze {
        account: String,
        coin: Coin,
    },
    AssetFTMsgUnfreeze {
        account: String,
        coin: Coin,
    },
    AssetFTMsgGloballyFreeze {
        denom: String,
    },
    AssetFTMsgGloballyUnfreeze {
        denom: String
    },
    AssetFTMsgSetWhitelistedLimit {
        account: String,
        coin: Coin,
    },
}

impl Into<CosmosMsg<AssetFTMessages>> for AssetFTMessages {
    fn into(self) -> CosmosMsg<AssetFTMessages> {
        CosmosMsg::Custom(self)
    }
}

impl CustomMsg for AssetFTMessages {}

// ********** Queries **********

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct AssetFTToken {
    pub denom: String,
    pub issuer: String,
    pub symbol: String,
    pub subunit: String,
    pub precision: u32,
    pub description: Option<String>,
    pub features: Option<Vec<u32>>,
    pub burn_rate: String,
    pub send_commission_rate: String,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct AssetFTTokenResponse {
    pub token: AssetFTToken,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct AssetFTFrozenBalanceResponse {
    pub balance: Coin,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct AssetFTWhitelistedBalanceResponse {
    pub balance: Coin,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum AssetFTQueries {
    AssetFTQueryToken { denom: String },
    AssetFTQueryFrozenBalance {
        account: String,
        denom: String,
    },
    AssetFTQueryWhitelistedBalance {
        account: String,
        denom: String,
    },
}

impl CustomQuery for AssetFTQueries {}
