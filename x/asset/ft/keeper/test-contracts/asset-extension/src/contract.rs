use cosmwasm_std::{entry_point, StdError};
use cosmwasm_std::{Binary, Coin, Deps, DepsMut, Env, MessageInfo, Response, StdResult, Uint128};
use cw2::set_contract_version;

use crate::error::ContractError;

use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::DENOM;

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

    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("owner", info.sender))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::ExtensionTransfer { amount, recipient } => {
            execute_extension_transfer(deps, env, info, amount, recipient)
        }
    }
}

pub fn execute_extension_transfer(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    amount: Uint128,
    recipient: String,
) -> Result<Response, ContractError> {
    // TODO(milad) check that amount is present in the attached funds, and attached funds
    // is enough to cover the transfer.
    // TODO remove this if statement.
    // This check is intended for POC testing, it must be replaced with a more
    // meaningful check.
    if amount == Uint128::new(7) {
        return Err(ContractError::Std(StdError::generic_err(
            "7 is not allowed",
        )));
    }

    let denom = DENOM.load(deps.storage)?;
    let transfer_msg = cosmwasm_std::BankMsg::Send {
        to_address: recipient,
        amount: vec![Coin {
            amount: amount,
            denom: denom.clone(),
        }],
    };

    Ok(Response::new()
        .add_attribute("method", "execute_transfer")
        .add_message(transfer_msg))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(_deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {}
}
