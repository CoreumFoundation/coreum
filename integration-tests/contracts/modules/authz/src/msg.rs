use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Addr, Uint128, Binary};

#[cw_serde]
pub struct InstantiateMsg {
    pub granter: Addr,
}

#[cw_serde]
pub enum ExecuteMsg {
    Transfer {
        address: String,
        amount: Uint128,
        denom: String,
    },
    Stargate {
        type_url: String,
        value: Binary,
    },
}
