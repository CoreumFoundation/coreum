use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Addr, Binary, Coin, Uint128};

#[cw_serde]
pub struct InstantiateMsg {
    // Granter used for transfering native tokens from the contract in behalf of him.
    // This is only used for Transfer ExecuteMsg.
    pub granter: Addr,
}

#[cw_serde]
pub enum ExecuteMsg {
    Transfer {
        address: String,
        amount: Uint128,
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
    Stargate {
        type_url: String,
        value: Binary,
    },
}
