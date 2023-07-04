use coreum_wasm_sdk::assetnft::{
    self, ClassResponse, ClassesResponse, FrozenResponse, ParamsResponse,
    WhitelistedAccountsForNFTResponse, WhitelistedResponse,
};
use coreum_wasm_sdk::core::{CoreumMsg, CoreumQueries, CoreumResult};
use coreum_wasm_sdk::nft;
use coreum_wasm_sdk::pagination::PageRequest;
use cosmwasm_std::{
    entry_point, to_binary, Binary, Deps, DepsMut, Env, MessageInfo, QueryRequest, Response,
    StdResult,
};
use cw2::set_contract_version;
use cw_ownable::{assert_owner, initialize_owner};

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::CLASS_ID;
// version info for migration info
const CONTRACT_NAME: &str = env!("CARGO_PKG_NAME");
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

// ********** Instantiate **********

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> CoreumResult<ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
    initialize_owner(deps.storage, deps.api, Some(info.sender.as_ref()))?;

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

    CLASS_ID.save(deps.storage, &class_id)?;

    Ok(Response::new()
        .add_attribute("owner", info.sender)
        .add_attribute("class_id", class_id)
        .add_message(issue_msg))
}

// ********** Execute **********

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> CoreumResult<ContractError> {
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

// ********** Transactions **********

fn mint(
    deps: DepsMut,
    info: MessageInfo,
    id: String,
    uri: Option<String>,
    uri_hash: Option<String>,
    data: Option<Binary>,
) -> CoreumResult<ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let msg = CoreumMsg::AssetNFT(assetnft::Msg::Mint {
        class_id: class_id.clone(),
        id: id.clone(),
        uri,
        uri_hash,
        data,
    });

    Ok(Response::new()
        .add_attribute("method", "mint")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(msg))
}

fn burn(deps: DepsMut, info: MessageInfo, id: String) -> CoreumResult<ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let msg = CoreumMsg::AssetNFT(assetnft::Msg::Burn {
        class_id: class_id.clone(),
        id: id.clone(),
    });

    Ok(Response::new()
        .add_attribute("method", "burn")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(msg))
}

fn freeze(deps: DepsMut, info: MessageInfo, id: String) -> CoreumResult<ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let msg = CoreumMsg::AssetNFT(assetnft::Msg::Freeze {
        class_id: class_id.clone(),
        id: id.clone(),
    });

    Ok(Response::new()
        .add_attribute("method", "freeze")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(msg))
}

fn unfreeze(deps: DepsMut, info: MessageInfo, id: String) -> CoreumResult<ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let msg = CoreumMsg::AssetNFT(assetnft::Msg::Unfreeze {
        class_id: class_id.clone(),
        id: id.clone(),
    });

    Ok(Response::new()
        .add_attribute("method", "unfreeze")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(msg))
}

fn add_to_white_list(
    deps: DepsMut,
    info: MessageInfo,
    id: String,
    account: String,
) -> CoreumResult<ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let msg = CoreumMsg::AssetNFT(assetnft::Msg::AddToWhitelist {
        class_id: class_id.clone(),
        id: id.clone(),
        account,
    });

    Ok(Response::new()
        .add_attribute("method", "add_to_white_list")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(msg))
}

fn remove_from_white_list(
    deps: DepsMut,
    info: MessageInfo,
    id: String,
    account: String,
) -> CoreumResult<ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let msg = CoreumMsg::AssetNFT(assetnft::Msg::RemoveFromWhitelist {
        class_id: class_id.clone(),
        id: id.clone(),
        account,
    });

    Ok(Response::new()
        .add_attribute("method", "remove_from_white_list")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(msg))
}

fn send(
    deps: DepsMut,
    info: MessageInfo,
    id: String,
    receiver: String,
) -> CoreumResult<ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let msg = CoreumMsg::NFT(nft::Msg::Send {
        class_id: class_id.clone(),
        id: id.clone(),
        receiver,
    });

    Ok(Response::new()
        .add_attribute("method", "send")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(msg))
}

