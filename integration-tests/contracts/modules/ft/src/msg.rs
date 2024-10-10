use coreum_wasm_sdk::types::coreum::asset::ft::v1::ExtensionIssueSettings;
use cosmwasm_schema::cw_serde;
use cosmwasm_std::Uint128;

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
    UpgradeTokenV1 {
        ibc_enabled: bool,
    },
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
