use cosmwasm_schema::cw_serde;
use cosmwasm_std::Coin;

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum ExecuteMsg {
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
