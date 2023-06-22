use cosmwasm_schema::cw_serde;
use cosmwasm_std::Addr;

#[cw_serde]
pub struct InstantiateMsg {
    pub granter: Addr,
}

#[cw_serde]
pub enum ExecuteMsg {
    Transfer {
        address: Addr,
        amount: u64,
        denom: String,
    },
}
