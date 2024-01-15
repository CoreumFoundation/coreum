use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Addr, Coin};
use cw_storage_plus::Map;

// We keep NFT offers here, key is (class_id, id)
pub const NFT_OFFERS: Map<(String, String), Offer> = Map::new("nft_offers");

#[cw_serde]
pub struct Offer {
    pub address: Addr,
    pub price: Coin,
}
