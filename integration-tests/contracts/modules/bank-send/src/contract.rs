use cosmwasm_std::entry_point;
use cosmwasm_std::{Coin, CosmosMsg, Uint128};
use cosmwasm_std::{DepsMut, Env, MessageInfo, Response};
use cw2::set_contract_version;
use cw_ownable::{assert_owner, initialize_owner};

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg};
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
    initialize_owner(deps.storage, deps.api, Some(info.sender.as_ref()))?;

    match msg.amount {
        None => Ok(Response::new()
            .add_attribute("method", "instantiate")
            .add_attribute("owner", info.sender)),
        Some(_) => {
            let msg_res = prepare_withdraw(
                deps,
                info.clone(),
                msg.denom.unwrap(),
                msg.amount.unwrap(),
                msg.recipient.unwrap(),
            );

            match msg_res {
                Err(err) => Err(err),
                Ok(msg) => Ok(Response::new()
                    .add_attribute("method", "instantiate")
                    .add_attribute("owner", info.sender)
                    .add_message(msg))
            }
        }
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::Withdraw {
            denom,
            amount,
            recipient,
        } => try_withdraw(deps, info, denom, amount, recipient),
    }
}

pub fn try_withdraw(
    deps: DepsMut,
    info: MessageInfo,
    denom: String,
    amount: Uint128,
    recipient: String,
) -> Result<Response, ContractError> {
    let recipient_addr = deps.api.addr_validate(&recipient)?;

    let msg_res = prepare_withdraw(
        deps,
        info,
        denom,
        amount,
        recipient,
    );

    match msg_res {
        Err(e) => Err(e),
        Ok(msg) => {
            Ok(Response::new()
                .add_attribute("method", "try_withdraw")
                .add_attribute("to", recipient_addr)
                .add_attribute("amount", amount)
                .add_message(msg))
        }
    }
}

fn prepare_withdraw(
    deps: DepsMut,
    info: MessageInfo,
    denom: String,
    amount: Uint128,
    recipient: String,
) -> Result<CosmosMsg, ContractError> {
    let recipient_addr = deps.api.addr_validate(&recipient)?;

    assert_owner(deps.storage, &info.sender)?;
    if amount == Uint128::zero() {
        return Err(ContractError::InvalidZeroAmount {});
    }

    let transfer_bank_msg = cosmwasm_std::BankMsg::Send {
        to_address: recipient_addr.to_string(),
        amount: vec![Coin {
            amount: amount,
            denom: denom,
        }],
    };

    Ok(transfer_bank_msg.into())
}
