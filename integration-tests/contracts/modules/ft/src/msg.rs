use coreum_wasm_sdk::types::coreum::asset::ft::v1::DexSettings;
use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Coin, Uint128};

#[cw_serde]
pub struct InstantiateMsg {
    pub symbol: String,
    pub subunit: String,
    pub precision: u32,
    pub initial_amount: Uint128,
    pub description: Option<String>,
    pub features: Option<Vec<i32>>,
    pub burn_rate: String,
    pub send_commission_rate: String,
    pub uri: Option<String>,
    pub uri_hash: Option<String>,
    pub extension_settings: Option<ExtensionIssueSettings>,
    pub dex_settings: Option<DexSettings>,
}

#[cw_serde]
pub struct ExtensionIssueSettings {
    pub code_id: u64,
    pub label: String,
    pub funds: Vec<Coin>,
    pub issuance_msg: IssuanceMsg,
}

#[cw_serde]
pub struct IssuanceMsg {
    pub extra_data: String,
}

#[cw_serde]
pub enum ExecuteMsg {
    Mint {
        amount: Uint128,
        recipient: Option<String>,
    },
    Burn {
        amount: Uint128,
    },
    Freeze {
        account: String,
        amount: Uint128,
    },
    Unfreeze {
        account: String,
        amount: Uint128,
    },
    SetFrozen {
        account: String,
        amount: Uint128,
    },
    GloballyFreeze {},
    GloballyUnfreeze {},
    SetWhitelistedLimit {
        account: String,
        amount: Uint128,
    },
    Clawback {
        account: String,
        amount: Uint128,
    },
    TransferAdmin {
        account: String,
    },
    ClearAdmin {},
}

#[cw_serde]
pub enum QueryMsg {
    Params {},
    Token {},
    Tokens { issuer: String },
    Balance { account: String },
    FrozenBalances { account: String },
    FrozenBalance { account: String },
    WhitelistedBalances { account: String },
    WhitelistedBalance { account: String },
}
