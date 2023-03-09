use coreum_wasm_sdk::assetft;
use coreum_wasm_sdk::core::{CoreumMsg, CoreumQueries};
use cosmwasm_std::{entry_point, to_binary, Binary, Deps, QueryRequest, StdResult};
use cosmwasm_std::{Coin, DepsMut, Env, MessageInfo, Response, StdError, SubMsg, Uint128};
use cw2::set_contract_version;
use cw_storage_plus::Item;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};
use thiserror::Error;

// version info for migration info
const CONTRACT_NAME: &str = "creates.io:ft";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct InstantiateMsg {
    pub symbol: String,
    pub subunit: String,
    pub precision: u32,
    pub initial_amount: Uint128,
    pub description: Option<String>,
    pub features: Option<Vec<u32>>,
    pub burn_rate: Option<String>,
    pub send_commission_rate: Option<String>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct State {
    pub owner: String,
    pub denom: String,
}

pub const STATE: Item<State> = Item::new("state");

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("Unauthorized")]
    Unauthorized {},

    #[error("Invalid input")]
    InvalidInput(String),

    #[error("Custom Error val: {val:?}")]
    CustomError { val: String },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum ExecuteMsg {
    Mint { amount: u128 },
    Burn { amount: u128 },
    Freeze { account: String, amount: u128 },
    Unfreeze { account: String, amount: u128 },
    GloballyFreeze {},
    GloballyUnfreeze {},
    SetWhitelistedLimit { account: String, amount: u128 },
    // custom message we use to show the submission of multiple messages
    MintAndSend { account: String, amount: u128 },
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response<CoreumMsg>, ContractError> {
    match msg {
        ExecuteMsg::Mint { amount } => mint(deps, info, amount),
        ExecuteMsg::Burn { amount } => burn(deps, info, amount),
        ExecuteMsg::Freeze { account, amount } => freeze(deps, info, account, amount),
        ExecuteMsg::Unfreeze { account, amount } => unfreeze(deps, info, account, amount),
        ExecuteMsg::GloballyFreeze {} => globally_freeze(deps, info),
        ExecuteMsg::GloballyUnfreeze {} => globally_unfreeze(deps, info),
        ExecuteMsg::SetWhitelistedLimit { account, amount } => {
            set_whitelisted_limit(deps, info, account, amount)
        }
        ExecuteMsg::MintAndSend { account, amount } => mint_and_send(deps, info, account, amount),
    }
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum QueryMsg {
    Token {},
    FrozenBalance { account: String },
    WhitelistedBalance { account: String },
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps<CoreumQueries>, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Token {} => to_binary(&token(deps)?),
        QueryMsg::FrozenBalance { account } => to_binary(&frozen_balance(deps, account)?),
        QueryMsg::WhitelistedBalance { account } => to_binary(&whitelisted_balance(deps, account)?),
    }
}

// ********** Instantiate **********

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response<CoreumMsg>, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
    let issue_msg = CoreumMsg::AssetFT(assetft::Msg::Issue {
        symbol: msg.symbol,
        subunit: msg.subunit.clone(),
        precision: msg.precision,
        initial_amount: msg.initial_amount,
        description: msg.description,
        features: msg.features,
        burn_rate: msg.burn_rate,
        send_commission_rate: msg.send_commission_rate,
    });

    let denom = format!("{}-{}", msg.subunit, env.contract.address).to_lowercase();

    let state = State {
        owner: info.sender.into(),
        denom,
    };
    STATE.save(deps.storage, &state)?;

    Ok(Response::new()
        .add_attribute("owner", state.owner)
        .add_attribute("denom", state.denom)
        .add_message(issue_msg))
}

// ********** Transactions **********

fn mint(
    deps: DepsMut,
    info: MessageInfo,
    amount: u128,
) -> Result<Response<CoreumMsg>, ContractError> {
    let state = STATE.load(deps.storage)?;
    if info.sender != state.owner {
        return Err(ContractError::Unauthorized {});
    }

    let msg = CoreumMsg::AssetFT(assetft::Msg::Mint {
        coin: Coin::new(amount, state.denom.clone()),
    });

    Ok(Response::new()
        .add_attribute("method", "mint")
        .add_attribute("denom", state.denom)
        .add_attribute("amount", amount.to_string())
        .add_message(msg))
}

fn burn(
    deps: DepsMut,
    info: MessageInfo,
    amount: u128,
) -> Result<Response<CoreumMsg>, ContractError> {
    let state = STATE.load(deps.storage)?;
    if info.sender != state.owner {
        return Err(ContractError::Unauthorized {});
    }

    let msg = CoreumMsg::AssetFT(assetft::Msg::Burn {
        coin: Coin::new(amount, state.denom.clone()),
    });

    Ok(Response::new()
        .add_attribute("method", "burn")
        .add_attribute("denom", state.denom)
        .add_attribute("amount", amount.to_string())
        .add_message(msg))
}

