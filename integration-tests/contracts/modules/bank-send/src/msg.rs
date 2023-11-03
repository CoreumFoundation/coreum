use cosmwasm_schema::cw_serde;
use cosmwasm_std::Uint128;

#[cw_serde]
pub struct InstantiateMsg {
    pub denom: Option<String>,
    pub amount: Option<Uint128>,
    pub recipient: Option<String>,
}

#[cw_serde]
pub enum ExecuteMsg {
    Withdraw {
        denom: String,
        amount: Uint128,
        recipient: String,
    },
}
