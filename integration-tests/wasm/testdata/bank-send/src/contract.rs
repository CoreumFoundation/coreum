use cosmwasm_std::entry_point;
use cosmwasm_std::{Addr, Coin, CosmosMsg, Uint128};
use cosmwasm_std::{DepsMut, Env, MessageInfo, Response, StdError};
use cw2::set_contract_version;
use cw_storage_plus::Item;

use schemars::JsonSchema;
use serde::{Deserialize, Serialize};
use thiserror::Error;

// version info for migration info
const CONTRACT_NAME: &str = "creates.io:bank-send";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct State {
    pub owner: Addr,
}

pub const STATE: Item<State> = Item::new("state");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    _msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    let state = State {
        owner: info.sender.clone(),
    };

    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
    STATE.save(deps.storage, &state)?;

    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("owner", info.sender))
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

    if amount == Uint128::zero() {
        return Err(ContractError::InvalidZeroAmount {});
    }

    let state = STATE.load(deps.storage)?;
    if info.sender != state.owner {
        return Err(ContractError::Unauthorized {});
    }

    let transfer_bank_msg = cosmwasm_std::BankMsg::Send {
        to_address: recipient_addr.to_string(),
        amount: vec![Coin {
            amount: amount,
            denom: denom,
        }],
    };

    let transfer_bank_cosmos_msg: CosmosMsg = transfer_bank_msg.into();

    let res = Response::new()
        .add_attribute("method", "try_withdraw")
        .add_attribute("to", recipient_addr)
        .add_attribute("amount", amount)
        .add_message(transfer_bank_cosmos_msg);
    Ok(res)
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct InstantiateMsg {}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum ExecuteMsg {
    Withdraw {
        denom: String,
        amount: Uint128,
        recipient: String,
    },
}

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("Unauthorized")]
    Unauthorized {},

    #[error("Invalid zero amount")]
    InvalidZeroAmount {},

    #[error("Custom Error val: {val:?}")]
    CustomError { val: String },
    // Add any other custom errors you like here.
    // Look at https://docs.rs/thiserror/1.0.21/thiserror/ for details.
}
