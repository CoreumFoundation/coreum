use coreum_wasm_sdk::shim;
use coreum_wasm_sdk::types::coreum::asset::nft::v1::{
    self, DataBytes, DataDynamic, DataDynamicIndexedItem, DataDynamicItem, DataEditor,
    MsgAddToClassWhitelist, MsgAddToWhitelist, MsgBurn, MsgClassFreeze, MsgClassUnfreeze,
    MsgFreeze, MsgIssueClass, MsgMint, MsgRemoveFromClassWhitelist, MsgRemoveFromWhitelist,
    MsgUnfreeze, MsgUpdateData, QueryBurntNfTsInClassRequest, QueryBurntNfTsInClassResponse,
    QueryBurntNftRequest, QueryBurntNftResponse, QueryClassFrozenAccountsRequest,
    QueryClassFrozenAccountsResponse, QueryClassFrozenRequest, QueryClassFrozenResponse,
    QueryClassWhitelistedAccountsRequest, QueryClassWhitelistedAccountsResponse,
    QueryFrozenRequest, QueryFrozenResponse, QueryParamsRequest, QueryParamsResponse,
    QueryWhitelistedAccountsForNftRequest, QueryWhitelistedAccountsForNftResponse,
    QueryWhitelistedRequest, QueryWhitelistedResponse,
};
use coreum_wasm_sdk::types::coreum::nft::v1beta1::{
    self, QueryBalanceRequest, QueryBalanceResponse, QueryNfTsRequest, QueryNfTsResponse,
    QueryNftRequest, QueryNftResponse, QueryOwnerRequest, QueryOwnerResponse, QuerySupplyRequest,
    QuerySupplyResponse,
};
use coreum_wasm_sdk::types::cosmos::base::query::v1beta1::PageRequest;
use coreum_wasm_sdk::types::cosmos::nft::v1beta1::MsgSend;
use cosmwasm_std::{
    entry_point, to_json_binary, Binary, CosmosMsg, Deps, DepsMut, Env, MessageInfo, Response,
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
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
    initialize_owner(deps.storage, deps.api, Some(info.sender.as_ref()))?;

    let data = msg.data.map(|data| {
        DataBytes {
            data: data.to_vec(),
        }
        .to_any()
    });

    let issue = MsgIssueClass {
        issuer: env.contract.address.to_string(),
        name: msg.name,
        symbol: msg.symbol.clone(),
        description: msg.description.unwrap_or_default(),
        uri: msg.uri.unwrap_or_default(),
        uri_hash: msg.uri_hash.unwrap_or_default(),
        data: data.map(|d| shim::Any {
            type_url: d.type_url,
            value: d.value.to_vec(),
        }),
        features: msg.features.unwrap_or_default(),
        royalty_rate: msg.royalty_rate.unwrap_or_default(),
    };

    let class_id = format!("{}-{}", msg.symbol, env.contract.address).to_lowercase();

    CLASS_ID.save(deps.storage, &class_id)?;

    Ok(Response::new()
        .add_attribute("owner", info.sender)
        .add_attribute("class_id", class_id)
        .add_message(CosmosMsg::Any(issue.to_any())))
}

// ********** Execute **********

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::MintImmutable {
            id,
            uri,
            uri_hash,
            data,
            recipient,
        } => mint_immutable(deps, info, env, id, uri, uri_hash, data, recipient),
        ExecuteMsg::MintMutable {
            id,
            uri,
            uri_hash,
            data,
            recipient,
        } => mint_mutable(deps, info, env, id, uri, uri_hash, data, recipient),
        ExecuteMsg::Burn { id } => burn(deps, env, info, id),
        ExecuteMsg::Freeze { id } => freeze(deps, env, info, id),
        ExecuteMsg::Unfreeze { id } => unfreeze(deps, env, info, id),
        ExecuteMsg::AddToWhitelist { id, account } => {
            add_to_white_list(deps, env, info, id, account)
        }
        ExecuteMsg::RemoveFromWhitelist { id, account } => {
            remove_from_white_list(deps, env, info, id, account)
        }
        ExecuteMsg::Send { id, receiver } => send(deps, env, info, id, receiver),
        ExecuteMsg::ClassFreeze { account } => class_freeze(deps, env, info, account),
        ExecuteMsg::ClassUnfreeze { account } => class_unfreeze(deps, env, info, account),
        ExecuteMsg::AddToClassWhitelist { account } => {
            add_to_class_whitelist(deps, env, info, account)
        }
        ExecuteMsg::RemoveFromClassWhitelist { account } => {
            remove_from_class_whitelist(deps, env, info, account)
        }
        ExecuteMsg::ModifyData { id, data } => modify_data(deps, info, env, id, data),
    }
}