// ********** Queries **********

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps<CoreumQueries>, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Params {} => to_binary(&query_params(deps)?),
        QueryMsg::Class {} => to_binary(&query_class(deps)?),
        QueryMsg::Classes { issuer } => to_binary(&query_classes(deps, issuer)?),
        QueryMsg::Frozen { id } => to_binary(&query_frozen(deps, id)?),
        QueryMsg::Whitelisted { id, account } => to_binary(&query_whitelisted(deps, id, account)?),
        QueryMsg::WhitelistedAccountsForNft { id } => {
            to_binary(&query_whitelisted_accounts_for_nft(deps, id)?)
        }
        QueryMsg::Balance { owner } => to_binary(&query_balance(deps, owner)?),
        QueryMsg::Owner { id } => to_binary(&query_owner(deps, id)?),
        QueryMsg::Supply {} => to_binary(&query_supply(deps)?),
        QueryMsg::Nft { id } => to_binary(&query_nft(deps, id)?),
        QueryMsg::Nfts { owner } => to_binary(&query_nfts(deps, owner)?),
        QueryMsg::ClassNft {} => to_binary(&query_nft_class(deps)?),
        QueryMsg::ClassesNft {} => to_binary(&query_nft_classes(deps)?),
    }
}

fn query_params(deps: Deps<CoreumQueries>) -> StdResult<ParamsResponse> {
    let request: QueryRequest<CoreumQueries> =
        CoreumQueries::AssetNFT(assetnft::Query::Params {}).into();
    let res = deps.querier.query(&request)?;
    Ok(res)
}

fn query_class(deps: Deps<CoreumQueries>) -> StdResult<ClassResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let request: QueryRequest<CoreumQueries> =
        CoreumQueries::AssetNFT(assetnft::Query::Class { id: class_id }).into();
    let res = deps.querier.query(&request)?;
    Ok(res)
}

fn query_classes(deps: Deps<CoreumQueries>, issuer: String) -> StdResult<ClassesResponse> {
    let mut pagination = None;
    let mut classes = vec![];
    let mut res: ClassesResponse;
    loop {
        let request = CoreumQueries::AssetNFT(assetnft::Query::Classes {
            pagination,
            issuer: issuer.clone(),
        })
        .into();
        res = deps.querier.query(&request)?;
        classes.append(&mut res.classes);
        if res.pagination.next_key.is_none() {
            break;
        } else {
            pagination = Some(PageRequest {
                key: res.pagination.next_key,
                offset: None,
                limit: None,
                count_total: None,
                reverse: None,
            })
        }
    }
    let res = ClassesResponse {
        pagination: res.pagination,
        classes,
    };
    Ok(res)
}

fn query_frozen(deps: Deps<CoreumQueries>, id: String) -> StdResult<FrozenResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let request: QueryRequest<CoreumQueries> =
        CoreumQueries::AssetNFT(assetnft::Query::Frozen { id, class_id }).into();
    let res = deps.querier.query(&request)?;
    Ok(res)
}

fn query_whitelisted(
    deps: Deps<CoreumQueries>,
    id: String,
    account: String,
) -> StdResult<WhitelistedResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let request: QueryRequest<CoreumQueries> =
        CoreumQueries::AssetNFT(assetnft::Query::Whitelisted {
            id,
            class_id,
            account,
        })
        .into();
    let res = deps.querier.query(&request)?;
    Ok(res)
}

fn query_whitelisted_accounts_for_nft(
    deps: Deps<CoreumQueries>,
    id: String,
) -> StdResult<WhitelistedAccountsForNFTResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let mut pagination = None;
    let mut accounts = vec![];
    let mut res: WhitelistedAccountsForNFTResponse;
    loop {
        let request = CoreumQueries::AssetNFT(assetnft::Query::WhitelistedAccountsForNFT {
            pagination,
            id: id.clone(),
            class_id: class_id.clone(),
        })
        .into();
        res = deps.querier.query(&request)?;
        accounts.append(&mut res.accounts);
        if res.pagination.next_key.is_none() {
            break;
        } else {
            pagination = Some(PageRequest {
                key: res.pagination.next_key,
                offset: None,
                limit: None,
                count_total: None,
                reverse: None,
            })
        }
    }
    let res = WhitelistedAccountsForNFTResponse {
        pagination: res.pagination,
        accounts,
    };
    Ok(res)
}

