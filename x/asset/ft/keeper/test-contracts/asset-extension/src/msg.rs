use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{Uint128};
use coreum_wasm_sdk::types::cosmos::base::v1beta1::Coin;

#[cw_serde]
pub struct InstantiateMsg {
    pub denom: String,
    pub issuance_msg: IssuanceMsg,
}

#[cw_serde]
pub struct IssuanceMsg {
    pub extra_data: Option<String>,
}

#[cw_serde]
pub enum ExecuteMsg {}

#[cw_serde]
pub struct DEXOrder {
    pub creator: String,
    #[serde(rename = "type")]
    pub order_type: String,
    pub id: String,
    pub base_denom: String,
    pub quote_denom: String,
    pub price: Option<String>,
    pub quantity: Uint128,
    pub side: String,
}

#[cw_serde]
pub enum SudoMsg {
    ExtensionTransfer {
        recipient: String,
        sender: String,
        transfer_amount: Uint128,
        commission_amount: Uint128,
        burn_amount: Uint128,
        context: TransferContext,
    },
    ExtensionPlaceOrder {
        order: DEXOrder,
        expected_to_spend: Coin,
        expected_to_receive: Coin,
    },
}

#[cw_serde]
pub struct TransferContext {
    pub sender_is_smart_contract: bool,
    pub recipient_is_smart_contract: bool,
    pub ibc_purpose: IBCPurpose,
}

#[cw_serde]
pub enum IBCPurpose {
    None,
    Out,
    In,
    Ack,
    Timeout,
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(QueryIssuanceMsgResponse)]
    QueryIssuanceMsg {},
}

#[cw_serde]
pub struct QueryIssuanceMsgResponse {
    pub test: Option<String>,
}
