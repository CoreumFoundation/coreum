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
    Clawback {
        account: String,
        amount: u128,
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
pub enum QueryMsg {}
