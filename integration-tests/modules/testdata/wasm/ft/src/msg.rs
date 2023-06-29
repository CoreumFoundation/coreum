use cosmwasm_schema::cw_serde;
use cosmwasm_std::Uint128;

#[cw_serde]
pub struct InstantiateMsg {
    pub symbol: String,
    pub subunit: String,
    pub precision: u32,
    pub initial_amount: Uint128,
    pub description: Option<String>,
    pub features: Option<Vec<u32>>,
    pub burn_rate: Option<String>,
    pub send_commission_rate: Option<String>,
}

#[cw_serde]
pub enum ExecuteMsg {
    Mint { amount: u128 },
    Burn { amount: u128 },
    Freeze { account: String, amount: u128 },
    Unfreeze { account: String, amount: u128 },
    GloballyFreeze {},
    GloballyUnfreeze {},
    SetWhitelistedLimit { account: String, amount: u128 },
    // custom message we use to show the submission of multiple messages
    MintAndSend { account: String, amount: u128 },
    UpgradeTokenV1 { ibc_enabled: bool },
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
