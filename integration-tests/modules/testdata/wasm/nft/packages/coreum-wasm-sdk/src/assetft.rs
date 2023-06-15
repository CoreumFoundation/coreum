use cosmwasm_schema::QueryResponses;
use cosmwasm_std::{Coin, Uint128};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use crate::pagination::{PageRequest, PageResponse};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct Params {
    pub issue_fee: Coin,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct ParamsResponse {
    pub params: Params,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct Token {
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
pub struct TokensResponse {
    pub pagination: PageResponse,
    pub tokens: Vec<Token>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct TokenResponse {
    pub token: Token,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct BalanceResponse {
    pub balance: String,
    pub whitelisted: String,
    pub frozen: String,
    pub locked: String,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct FrozenBalancesResponse {
    pub pagination: PageResponse,
    pub balances: Vec<Coin>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct FrozenBalanceResponse {
    pub balance: Coin,
}
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct WhitelistedBalancesResponse {
    pub pagination: PageResponse,
    pub balances: Vec<Coin>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct WhitelistedBalanceResponse {
    pub balance: Coin,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum Msg {
    Issue {
        symbol: String,
        subunit: String,
        precision: u32,
        initial_amount: Uint128,
        description: Option<String>,
        features: Option<Vec<u32>>,
        burn_rate: Option<String>,
        send_commission_rate: Option<String>,
    },
    Mint {
        coin: Coin,
    },
    Burn {
        coin: Coin,
    },
    Freeze {
        account: String,
        coin: Coin,
    },
    Unfreeze {
        account: String,
        coin: Coin,
    },
    GloballyFreeze {
        denom: String,
    },
    GloballyUnfreeze {
        denom: String,
    },
    SetWhitelistedLimit {
        account: String,
        coin: Coin,
    },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema, QueryResponses)]
pub enum Query {
    #[returns(ParamsResponse)]
    Params {},

    #[returns(TokensResponse)]
    Tokens {
        pagination: Option<PageRequest>,
        issuer: String,
    },

    #[returns(TokenResponse)]
    Token { denom: String },

    #[returns(BalanceResponse)]
    Balance { account: String, denom: String },

    #[returns(FrozenBalancesResponse)]
    FrozenBalances {
        pagination: Option<PageRequest>,
        account: String,
    },

    #[returns(FrozenBalanceResponse)]
    FrozenBalance { account: String, denom: String },

    #[returns(WhitelistedBalancesResponse)]
    WhitelistedBalances {
        pagination: Option<PageRequest>,
        account: String,
    },

    #[returns(WhitelistedBalanceResponse)]
    WhitelistedBalance { account: String, denom: String },
}
