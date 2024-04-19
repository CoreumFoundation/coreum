use cosmwasm_schema::cw_serde;
use cosmwasm_std::Uint128;
use std::collections::HashMap;

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum ExecuteMsg {
    ExtensionTransfer {
        amount: Uint128,
        recipients: HashMap<String, Uint128>,
    },
}

#[cw_serde]
pub enum QueryMsg {}
