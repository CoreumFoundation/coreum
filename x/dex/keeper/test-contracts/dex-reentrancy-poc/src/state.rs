use cw_storage_plus::Item;

pub const DENOM: Item<String> = Item::new("state");
pub const EXTRA_DATA: Item<String> = Item::new("extradata");
