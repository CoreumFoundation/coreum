use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Addr, Coin};

use cw_storage_plus::{Item, Map};

// We keep the granter address here
pub const GRANTER: Item<Addr> = Item::new("granter");
// We keep NFT offers here, key is (class_id, id)
pub const NFT_OFFERS: Map<(String, String), Offer> = Map::new("nft_offers");

#[cw_serde]
pub struct Offer {
    pub address: Addr,
    pub price: Coin,
}