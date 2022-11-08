use cosmwasm_std::{CustomMsg, entry_point};
use cosmwasm_std::{Addr, CosmosMsg, Uint128};
use cosmwasm_std::{DepsMut, Env, MessageInfo, Response, StdError};
use cw2::set_contract_version;
use cw_storage_plus::Item;

use schemars::JsonSchema;
use serde::{Deserialize, Serialize};
use thiserror::Error;

// version info for migration info
const CONTRACT_NAME: &str = "crates.io:hello";
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
) -> Result<Response<FungibleTokenMsg>, ContractError> {
    match msg {
        ExecuteMsg::Issue {
            symbol,
            amount,
            recipient,
        } => create_token(deps, info, symbol, amount, recipient),
    }
}

fn create_token(
    deps: DepsMut,
    info: MessageInfo,
    symbol: String,
    amount: Uint128,
    recipient: String,
) -> Result<Response<FungibleTokenMsg>, ContractError> {
    let recipient_addr = deps.api.addr_validate(&recipient)?;

    if amount == Uint128::zero() {
        return Err(ContractError::InvalidZeroAmount {});
    }

    let state = STATE.load(deps.storage)?;
    if info.sender != state.owner {
        return Err(ContractError::Unauthorized {});
    }

    let create_token_msg = FungibleTokenMsg::MsgIssueFungibleToken {
        symbol: symbol,
        recipient: recipient_addr.to_string(),
        initial_amount: amount,
    };

    // let create_token_cosmos_msg: CosmosMsg<FungibleTokenMsg> = create_token_msg.into();

    let res: Response<FungibleTokenMsg> = Response::new()
        .add_attribute("method", "create_token")
        .add_attribute("recipient", recipient_addr)
        .add_attribute("amount", amount)
        .add_message(create_token_msg);
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

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum FungibleTokenMsg {
    MsgIssueFungibleToken {
        symbol: String,
        recipient: String,
        initial_amount: Uint128,
    },
}

impl Into<CosmosMsg<FungibleTokenMsg>> for FungibleTokenMsg {
    fn into(self) -> CosmosMsg<FungibleTokenMsg> {
        CosmosMsg::Custom(self)
    }
}

impl CustomMsg for FungibleTokenMsg {}