// ********** Transactions **********

#[allow(clippy::too_many_arguments)]
fn mint_immutable(
    deps: DepsMut,
    info: MessageInfo,
    env: Env,
    id: String,
    uri: Option<String>,
    uri_hash: Option<String>,
    data: Option<Binary>,
    recipient: Option<String>,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let data = data.map(|data| {
        DataBytes {
            data: data.to_vec(),
        }
        .to_any()
    });

    let mint = MsgMint {
        sender: env.contract.address.to_string(),
        class_id: class_id.clone(),
        id: id.clone(),
        uri: uri.unwrap_or_default(),
        uri_hash: uri_hash.unwrap_or_default(),
        data: data.map(|d| shim::Any {
            type_url: d.type_url,
            value: d.value.to_vec(),
        }),
        recipient: recipient.unwrap_or_default(),
    };

    Ok(Response::new()
        .add_attribute("method", "mint_immutable")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(CosmosMsg::Any(mint.to_any())))
}

#[allow(clippy::too_many_arguments)]
fn mint_mutable(
    deps: DepsMut,
    info: MessageInfo,
    env: Env,
    id: String,
    uri: Option<String>,
    uri_hash: Option<String>,
    data: Option<Binary>,
    recipient: Option<String>,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let data = data.map(|data| {
        DataDynamic {
            items: [DataDynamicItem {
                editors: [DataEditor::Admin as i32, DataEditor::Owner as i32].to_vec(),
                data: data.to_vec(),
            }]
            .to_vec(),
        }
        .to_any()
    });

    let mint = MsgMint {
        sender: env.contract.address.to_string(),
        class_id: class_id.clone(),
        id: id.clone(),
        uri: uri.unwrap_or_default(),
        uri_hash: uri_hash.unwrap_or_default(),
        data: data.map(|d| shim::Any {
            type_url: d.type_url,
            value: d.value.to_vec(),
        }),
        recipient: recipient.unwrap_or_default(),
    };

    Ok(Response::new()
        .add_attribute("method", "mint_mutable")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(CosmosMsg::Any(mint.to_any())))
}

fn modify_data(
    deps: DepsMut,
    info: MessageInfo,
    env: Env,
    id: String,
    data: Binary,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let modify_data = MsgUpdateData {
        sender: env.contract.address.to_string(),
        class_id: class_id.clone(),
        id: id.clone(),
        items: [DataDynamicIndexedItem {
            index: 0,
            data: data.to_vec(),
        }]
        .to_vec(),
    };

    Ok(Response::new()
        .add_attribute("method", "modify_data")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(CosmosMsg::Any(modify_data.to_any())))
}

fn burn(deps: DepsMut, env: Env, info: MessageInfo, id: String) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let burn = MsgBurn {
        sender: env.contract.address.to_string(),
        class_id: class_id.clone(),
        id: id.clone(),
    };

    Ok(Response::new()
        .add_attribute("method", "burn")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(CosmosMsg::Any(burn.to_any())))
}

fn freeze(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    id: String,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let freeze = MsgFreeze {
        sender: env.contract.address.to_string(),
        class_id: class_id.clone(),
        id: id.clone(),
    };

    Ok(Response::new()
        .add_attribute("method", "freeze")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(CosmosMsg::Any(freeze.to_any())))
}

fn unfreeze(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    id: String,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let unfreeze = MsgUnfreeze {
        sender: env.contract.address.to_string(),
        class_id: class_id.clone(),
        id: id.clone(),
    };

    Ok(Response::new()
        .add_attribute("method", "unfreeze")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(CosmosMsg::Any(unfreeze.to_any())))
}

fn add_to_white_list(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    id: String,
    account: String,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let add_to_whitelist = MsgAddToWhitelist {
        sender: env.contract.address.to_string(),
        class_id: class_id.clone(),
        id: id.clone(),
        account: account.clone(),
    };

    Ok(Response::new()
        .add_attribute("method", "add_to_white_list")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(CosmosMsg::Any(add_to_whitelist.to_any())))
}

fn remove_from_white_list(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    id: String,
    account: String,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let remove_from_whitelist = MsgRemoveFromWhitelist {
        sender: env.contract.address.to_string(),
        class_id: class_id.clone(),
        id: id.clone(),
        account: account.clone(),
    };

    Ok(Response::new()
        .add_attribute("method", "remove_from_white_list")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(CosmosMsg::Any(remove_from_whitelist.to_any())))
}