fn freeze(
    deps: DepsMut,
    info: MessageInfo,
    account: String,
    amount: u128,
) -> Result<Response<CoreumMsg>, ContractError> {
    let state = STATE.load(deps.storage)?;
    if info.sender != state.owner {
        return Err(ContractError::Unauthorized {});
    }

    let msg = CoreumMsg::AssetFT(assetft::Msg::Freeze {
        account,
        coin: Coin::new(amount, state.denom.clone()),
    });

    Ok(Response::new()
        .add_attribute("method", "freeze")
        .add_attribute("denom", state.denom)
        .add_attribute("amount", amount.to_string())
        .add_message(msg))
}

fn unfreeze(
    deps: DepsMut,
    info: MessageInfo,
    account: String,
    amount: u128,
) -> Result<Response<CoreumMsg>, ContractError> {
    let state = STATE.load(deps.storage)?;
    if info.sender != state.owner {
        return Err(ContractError::Unauthorized {});
    }

    let msg = CoreumMsg::AssetFT(assetft::Msg::Unfreeze {
        account,
        coin: Coin::new(amount, state.denom.clone()),
    });

    Ok(Response::new()
        .add_attribute("method", "unfreeze")
        .add_attribute("denom", state.denom)
        .add_attribute("amount", amount.to_string())
        .add_message(msg))
}

fn globally_freeze(deps: DepsMut, info: MessageInfo) -> Result<Response<CoreumMsg>, ContractError> {
    let state = STATE.load(deps.storage)?;
    if info.sender != state.owner {
        return Err(ContractError::Unauthorized {});
    }

    let msg = CoreumMsg::AssetFT(assetft::Msg::GloballyFreeze {
        denom: state.denom.clone(),
    });

    Ok(Response::new()
        .add_attribute("method", "globally_freeze")
        .add_attribute("denom", state.denom)
        .add_message(msg))
}

fn globally_unfreeze(
    deps: DepsMut,
    info: MessageInfo,
) -> Result<Response<CoreumMsg>, ContractError> {
    let state = STATE.load(deps.storage)?;
    if info.sender != state.owner {
        return Err(ContractError::Unauthorized {});
    }

    let msg = CoreumMsg::AssetFT(assetft::Msg::GloballyUnfreeze {
        denom: state.denom.clone(),
    });

    Ok(Response::new()
        .add_attribute("method", "globally_unfreeze")
        .add_attribute("denom", state.denom)
        .add_message(msg))
}

fn set_whitelisted_limit(
    deps: DepsMut,
    info: MessageInfo,
    account: String,
    amount: u128,
) -> Result<Response<CoreumMsg>, ContractError> {
    let state = STATE.load(deps.storage)?;
    if info.sender != state.owner {
        return Err(ContractError::Unauthorized {});
    }

    let msg = CoreumMsg::AssetFT(assetft::Msg::SetWhitelistedLimit {
        account,
        coin: Coin::new(amount, state.denom.clone()),
    });

    Ok(Response::new()
        .add_attribute("method", "set_whitelisted_limit")
        .add_attribute("denom", state.denom)
        .add_attribute("amount", amount.to_string())
        .add_message(msg))
}

fn mint_and_send(
    deps: DepsMut,
    info: MessageInfo,
    account: String,
    amount: u128,
) -> Result<Response<CoreumMsg>, ContractError> {
    let state = STATE.load(deps.storage)?;
    if info.sender != state.owner {
        return Err(ContractError::Unauthorized {});
    }

    let mint_msg = SubMsg::new(CoreumMsg::AssetFT(assetft::Msg::Mint {
        coin: Coin::new(amount, state.denom.clone()),
    }));

    let send_msg = SubMsg::new(cosmwasm_std::BankMsg::Send {
        to_address: account.to_string(),
        amount: vec![Coin {
            amount: amount.into(),
            denom: state.denom.clone(),
        }],
    });

    Ok(Response::new()
        .add_attribute("method", "mint_and_send")
        .add_attribute("denom", state.denom)
        .add_attribute("amount", amount.to_string())
        .add_submessages([mint_msg, send_msg]))
}

// ********** Queries **********

fn token(deps: Deps<CoreumQueries>) -> StdResult<assetft::TokenResponse> {
    let state = STATE.load(deps.storage)?;
    let request: QueryRequest<CoreumQueries> =
        CoreumQueries::AssetFT(assetft::Query::Token { denom: state.denom }).into();
    let res: assetft::TokenResponse = deps.querier.query(&request)?;
    Ok(res)
}

fn frozen_balance(
    deps: Deps<CoreumQueries>,
    account: String,
) -> StdResult<assetft::FrozenBalanceResponse> {
    let state = STATE.load(deps.storage)?;
    let request: QueryRequest<CoreumQueries> =
        CoreumQueries::AssetFT(assetft::Query::FrozenBalance {
            denom: state.denom,
            account,
        })
            .into();
    let res: assetft::FrozenBalanceResponse = deps.querier.query(&request)?;
    Ok(res)
}

fn whitelisted_balance(
    deps: Deps<CoreumQueries>,
    account: String,
) -> StdResult<assetft::WhitelistedBalanceResponse> {
    let state = STATE.load(deps.storage)?;
    let request: QueryRequest<CoreumQueries> =
        CoreumQueries::AssetFT(assetft::Query::WhitelistedBalance {
            denom: state.denom,
            account,
        })
            .into();
    let res: assetft::WhitelistedBalanceResponse = deps.querier.query(&request)?;
    Ok(res)
}
