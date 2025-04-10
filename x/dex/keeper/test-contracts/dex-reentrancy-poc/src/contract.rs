use coreum_wasm_sdk::types::coreum::dex::v1::{MsgPlaceOrder, OrderType, Side, TimeInForce};

use crate::error::ContractError;
use crate::msg::{
    DEXOrder, ExecuteMsg, InstantiateMsg, QueryIssuanceMsgResponse, QueryMsg, SudoMsg,
    TransferContext,
};
use crate::state::{DENOM, EXTRA_DATA};
use coreum_wasm_sdk::types::cosmos::bank::v1beta1::MsgSend;
use coreum_wasm_sdk::types::cosmos::base::v1beta1::Coin;
use cosmwasm_std::{entry_point, to_json_binary, CosmosMsg};
use cosmwasm_std::{Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult, Uint128};
use cw2::set_contract_version;
use std::string::ToString;

// version info for migration info
const CONTRACT_NAME: &str = env!("CARGO_PKG_NAME");
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    DENOM.save(deps.storage, &msg.denom)?;
    EXTRA_DATA.save(
        deps.storage,
        &msg.issuance_msg.extra_data.unwrap_or_default(),
    )?;

    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("owner", info.sender))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    _deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {}
}

#[entry_point]
pub fn sudo(deps: DepsMut, env: Env, msg: SudoMsg) -> Result<Response, ContractError> {
    match msg {
        SudoMsg::ExtensionTransfer {
            sender,
            recipient,
            transfer_amount,
            commission_amount,
            burn_amount,
            context,
        } => sudo_extension_transfer(
            deps,
            env,
            transfer_amount,
            sender,
            recipient,
            commission_amount,
            burn_amount,
            context,
        ),
        SudoMsg::ExtensionPlaceOrder {
            order,
            spent,
            received,
        } => sudo_extension_place_order(deps, env, order, spent, received),
    }
}

pub fn sudo_extension_transfer(
    deps: DepsMut,
    env: Env,
    amount: Uint128,
    _: String,
    recipient: String,
    _: Uint128,
    _: Uint128,
    _: TransferContext,
) -> Result<Response, ContractError> {
    if amount.is_zero() {
        return Err(ContractError::InvalidAmountError {});
    }

    let denom = DENOM.load(deps.storage)?;
    if recipient == env.contract.address.as_str() {
        return Ok(Response::new()
            .add_attribute("method", "execute_transfer")
            .add_attribute("skip_checks", "self_recipient"));
    }
    let transfer_msg = MsgSend {
        from_address: env.contract.address.to_string(),
        to_address: recipient.to_string(),
        amount: [Coin {
            denom: denom.to_string(),
            amount: amount.to_string(),
        }]
        .to_vec(),
    };
    let response = Response::new()
        .add_attribute("method", "execute_transfer")
        .add_message(CosmosMsg::Any(transfer_msg.to_any()));
    Ok(response)
}

pub fn sudo_extension_place_order(
    _: DepsMut,
    env: Env,
    order: DEXOrder,
    _: Coin,
    _: Coin,
) -> Result<Response, ContractError> {
    if order.id == "hackid0" {
        let order = MsgPlaceOrder {
            sender: env.contract.address.to_string(),
            r#type: OrderType::Limit as i32,
            id: "hackid1".to_string().into(),
            base_denom: order.base_denom,
            quote_denom: order.quote_denom,
            price: order.price.unwrap(),
            quantity: "1500000".into(),
            side: Side::Buy as i32,
            good_til: None,
            time_in_force: TimeInForce::Gtc as i32,
        };
        Ok(Response::new()
            .add_message(CosmosMsg::Any(order.clone().to_any()))
            .add_attribute("order:", format!("{:?}", order)))
    } else {
        Ok(Response::new().add_attribute("method", "extension_place_order"))
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::QueryIssuanceMsg {} => query_issuance_msg(deps),
    }
}

fn query_issuance_msg(deps: Deps) -> StdResult<Binary> {
    let test = EXTRA_DATA.load(deps.storage).ok();
    let resp = QueryIssuanceMsgResponse { test };
    to_json_binary(&resp)
}
