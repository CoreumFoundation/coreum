use cosmwasm_std::{Coin, CosmosMsg, CustomMsg, CustomQuery, Empty, QueryRequest, Uint128};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

// ********** Generic payload **********

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum CoreumMsg<T = Empty> {
    Msg {
        name: String,
        payload: T,
    }
}

impl Into<CosmosMsg<CoreumMsg>> for CoreumMsg {
    fn into(self) -> CosmosMsg<CoreumMsg> {
        CosmosMsg::Custom(self)
    }
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum CoreumQuery<T = Empty> {
    Query {
        name: String,
        payload: T,
    }
}

impl Into<QueryRequest<CoreumQuery>> for CoreumQuery {
    fn into(self) -> QueryRequest<CoreumQuery> {
        QueryRequest::Custom(self)
    }
}

// ********** Transactions **********

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum AssetFTMsg {
    MsgIssue {
        symbol: String,
        subunit: String,
        precision: u32,
        initial_amount: Uint128,
        description: Option<String>,
        features: Option<Vec<u32>>,
        burn_rate: Option<String>,
        send_commission_rate: Option<String>,
    },
    MsgMint {
        coin: Coin,
    },
    MsgBurn {
        coin: Coin,
    },
    MsgFreeze {
        account: String,
        coin: Coin,
    },
    MsgUnfreeze {
        account: String,
        coin: Coin,
    },
    MsgGloballyFreeze {
        denom: String,
    },
    MsgGloballyUnfreeze {
        denom: String
    },
    MsgSetWhitelistedLimit {
        account: String,
        coin: Coin,
    },
}

impl AssetFTMsg {
    pub fn to_coreum_msg(self) -> CoreumMsg<AssetFTMsg> {
        match self {
            AssetFTMsg::MsgIssue { .. } => CoreumMsg::Msg {
                name: "coreum.asset.ft.v1.MsgIssue".into(),
                payload: self.clone(),
            },
            AssetFTMsg::MsgMint { .. } => CoreumMsg::Msg {
                name: "coreum.asset.ft.v1.MsgMint".into(),
                payload: self.clone(),
            },
            AssetFTMsg::MsgBurn { .. } => CoreumMsg::Msg {
                name: "coreum.asset.ft.v1.MsgBurn".into(),
                payload: self,
            },
            AssetFTMsg::MsgFreeze { .. } => CoreumMsg::Msg {
                name: "coreum.asset.ft.v1.MsgFreeze".into(),
                payload: self,
            },
            AssetFTMsg::MsgUnfreeze { .. } => CoreumMsg::Msg {
                name: "coreum.asset.ft.v1.MsgUnfreeze".into(),
                payload: self,
            },
            AssetFTMsg::MsgGloballyFreeze { .. } => CoreumMsg::Msg {
                name: "coreum.asset.ft.v1.MsgGloballyFreeze".into(),
                payload: self,
            },
            AssetFTMsg::MsgGloballyUnfreeze { .. } => CoreumMsg::Msg {
                name: "coreum.asset.ft.v1.MsgGloballyUnfreeze".into(),
                payload: self,
            },
            AssetFTMsg::MsgSetWhitelistedLimit { .. } => CoreumMsg::Msg {
                name: "coreum.asset.ft.v1.MsgSetWhitelistedLimit".into(),
                payload: self,
            },
        }
    }
}

impl Into<CosmosMsg<AssetFTMsg>> for AssetFTMsg {
    fn into(self) -> CosmosMsg<AssetFTMsg> {
        CosmosMsg::Custom(self)
    }
}

impl From<CoreumMsg<AssetFTMsg>> for CosmosMsg<CoreumMsg<AssetFTMsg>> {
    fn from(item: CoreumMsg<AssetFTMsg>) -> Self {
        CosmosMsg::Custom(item)
    }
}

impl CustomMsg for CoreumMsg<AssetFTMsg> {}

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
pub enum AssetFTQuery {
    Token { denom: String },
    FrozenBalance {
        account: String,
        denom: String,
    },
    WhitelistedBalance {
        account: String,
        denom: String,
    },
}

impl AssetFTQuery {
    pub fn to_coreum_query(self) -> CoreumQuery<AssetFTQuery> {
        match self {
            AssetFTQuery::Token { .. } => CoreumQuery::Query {
                name: "coreum.asset.ft.v1.QueryTokenRequest".into(),
                payload: self.clone(),
            },
            AssetFTQuery::FrozenBalance { .. } => CoreumQuery::Query {
                name: "coreum.asset.ft.v1.QueryFrozenBalanceRequest".into(),
                payload: self.clone(),
            },
            AssetFTQuery::WhitelistedBalance { .. } => CoreumQuery::Query {
                name: "coreum.asset.ft.v1.QueryWhitelistedBalanceRequest".into(),
                payload: self.clone(),
            },
        }
    }
}

impl CustomQuery for CoreumQuery<AssetFTQuery> {}
