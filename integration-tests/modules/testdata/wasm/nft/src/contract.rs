use coreum_wasm_sdk::assetnft;
use coreum_wasm_sdk::core::{CoreumMsg, CoreumQueries};
use coreum_wasm_sdk::nft;
use cosmwasm_std::{entry_point, to_binary, Binary, Deps, QueryRequest, StdResult};
use cosmwasm_std::{DepsMut, Env, MessageInfo, Response, StdError};
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
    pub name: String,
    pub symbol: String,
    pub description: Option<String>,
    pub uri: Option<String>,
    pub uri_hash: Option<String>,
    pub data: Option<String>,
    pub features: Option<Vec<u32>>,
    pub royalty_rate: Option<String>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct State {
    pub owner: String,
    pub class_id: String,
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
    Mint {
        id: String,
        uri: Option<String>,
        uri_hash: Option<String>,
        data: Option<String>,
    },
    Burn {
        id: String,
    },
    Freeze {
        id: String,
    },
    Unfreeze {
        id: String,
    },
    AddToWhitelist {
        id: String,
        account: String,
    },
    RemoveFromWhitelist {
        id: String,
        account: String,
    },
    Send {
        id: String,
        receiver: String,
    },
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response<CoreumMsg>, ContractError> {
    match msg {
        ExecuteMsg::Mint {
            id,
            uri,
            uri_hash,
            data,
        } => mint(deps, info, id, uri, uri_hash, data),
        ExecuteMsg::Burn { id } => burn(deps, info, id),
        ExecuteMsg::Freeze { id } => freeze(deps, info, id),
        ExecuteMsg::Unfreeze { id } => unfreeze(deps, info, id),
        ExecuteMsg::AddToWhitelist { id, account } => add_to_white_list(deps, info, id, account),
        ExecuteMsg::RemoveFromWhitelist { id, account } => {
            remove_from_white_list(deps, info, id, account)
        }
        ExecuteMsg::Send { id, receiver } => send(deps, info, id, receiver),
    }
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum QueryMsg {
    Class {},
    Frozen { id: String },
    Whitelisted { id: String, account: String },
    Balance { owner: String },
    Owner { id: String },
    Supply {},
    Nft { id: String }, // we use Nft not NFT since NFT is decoded as n_f_t
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps<CoreumQueries>, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Class {} => to_binary(&class(deps)?),
        QueryMsg::Frozen { id } => to_binary(&frozen(deps, id)?),
        QueryMsg::Whitelisted { id, account } => to_binary(&whitelisted(deps, id, account)?),
        QueryMsg::Balance { owner } => to_binary(&balance(deps, owner)?),
        QueryMsg::Owner { id } => to_binary(&owner(deps, id)?),
        QueryMsg::Supply {} => to_binary(&supply(deps)?),
        QueryMsg::Nft { id } => to_binary(&nft(deps, id)?),
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
    let issue_msg = CoreumMsg::AssetNFT(assetnft::Msg::IssueClass {
        name: msg.name,
        symbol: msg.symbol.clone(),
        description: msg.description,
        uri: msg.uri,
        uri_hash: msg.uri_hash,
        data: msg.data,
        features: msg.features,
        royalty_rate: msg.royalty_rate,
    });

    let class_id = format!("{}-{}", msg.symbol, env.contract.address).to_lowercase();

    let state = State {
        owner: info.sender.into(),
        class_id,
    };
    STATE.save(deps.storage, &state)?;

    Ok(Response::new()
        .add_attribute("owner", state.owner)
        .add_attribute("class_id", state.class_id)
        .add_message(issue_msg))
}

// ********** Transactions **********

fn mint(
    deps: DepsMut,
    info: MessageInfo,
    id: String,
    uri: Option<String>,
    uri_hash: Option<String>,
    data: Option<String>,
) -> Result<Response<CoreumMsg>, ContractError> {
    let state = STATE.load(deps.storage)?;
    if info.sender != state.owner {
        return Err(ContractError::Unauthorized {});
    }

    let msg = CoreumMsg::AssetNFT(assetnft::Msg::Mint {
        class_id: state.class_id.clone(),
        id: id.clone(),
        uri,
        uri_hash,
        data,
    });

    Ok(Response::new()
        .add_attribute("method", "mint")
        .add_attribute("class_id", state.class_id)
        .add_attribute("id", id)
        .add_message(msg))
}

fn burn(
    deps: DepsMut,
    info: MessageInfo,
    id: String,
) -> Result<Response<CoreumMsg>, ContractError> {
    let state = STATE.load(deps.storage)?;
    if info.sender != state.owner {
        return Err(ContractError::Unauthorized {});
    }

    let msg = CoreumMsg::AssetNFT(assetnft::Msg::Burn {
        class_id: state.class_id.clone(),
        id: id.clone(),
    });

    Ok(Response::new()
        .add_attribute("method", "burn")
        .add_attribute("class_id", state.class_id)
        .add_attribute("id", id)
        .add_message(msg))
}

fn freeze(
    deps: DepsMut,
    info: MessageInfo,
    id: String,
) -> Result<Response<CoreumMsg>, ContractError> {
    let state = STATE.load(deps.storage)?;
    if info.sender != state.owner {
        return Err(ContractError::Unauthorized {});
    }

    let msg = CoreumMsg::AssetNFT(assetnft::Msg::Freeze {
        class_id: state.class_id.clone(),
        id: id.clone(),
    });

    Ok(Response::new()
        .add_attribute("method", "freeze")
        .add_attribute("class_id", state.class_id)
        .add_attribute("id", id)
        .add_message(msg))
}

fn unfreeze(
    deps: DepsMut,
    info: MessageInfo,
    id: String,
) -> Result<Response<CoreumMsg>, ContractError> {
    let state = STATE.load(deps.storage)?;
    if info.sender != state.owner {
        return Err(ContractError::Unauthorized {});
    }

    let msg = CoreumMsg::AssetNFT(assetnft::Msg::Unfreeze {
        class_id: state.class_id.clone(),
        id: id.clone(),
    });

    Ok(Response::new()
        .add_attribute("method", "unfreeze")
        .add_attribute("class_id", state.class_id)
        .add_attribute("id", id)
        .add_message(msg))
}

fn add_to_white_list(
    deps: DepsMut,
    info: MessageInfo,
    id: String,
    account: String,
) -> Result<Response<CoreumMsg>, ContractError> {
    let state = STATE.load(deps.storage)?;
    if info.sender != state.owner {
        return Err(ContractError::Unauthorized {});
    }

    let msg = CoreumMsg::AssetNFT(assetnft::Msg::AddToWhitelist {
        class_id: state.class_id.clone(),
        id: id.clone(),
        account,
    });

    Ok(Response::new()
        .add_attribute("method", "add_to_white_list")
        .add_attribute("class_id", state.class_id)
        .add_attribute("id", id)
        .add_message(msg))
}

fn remove_from_white_list(
    deps: DepsMut,
    info: MessageInfo,
    id: String,
    account: String,
) -> Result<Response<CoreumMsg>, ContractError> {
    let state = STATE.load(deps.storage)?;
    if info.sender != state.owner {
        return Err(ContractError::Unauthorized {});
    }

    let msg = CoreumMsg::AssetNFT(assetnft::Msg::RemoveFromWhitelist {
        class_id: state.class_id.clone(),
        id: id.clone(),
        account,
    });

    Ok(Response::new()
        .add_attribute("method", "remove_from_white_list")
        .add_attribute("class_id", state.class_id)
        .add_attribute("id", id)
        .add_message(msg))
}

fn send(
    deps: DepsMut,
    info: MessageInfo,
    id: String,
    receiver: String,
) -> Result<Response<CoreumMsg>, ContractError> {
    let state = STATE.load(deps.storage)?;
    if info.sender != state.owner {
        return Err(ContractError::Unauthorized {});
    }

    let msg = CoreumMsg::NFT(nft::Msg::Send {
        class_id: state.class_id.clone(),
        id: id.clone(),
        receiver,
    });

    Ok(Response::new()
        .add_attribute("method", "send")
        .add_attribute("class_id", state.class_id)
        .add_attribute("id", id)
        .add_message(msg))
}

// ********** Queries **********

// ********** AssetNFT **********

fn class(deps: Deps<CoreumQueries>) -> StdResult<assetnft::ClassResponse> {
    let state = STATE.load(deps.storage)?;
    let request: QueryRequest<CoreumQueries> =
        CoreumQueries::AssetNFT(assetnft::Query::Class { id: state.class_id }).into();
    let res: assetnft::ClassResponse = deps.querier.query(&request)?;
    Ok(res)
}

fn frozen(deps: Deps<CoreumQueries>, id: String) -> StdResult<assetnft::FrozenResponse> {
    let state = STATE.load(deps.storage)?;
    let request: QueryRequest<CoreumQueries> = CoreumQueries::AssetNFT(assetnft::Query::Frozen {
        id,
        class_id: state.class_id,
    })
    .into();
    let res: assetnft::FrozenResponse = deps.querier.query(&request)?;
    Ok(res)
}

fn whitelisted(
    deps: Deps<CoreumQueries>,
    id: String,
    account: String,
) -> StdResult<assetnft::WhitelistedResponse> {
    let state = STATE.load(deps.storage)?;
    let request: QueryRequest<CoreumQueries> =
        CoreumQueries::AssetNFT(assetnft::Query::Whitelisted {
            id,
            class_id: state.class_id,
            account,
        })
        .into();
    let res: assetnft::WhitelistedResponse = deps.querier.query(&request)?;
    Ok(res)
}

// ********** NFT **********

fn balance(deps: Deps<CoreumQueries>, owner: String) -> StdResult<nft::BalanceResponse> {
    let state = STATE.load(deps.storage)?;
    let request: QueryRequest<CoreumQueries> = CoreumQueries::NFT(nft::Query::Balance {
        class_id: state.class_id,
        owner,
    })
    .into();
    let res: nft::BalanceResponse = deps.querier.query(&request)?;
    Ok(res)
}

fn owner(deps: Deps<CoreumQueries>, id: String) -> StdResult<nft::OwnerResponse> {
    let state = STATE.load(deps.storage)?;
    let request: QueryRequest<CoreumQueries> = CoreumQueries::NFT(nft::Query::Owner {
        class_id: state.class_id,
        id,
    })
    .into();
    let res: nft::OwnerResponse = deps.querier.query(&request)?;
    Ok(res)
}

fn supply(deps: Deps<CoreumQueries>) -> StdResult<nft::SupplyResponse> {
    let state = STATE.load(deps.storage)?;
    let request: QueryRequest<CoreumQueries> = CoreumQueries::NFT(nft::Query::Supply {
        class_id: state.class_id,
    })
    .into();
    let res: nft::SupplyResponse = deps.querier.query(&request)?;
    Ok(res)
}

fn nft(deps: Deps<CoreumQueries>, id: String) -> StdResult<nft::NFTResponse> {
    let state = STATE.load(deps.storage)?;
    let request: QueryRequest<CoreumQueries> = CoreumQueries::NFT(nft::Query::NFT {
        class_id: state.class_id,
        id,
    })
    .into();
    let res: nft::NFTResponse = deps.querier.query(&request)?;
    Ok(res)
}
