use cosmwasm_schema::cw_serde;
use cosmwasm_std::Uint128;

#[cw_serde]
pub struct InstantiateMsg {
    pub denom: String,
}

#[cw_serde]
pub enum ExecuteMsg {
    ExtensionTransfer { amount: Uint128, recipient: String },
}

#[cw_serde]
pub enum QueryMsg {}