fn send(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    id: String,
    receiver: String,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let send = MsgSend {
        sender: env.contract.address.to_string(),
        class_id: class_id.clone(),
        id: id.clone(),
        receiver: receiver.clone(),
    };

    Ok(Response::new()
        .add_attribute("method", "send")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(CosmosMsg::Any(send.to_any())))
}

fn class_freeze(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    account: String,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let class_freeze = MsgClassFreeze {
        sender: env.contract.address.to_string(),
        class_id: class_id.clone(),
        account: account.clone(),
    };

    Ok(Response::new()
        .add_attribute("method", "class_freeze")
        .add_attribute("class_id", class_id)
        .add_attribute("account", account)
        .add_message(CosmosMsg::Any(class_freeze.to_any())))
}

fn class_unfreeze(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    account: String,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let class_unfreeze = MsgClassUnfreeze {
        sender: env.contract.address.to_string(),
        class_id: class_id.clone(),
        account: account.clone(),
    };

    Ok(Response::new()
        .add_attribute("method", "class_unfreeze")
        .add_attribute("class_id", class_id)
        .add_attribute("account", account)
        .add_message(CosmosMsg::Any(class_unfreeze.to_any())))
}

fn add_to_class_whitelist(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    account: String,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let add_to_class_whitelist = MsgAddToClassWhitelist {
        sender: env.contract.address.to_string(),
        class_id: class_id.clone(),
        account: account.clone(),
    };

    Ok(Response::new()
        .add_attribute("method", "add_to_class_whitelist")
        .add_attribute("class_id", class_id)
        .add_attribute("account", account)
        .add_message(CosmosMsg::Any(add_to_class_whitelist.to_any())))
}

fn remove_from_class_whitelist(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    account: String,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let remove_from_class_whitelist = MsgRemoveFromClassWhitelist {
        sender: env.contract.address.to_string(),
        class_id: class_id.clone(),
        account: account.clone(),
    };

    Ok(Response::new()
        .add_attribute("method", "remove_from_class_whitelist")
        .add_attribute("class_id", class_id)
        .add_attribute("account", account)
        .add_message(CosmosMsg::Any(remove_from_class_whitelist.to_any())))
}

// ********** Queries **********

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Params {} => to_json_binary(&query_params(deps)?),
        QueryMsg::Class {} => to_json_binary(&query_class(deps)?),
        QueryMsg::Classes { issuer } => to_json_binary(&query_classes(deps, issuer)?),
        QueryMsg::Frozen { id } => to_json_binary(&query_frozen(deps, id)?),
        QueryMsg::Whitelisted { id, account } => {
            to_json_binary(&query_whitelisted(deps, id, account)?)
        }
        QueryMsg::WhitelistedAccountsForNft { id } => {
            to_json_binary(&query_whitelisted_accounts_for_nft(deps, id)?)
        }
        QueryMsg::Balance { owner } => to_json_binary(&query_balance(deps, owner)?),
        QueryMsg::Owner { id } => to_json_binary(&query_owner(deps, id)?),
        QueryMsg::Supply {} => to_json_binary(&query_supply(deps)?),
        QueryMsg::Nft { id } => to_json_binary(&query_nft(deps, id)?),
        QueryMsg::Nfts { owner } => to_json_binary(&query_nfts(deps, owner)?),
        QueryMsg::ClassNft {} => to_json_binary(&query_nft_class(deps)?),
        QueryMsg::ClassesNft {} => to_json_binary(&query_nft_classes(deps)?),
        QueryMsg::BurntNft { nft_id } => to_json_binary(&query_burnt_nft(deps, nft_id)?),
        QueryMsg::BurntNftsInClass {} => to_json_binary(&query_burnt_nfts_in_class(deps)?),
        QueryMsg::ClassFrozen { account } => to_json_binary(&query_class_frozen(deps, account)?),
        QueryMsg::ClassFrozenAccounts {} => to_json_binary(&query_class_frozen_accounts(deps)?),
        QueryMsg::ClassWhitelistedAccounts {} => {
            to_json_binary(&query_class_whitelisted_accounts(deps)?)
        }
        QueryMsg::ExternalNft { class_id, id } => {
            to_json_binary(&query_external_nft(deps, class_id, id)?)
        }
    }
}

fn query_params(deps: Deps) -> StdResult<QueryParamsResponse> {
    let req = QueryParamsRequest {};
    req.query(&deps.querier)
}

