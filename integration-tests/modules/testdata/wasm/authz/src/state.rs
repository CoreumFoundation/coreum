use cosmwasm_std::Addr;

use cw_storage_plus::Item;

// We keep the granter address here
pub const GRANTER: Item<Addr> = Item::new("granter");
