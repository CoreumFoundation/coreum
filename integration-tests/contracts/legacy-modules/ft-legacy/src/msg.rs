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
    pub burn_rate: String,
    pub send_commission_rate: String,
    pub uri: Option<String>,
    pub uri_hash: Option<String>,
}

#[cw_serde]
pub enum ExecuteMsg {
    Mint {
        amount: u128,
        recipient: Option<String>,
    },
    Burn {
        amount: u128,
    },
    Freeze {
        account: String,
        amount: u128,
    },
    Unfreeze {
        account: String,
        amount: u128,
    },
    SetFrozen {
        account: String,
        amount: u128,
    },
    GloballyFreeze {},
    GloballyUnfreeze {},
    SetWhitelistedLimit {
        account: String,
        amount: u128,
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
