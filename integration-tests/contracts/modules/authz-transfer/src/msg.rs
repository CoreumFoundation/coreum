use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Addr, Uint128};

#[cw_serde]
pub struct InstantiateMsg {
    // Granter used for transfering native tokens from the contract in behalf of him.
    pub granter: Addr,
}

#[cw_serde]
pub enum ExecuteMsg {
    Transfer {
        address: String,
        amount: Uint128,
        denom: String,
    },
}