fn query_class(deps: Deps) -> StdResult<v1::QueryClassResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let req = v1::QueryClassRequest { id: class_id };
    req.query(&deps.querier)
}

fn query_classes(deps: Deps, issuer: String) -> StdResult<v1::QueryClassesResponse> {
    let mut pagination = None;
    let mut classes = vec![];
    let mut res: v1::QueryClassesResponse;
    loop {
        let req = v1::QueryClassesRequest {
            pagination,
            issuer: issuer.clone(),
        };
        res = req.query(&deps.querier)?;
        classes.append(&mut res.classes);
        let next_key = res.pagination.clone().and_then(|p| p.next_key);
        if next_key.is_none() {
            break;
        } else {
            pagination = Some(PageRequest {
                key: next_key.unwrap(),
                offset: 0,
                limit: 0,
                count_total: false,
                reverse: false,
            })
        }
    }
    let res = v1::QueryClassesResponse {
        pagination: res.pagination,
        classes,
    };
    Ok(res)
}

fn query_frozen(deps: Deps, id: String) -> StdResult<QueryFrozenResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let req = QueryFrozenRequest { id, class_id };
    req.query(&deps.querier)
}

fn query_whitelisted(
    deps: Deps,
    id: String,
    account: String,
) -> StdResult<QueryWhitelistedResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let req = QueryWhitelistedRequest {
        id,
        class_id,
        account,
    };
    req.query(&deps.querier)
}

fn query_whitelisted_accounts_for_nft(
    deps: Deps,
    id: String,
) -> StdResult<QueryWhitelistedAccountsForNftResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let mut pagination = None;
    let mut accounts = vec![];
    let mut res: QueryWhitelistedAccountsForNftResponse;
    loop {
        let req = QueryWhitelistedAccountsForNftRequest {
            pagination,
            id: id.clone(),
            class_id: class_id.clone(),
        };
        res = req.query(&deps.querier)?;
        accounts.append(&mut res.accounts);
        let next_key = res.pagination.clone().and_then(|p| p.next_key);
        if next_key.is_none() {
            break;
        } else {
            pagination = Some(PageRequest {
                key: next_key.unwrap(),
                offset: 0,
                limit: 0,
                count_total: false,
                reverse: false,
            })
        }
    }
    let res = QueryWhitelistedAccountsForNftResponse {
        pagination: res.pagination,
        accounts,
    };
    Ok(res)
}

fn query_burnt_nft(deps: Deps, nft_id: String) -> StdResult<QueryBurntNftResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let req = QueryBurntNftRequest { class_id, nft_id };
    req.query(&deps.querier)
}

fn query_burnt_nfts_in_class(deps: Deps) -> StdResult<QueryBurntNfTsInClassResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let mut pagination = None;
    let mut nft_ids = vec![];
    let mut res: QueryBurntNfTsInClassResponse;
    loop {
        let req = QueryBurntNfTsInClassRequest {
            pagination,
            class_id: class_id.clone(),
        };
        res = req.query(&deps.querier)?;
        nft_ids.append(&mut res.nft_ids);
        let next_key = res.pagination.clone().and_then(|p| p.next_key);
        if next_key.is_none() {
            break;
        } else {
            pagination = Some(PageRequest {
                key: next_key.unwrap(),
                offset: 0,
                limit: 0,
                count_total: false,
                reverse: false,
            })
        }
    }
    let res = QueryBurntNfTsInClassResponse {
        pagination: res.pagination,
        nft_ids,
    };
    Ok(res)
}

fn query_class_frozen(deps: Deps, account: String) -> StdResult<QueryClassFrozenResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let req = QueryClassFrozenRequest { class_id, account };
    req.query(&deps.querier)
}

fn query_class_frozen_accounts(deps: Deps) -> StdResult<QueryClassFrozenAccountsResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let mut pagination = None;
    let mut accounts = vec![];
    let mut res: QueryClassFrozenAccountsResponse;
    loop {
        let req = QueryClassFrozenAccountsRequest {
            pagination,
            class_id: class_id.clone(),
        };
        res = req.query(&deps.querier)?;
        accounts.append(&mut res.accounts);
        let next_key = res.pagination.clone().and_then(|p| p.next_key);
        if next_key.is_none() {
            break;
        } else {
            pagination = Some(PageRequest {
                key: next_key.unwrap(),
                offset: 0,
                limit: 0,
                count_total: false,
                reverse: false,
            })
        }
    }
    let res = QueryClassFrozenAccountsResponse {
        pagination: res.pagination,
        accounts,
    };
    Ok(res)
}

