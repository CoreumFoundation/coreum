use coreum_wasm_sdk::types::coreum::asset::ft::v1::{DexSettings, ExtensionIssueSettings};
use coreum_wasm_sdk::types::coreum::dex::v1::{MsgPlaceOrder, Order};
use cosmwasm_schema::cw_serde;
use cosmwasm_std::Uint128;

#[cw_serde]
pub struct InstantiateMsg {
    pub symbol: String,
    pub subunit: String,
    pub precision: u32,
    pub initial_amount: Uint128,
    pub description: Option<String>,
    pub features: Option<Vec<i32>>,
    pub burn_rate: String,
    pub send_commission_rate: String,
    pub uri: Option<String>,
    pub uri_hash: Option<String>,
    pub extension_settings: Option<ExtensionIssueSettings>,
    pub dex_settings: Option<DexSettings>,
}

#[cw_serde]
pub enum ExecuteMsg {
    PlaceOrder { order: MsgPlaceOrder },
    CancelOrder { order_id: String },
    CancelOrdersByDenom { account: String, denom: String },
}

#[cw_serde]
pub enum QueryMsg {
    Params {},
    Order {
        acc: String,
        order_id: String,
    },
    Orders {
        creator: String,
    },
    OrderBooks {},
    OrderBookOrders {
        base_denom: String,
        quote_denom: String,
        side: i32,
    },
    AccountDenomOrdersCount {
        account: String,
        denom: String,
    },
}
