use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Addr, Coin};

#[cw_serde]
pub struct InstantiateMsg {
    //Granter used for transfering tokens from the contract in behalf of him.
    pub granter: Addr,
}

#[cw_serde]
pub enum ExecuteMsg {
    Transfer {
        address: Addr,
        amount: u64,
        denom: String,
    },
    OfferNft {
        class_id: String,
        id: String,
        price: Coin,
    },
    AcceptNftOffer {
        class_id: String,
        id: String,
    },
}
