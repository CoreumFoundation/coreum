use coreum_wasm_sdk::core::CoreumResult;

#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{AnyMsg, Binary, CosmosMsg, DepsMut, Env, MessageInfo, Response};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg};

const CONTRACT_NAME: &str = env!("CARGO_PKG_NAME");
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    Ok(Response::new()
        .add_attribute("contract", CONTRACT_NAME)
        .add_attribute("action", "instantiate"))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    _deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: ExecuteMsg,
) -> CoreumResult<ContractError> {
    match msg {
        ExecuteMsg::Stargate { type_url, value } => execute_stargate_message(type_url, value),
    }
}

fn execute_stargate_message(type_url: String, value: Binary) -> CoreumResult<ContractError> {
    let msg = CosmosMsg::Any(AnyMsg { type_url, value });

    Ok(Response::new()
        .add_attribute("method", "execute_authz_stargate")
        .add_message(msg))
}
