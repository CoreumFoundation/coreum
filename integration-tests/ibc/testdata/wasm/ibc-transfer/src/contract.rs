use cosmwasm_std::entry_point;
use cosmwasm_std::{Coin, CosmosMsg, DepsMut, Env, IbcMsg, IbcTimeout, MessageInfo, Response};
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

    Ok(Response::new().add_attribute("method", "instantiate"))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    _deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::Transfer {
            channel_id,
            to_address,
            amount,
            timeout,
        } => transfer(channel_id, to_address, amount, timeout),
    }
}

pub fn transfer(
    channel_id: String,
    to_address: String,
    amount: Coin,
    timeout: IbcTimeout,
) -> Result<Response, ContractError> {
    let ibc_transfer_msg: CosmosMsg = IbcMsg::Transfer {
        channel_id,
        to_address,
        amount,
        timeout,
    }
    .into();
    let res = Response::new()
        .add_attribute("method", "transfer")
        .add_message(ibc_transfer_msg);
    Ok(res)
}