// ********** NFT **********

fn query_balance(deps: Deps<CoreumQueries>, owner: String) -> StdResult<nft::BalanceResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let request: QueryRequest<CoreumQueries> =
        CoreumQueries::NFT(nft::Query::Balance { class_id, owner }).into();
    let res = deps.querier.query(&request)?;
    Ok(res)
}

fn query_owner(deps: Deps<CoreumQueries>, id: String) -> StdResult<nft::OwnerResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let request: QueryRequest<CoreumQueries> =
        CoreumQueries::NFT(nft::Query::Owner { class_id, id }).into();
    let res = deps.querier.query(&request)?;
    Ok(res)
}

fn query_supply(deps: Deps<CoreumQueries>) -> StdResult<nft::SupplyResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let request: QueryRequest<CoreumQueries> =
        CoreumQueries::NFT(nft::Query::Supply { class_id }).into();
    let res = deps.querier.query(&request)?;
    Ok(res)
}

fn query_nft(deps: Deps<CoreumQueries>, id: String) -> StdResult<nft::NFTResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let request: QueryRequest<CoreumQueries> =
        CoreumQueries::NFT(nft::Query::NFT { class_id, id }).into();
    let res = deps.querier.query(&request)?;
    Ok(res)
}

fn query_nfts(deps: Deps<CoreumQueries>, owner: Option<String>) -> StdResult<nft::NFTsResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let mut pagination = None;
    let mut nfts = vec![];
    let mut res: nft::NFTsResponse;
    if owner.is_none() {
        loop {
            let request = CoreumQueries::NFT(nft::Query::NFTs {
                class_id: Some(class_id.clone()),
                owner: None,
                pagination,
            })
            .into();
            res = deps.querier.query(&request)?;
            nfts.append(&mut res.nfts);
            if res.pagination.next_key.is_none() {
                break;
            } else {
                pagination = Some(PageRequest {
                    key: res.pagination.next_key,
                    offset: None,
                    limit: None,
                    count_total: None,
                    reverse: None,
                })
            }
        }
        let res = nft::NFTsResponse {
            nfts,
            pagination: res.pagination,
        };
        Ok(res)
    } else {
        loop {
            let request = CoreumQueries::NFT(nft::Query::NFTs {
                class_id: None,
                owner: Some(owner.clone().unwrap()),
                pagination,
            })
            .into();
            res = deps.querier.query(&request)?;
            nfts.append(&mut res.nfts);
            if res.pagination.next_key.is_none() {
                break;
            } else {
                pagination = Some(PageRequest {
                    key: res.pagination.next_key,
                    offset: None,
                    limit: None,
                    count_total: None,
                    reverse: None,
                })
            }
        }
        let res = nft::NFTsResponse {
            nfts,
            pagination: res.pagination,
        };
        Ok(res)
    }
}

fn query_nft_class(deps: Deps<CoreumQueries>) -> StdResult<nft::ClassResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let request: QueryRequest<CoreumQueries> =
        CoreumQueries::NFT(nft::Query::Class { class_id }).into();
    let res = deps.querier.query(&request)?;
    Ok(res)
}

fn query_nft_classes(deps: Deps<CoreumQueries>) -> StdResult<nft::ClassesResponse> {
    let mut pagination = None;
    let mut classes = vec![];
    let mut res: nft::ClassesResponse;
    loop {
        let request = CoreumQueries::NFT(nft::Query::Classes { pagination }).into();
        res = deps.querier.query(&request)?;
        classes.append(&mut res.classes);
        if res.pagination.next_key.is_none() {
            break;
        } else {
            pagination = Some(PageRequest {
                key: res.pagination.next_key,
                offset: None,
                limit: None,
                count_total: None,
                reverse: None,
            })
        }
    }
    let res = nft::ClassesResponse {
        classes,
        pagination: res.pagination,
    };
    Ok(res)
}
