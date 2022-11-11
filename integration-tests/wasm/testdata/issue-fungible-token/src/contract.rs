use cosmwasm_std::{entry_point};
use cosmwasm_std::{DepsMut, Env, MessageInfo, Response, StdError, Uint128};
use cw2::set_contract_version;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};
use thiserror::Error;
use crate::sdk;

// version info for migration info
const CONTRACT_NAME: &str = "creates.io:issue-fungible-token";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    _msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("owner", info.sender))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response<sdk::FungibleTokenMsg>, ContractError> {
    match msg {
        ExecuteMsg::Issue {
            symbol,
            amount,
            recipient,
        } => issue_token(deps, symbol, amount, recipient),
    }
}

fn issue_token(
    deps: DepsMut,
    symbol: String,
    amount: Uint128,
    recipient: String,
) -> Result<Response<sdk::FungibleTokenMsg>, ContractError> {
    let recipient_addr = deps.api.addr_validate(&recipient)?;

    if amount == Uint128::zero() {
        return Err(ContractError::InvalidZeroAmount {});
    }

    let issue_token_msg = sdk::FungibleTokenMsg::MsgIssueFungibleToken {
        symbol: symbol.clone(),
        recipient: recipient_addr.to_string(),
        initial_amount: amount,
    };

    let res: Response<sdk::FungibleTokenMsg> = Response::new()
        .add_attribute("method", "issue_token")
        .add_attribute("symbol", symbol)
        .add_attribute("recipient", recipient_addr)
        .add_attribute("amount", amount)
        .add_message(issue_token_msg);
    Ok(res)
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct InstantiateMsg {}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum ExecuteMsg {
    Issue {
        symbol: String,
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