fn query_class_whitelisted_accounts(
    deps: Deps,
) -> StdResult<QueryClassWhitelistedAccountsResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let mut pagination = None;
    let mut accounts = vec![];
    let mut res: QueryClassWhitelistedAccountsResponse;
    loop {
        let req = QueryClassWhitelistedAccountsRequest {
            pagination,
            class_id: class_id.clone(),
        };
        res = req.query(&deps.querier)?;
        accounts.append(&mut res.accounts);
        let next_key = res.pagination.clone().and_then(|p| p.next_key);
        if next_key.is_none() {
            break;
        } else {
            pagination = Some(PageRequest {
                key: next_key.unwrap(),
                offset: 0,
                limit: 0,
                count_total: false,
                reverse: false,
            })
        }
    }
    let res = QueryClassWhitelistedAccountsResponse {
        pagination: res.pagination,
        accounts,
    };
    Ok(res)
}

// ********** NFT **********

fn query_balance(deps: Deps, owner: String) -> StdResult<QueryBalanceResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let req = QueryBalanceRequest { class_id, owner };
    req.query(&deps.querier)
}

fn query_owner(deps: Deps, id: String) -> StdResult<QueryOwnerResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let req = QueryOwnerRequest { class_id, id };
    req.query(&deps.querier)
}

fn query_supply(deps: Deps) -> StdResult<QuerySupplyResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let req = QuerySupplyRequest { class_id };
    req.query(&deps.querier)
}

fn query_nft(deps: Deps, id: String) -> StdResult<QueryNftResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let req = QueryNftRequest { class_id, id };
    req.query(&deps.querier)
}

fn query_nfts(deps: Deps, owner: Option<String>) -> StdResult<QueryNfTsResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let mut pagination = None;
    let mut nfts = vec![];
    let mut res: QueryNfTsResponse;
    if owner.is_none() {
        loop {
            let req = QueryNfTsRequest {
                class_id: class_id.clone(),
                owner: "".to_string(),
                pagination,
            };
            res = req.query(&deps.querier)?;
            nfts.append(&mut res.nfts);
            let next_key = res.pagination.clone().and_then(|p| p.next_key);
            if next_key.is_none() {
                break;
            } else {
                pagination = Some(PageRequest {
                    key: next_key.unwrap(),
                    offset: 0,
                    limit: 0,
                    count_total: false,
                    reverse: false,
                })
            }
        }
        let res = QueryNfTsResponse {
            nfts,
            pagination: res.pagination,
        };
        Ok(res)
    } else {
        loop {
            let req = QueryNfTsRequest {
                class_id: "".to_string(),
                owner: owner.clone().unwrap(),
                pagination,
            };
            res = req.query(&deps.querier)?;
            nfts.append(&mut res.nfts);
            let next_key = res.pagination.clone().and_then(|p| p.next_key);
            if next_key.is_none() {
                break;
            } else {
                pagination = Some(PageRequest {
                    key: next_key.unwrap(),
                    offset: 0,
                    limit: 0,
                    count_total: false,
                    reverse: false,
                })
            }
        }
        let res = QueryNfTsResponse {
            nfts,
            pagination: res.pagination,
        };
        Ok(res)
    }
}

fn query_nft_class(deps: Deps) -> StdResult<v1beta1::QueryClassResponse> {
    let class_id = CLASS_ID.load(deps.storage)?;
    let req = v1beta1::QueryClassRequest { class_id };
    req.query(&deps.querier)
}

fn query_nft_classes(deps: Deps) -> StdResult<v1beta1::QueryClassesResponse> {
    let mut pagination = None;
    let mut classes = vec![];
    let mut res: v1beta1::QueryClassesResponse;
    loop {
        let req = v1beta1::QueryClassesRequest { pagination };
        res = req.query(&deps.querier)?;
        classes.append(&mut res.classes);
        let next_key = res.pagination.clone().and_then(|p| p.next_key);
        if next_key.is_none() {
            break;
        } else {
            pagination = Some(PageRequest {
                key: next_key.unwrap(),
                offset: 0,
                limit: 0,
                count_total: false,
                reverse: false,
            })
        }
    }
    let res = v1beta1::QueryClassesResponse {
        classes,
        pagination: res.pagination,
    };
    Ok(res)
}

fn query_external_nft(deps: Deps, class_id: String, id: String) -> StdResult<QueryNftResponse> {
    let req = QueryNftRequest { class_id, id };
    req.query(&deps.querier)
}
