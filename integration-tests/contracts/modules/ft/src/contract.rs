use coreum_wasm_sdk::types::coreum::asset::ft::v1::{
    ExtensionIssueSettings, MsgBurn, MsgClawback, MsgClearAdmin, MsgFreeze, MsgGloballyFreeze,
    MsgGloballyUnfreeze, MsgIssue, MsgMint, MsgSetFrozen, MsgSetWhitelistedLimit, MsgTransferAdmin,
    MsgUnfreeze, MsgUpgradeTokenV1, QueryBalanceRequest, QueryBalanceResponse,
    QueryFrozenBalanceRequest, QueryFrozenBalanceResponse, QueryFrozenBalancesRequest,
    QueryFrozenBalancesResponse, QueryParamsRequest, QueryParamsResponse, QueryTokenRequest,
    QueryTokenResponse, QueryTokensRequest, QueryTokensResponse, QueryWhitelistedBalanceRequest,
    QueryWhitelistedBalanceResponse, QueryWhitelistedBalancesRequest,
    QueryWhitelistedBalancesResponse,
};
use coreum_wasm_sdk::types::cosmos::base::{query::v1beta1::PageRequest, v1beta1::Coin};
use cosmwasm_std::{
    entry_point, to_json_binary, to_json_vec, Binary, CosmosMsg, Deps, StdResult, Uint128,
};
use cosmwasm_std::{DepsMut, Env, MessageInfo, Response};
use cw2::set_contract_version;
use cw_ownable::{assert_owner, initialize_owner};

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::DENOM;

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

    let issue = MsgIssue {
        issuer: env.contract.address.to_string(),
        symbol: msg.symbol.clone(),
        subunit: msg.subunit.clone(),
        precision: msg.precision,
        initial_amount: msg.initial_amount.to_string(),
        description: msg.description.unwrap_or_default(),
        features: msg.features.unwrap_or_default(),
        burn_rate: msg.burn_rate,
        send_commission_rate: msg.send_commission_rate,
        uri: msg.uri.unwrap_or_default(),
        uri_hash: msg.uri_hash.unwrap_or_default(),
        extension_settings: msg.extension_settings.map(|s| ExtensionIssueSettings {
            code_id: s.code_id,
            label: s.label,
            funds: s
                .funds
                .iter()
                .map(|f| Coin {
                    denom: f.denom.to_string(),
                    amount: f.amount.to_string(),
                })
                .collect(),
            issuance_msg: to_json_vec(&s.issuance_msg).unwrap(),
        }),
        dex_settings: msg.dex_settings,
    };

    let denom = format!("{}-{}", msg.subunit, env.contract.address).to_lowercase();

    DENOM.save(deps.storage, &denom)?;

    Ok(Response::new()
        .add_attribute("owner", info.sender)
        .add_attribute("denom", denom)
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
        ExecuteMsg::Mint { amount, recipient } => mint(deps, env, info, amount, recipient),
        ExecuteMsg::Burn { amount } => burn(deps, env, info, amount),
        ExecuteMsg::Freeze { account, amount } => freeze(deps, env, info, account, amount),
        ExecuteMsg::Unfreeze { account, amount } => unfreeze(deps, env, info, account, amount),
        ExecuteMsg::SetFrozen { account, amount } => set_frozen(deps, env, info, account, amount),
        ExecuteMsg::GloballyFreeze {} => globally_freeze(deps, env, info),
        ExecuteMsg::GloballyUnfreeze {} => globally_unfreeze(deps, env, info),
        ExecuteMsg::Clawback { account, amount } => clawback(deps, env, info, account, amount),
        ExecuteMsg::SetWhitelistedLimit { account, amount } => {
            set_whitelisted_limit(deps, env, info, account, amount)
        }
        ExecuteMsg::TransferAdmin { account } => transfer_admin(deps, env, info, account),
        ExecuteMsg::ClearAdmin {} => clear_admin(deps, env, info),
        ExecuteMsg::UpgradeTokenV1 { ibc_enabled } => {
            upgrate_token_v1(deps, env, info, ibc_enabled)
        }
    }
}

// ********** Transactions **********

fn mint(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    amount: Uint128,
    recipient: Option<String>,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let denom = DENOM.load(deps.storage)?;

    let mint = MsgMint {
        sender: env.contract.address.to_string(),
        coin: Some(Coin {
            denom: denom.clone(),
            amount: amount.to_string(),
        }),
        recipient: recipient.unwrap_or_default(),
    };

    Ok(Response::new()
        .add_attribute("method", "mint")
        .add_attribute("denom", denom)
        .add_attribute("amount", amount.to_string())
        .add_message(CosmosMsg::Any(mint.to_any())))
}

fn burn(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    amount: Uint128,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let denom = DENOM.load(deps.storage)?;

    let burn = MsgBurn {
        sender: env.contract.address.to_string(),
        coin: Some(Coin {
            denom: denom.clone(),
            amount: amount.to_string(),
        }),
    };

    Ok(Response::new()
        .add_attribute("method", "burn")
        .add_attribute("denom", denom)
        .add_attribute("amount", amount.to_string())
        .add_message(CosmosMsg::Any(burn.to_any())))
}

fn freeze(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    account: String,
    amount: Uint128,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let denom = DENOM.load(deps.storage)?;

    let freeze = MsgFreeze {
        sender: env.contract.address.to_string(),
        account,
        coin: Some(Coin {
            denom: denom.clone(),
            amount: amount.to_string(),
        }),
    };

    Ok(Response::new()
        .add_attribute("method", "freeze")
        .add_attribute("denom", denom)
        .add_attribute("amount", amount.to_string())
        .add_message(CosmosMsg::Any(freeze.to_any())))
}

fn unfreeze(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    account: String,
    amount: Uint128,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let denom = DENOM.load(deps.storage)?;

    let unfreeze = MsgUnfreeze {
        sender: env.contract.address.to_string(),
        account,
        coin: Some(Coin {
            denom: denom.clone(),
            amount: amount.to_string(),
        }),
    };

    Ok(Response::new()
        .add_attribute("method", "unfreeze")
        .add_attribute("denom", denom)
        .add_attribute("amount", amount.to_string())
        .add_message(CosmosMsg::Any(unfreeze.to_any())))
}

fn set_frozen(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    account: String,
    amount: Uint128,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let denom = DENOM.load(deps.storage)?;

    let set_frozen = MsgSetFrozen {
        sender: env.contract.address.to_string(),
        account,
        coin: Some(Coin {
            denom: denom.clone(),
            amount: amount.to_string(),
        }),
    };

    Ok(Response::new()
        .add_attribute("method", "set_frozen")
        .add_attribute("denom", denom)
        .add_attribute("amount", amount.to_string())
        .add_message(CosmosMsg::Any(set_frozen.to_any())))
}

fn globally_freeze(deps: DepsMut, env: Env, info: MessageInfo) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let denom = DENOM.load(deps.storage)?;

    let globally_freeze = MsgGloballyFreeze {
        sender: env.contract.address.to_string(),
        denom: denom.clone(),
    };

    Ok(Response::new()
        .add_attribute("method", "globally_freeze")
        .add_attribute("denom", denom)
        .add_message(CosmosMsg::Any(globally_freeze.to_any())))
}

fn globally_unfreeze(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let denom = DENOM.load(deps.storage)?;

    let globally_unfreeze = MsgGloballyUnfreeze {
        sender: env.contract.address.to_string(),
        denom: denom.clone(),
    };

    Ok(Response::new()
        .add_attribute("method", "globally_unfreeze")
        .add_attribute("denom", denom)
        .add_message(CosmosMsg::Any(globally_unfreeze.to_any())))
}

fn clawback(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    account: String,
    amount: Uint128,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let denom = DENOM.load(deps.storage)?;

    let clawback = MsgClawback {
        sender: env.contract.address.to_string(),
        account,
        coin: Some(Coin {
            denom: denom.clone(),
            amount: amount.to_string(),
        }),
    };

    Ok(Response::new()
        .add_attribute("method", "clawback")
        .add_attribute("denom", denom)
        .add_message(CosmosMsg::Any(clawback.to_any())))
}

fn set_whitelisted_limit(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    account: String,
    amount: Uint128,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let denom = DENOM.load(deps.storage)?;

    let set_whitelisted_limit = MsgSetWhitelistedLimit {
        sender: env.contract.address.to_string(),
        account,
        coin: Some(Coin {
            denom: denom.clone(),
            amount: amount.to_string(),
        }),
    };

    Ok(Response::new()
        .add_attribute("method", "set_whitelisted_limit")
        .add_attribute("denom", denom)
        .add_attribute("amount", amount.to_string())
        .add_message(CosmosMsg::Any(set_whitelisted_limit.to_any())))
}

fn transfer_admin(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    account: String,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let denom = DENOM.load(deps.storage)?;

    let transfer_admin = MsgTransferAdmin {
        sender: env.contract.address.to_string(),
        account,
        denom: denom.clone(),
    };

    Ok(Response::new()
        .add_attribute("method", "transfer_admin")
        .add_attribute("denom", denom)
        .add_message(CosmosMsg::Any(transfer_admin.to_any())))
}

fn clear_admin(deps: DepsMut, env: Env, info: MessageInfo) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let denom = DENOM.load(deps.storage)?;

    let clear_admin = MsgClearAdmin {
        sender: env.contract.address.to_string(),
        denom: denom.clone(),
    };

    Ok(Response::new()
        .add_attribute("method", "clear_admin")
        .add_attribute("denom", denom)
        .add_message(CosmosMsg::Any(clear_admin.to_any())))
}

fn upgrate_token_v1(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    ibc_enabled: bool,
) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let denom = DENOM.load(deps.storage)?;

    let upgrade_token_v1 = MsgUpgradeTokenV1 {
        sender: env.contract.address.to_string(),
        denom: denom.clone(),
        ibc_enabled,
    };

    Ok(Response::new()
        .add_attribute("method", "upgrade_token_v1")
        .add_attribute("denom", denom)
        .add_attribute("ibc_enabled", ibc_enabled.to_string())
        .add_message(CosmosMsg::Any(upgrade_token_v1.to_any())))
}

// ********** Queries **********
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Params {} => to_json_binary(&query_params(deps)?),
        QueryMsg::Token {} => to_json_binary(&query_token(deps)?),
        QueryMsg::Tokens { issuer } => to_json_binary(&query_tokens(deps, issuer)?),
        QueryMsg::FrozenBalance { account } => {
            to_json_binary(&query_frozen_balance(deps, account)?)
        }
        QueryMsg::WhitelistedBalance { account } => {
            to_json_binary(&query_whitelisted_balance(deps, account)?)
        }
        QueryMsg::Balance { account } => to_json_binary(&query_balance(deps, account)?),
        QueryMsg::FrozenBalances { account } => {
            to_json_binary(&query_frozen_balances(deps, account)?)
        }
        QueryMsg::WhitelistedBalances { account } => {
            to_json_binary(&query_whitelisted_balances(deps, account)?)
        }
    }
}

fn query_params(deps: Deps) -> StdResult<QueryParamsResponse> {
    let request = QueryParamsRequest {};
    request.query(&deps.querier)
}

fn query_token(deps: Deps) -> StdResult<QueryTokenResponse> {
    let denom = DENOM.load(deps.storage)?;
    let request = QueryTokenRequest { denom };
    request.query(&deps.querier)
}

fn query_tokens(deps: Deps, issuer: String) -> StdResult<QueryTokensResponse> {
    let mut pagination = None;
    let mut tokens = vec![];
    let mut res: QueryTokensResponse;
    loop {
        let request = QueryTokensRequest {
            pagination,
            issuer: issuer.clone(),
        };
        res = request.query(&deps.querier)?;
        tokens.append(&mut res.tokens);
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
    let res = QueryTokensResponse {
        pagination: res.pagination,
        tokens,
    };
    Ok(res)
}

fn query_balance(deps: Deps, account: String) -> StdResult<QueryBalanceResponse> {
    let denom = DENOM.load(deps.storage)?;
    let request = QueryBalanceRequest { account, denom };
    request.query(&deps.querier)
}

fn query_frozen_balance(deps: Deps, account: String) -> StdResult<QueryFrozenBalanceResponse> {
    let denom = DENOM.load(deps.storage)?;
    let request = QueryFrozenBalanceRequest { denom, account };
    request.query(&deps.querier)
}

fn query_frozen_balances(deps: Deps, account: String) -> StdResult<QueryFrozenBalancesResponse> {
    let mut pagination = None;
    let mut balances = vec![];
    let mut res: QueryFrozenBalancesResponse;
    loop {
        let request = QueryFrozenBalancesRequest {
            pagination,
            account: account.clone(),
        };
        res = request.query(&deps.querier)?;
        balances.append(&mut res.balances);
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
    let res = QueryFrozenBalancesResponse {
        pagination: res.pagination,
        balances,
    };
    Ok(res)
}

fn query_whitelisted_balance(
    deps: Deps,
    account: String,
) -> StdResult<QueryWhitelistedBalanceResponse> {
    let denom = DENOM.load(deps.storage)?;
    let request = QueryWhitelistedBalanceRequest { denom, account };
    request.query(&deps.querier)
}

fn query_whitelisted_balances(
    deps: Deps,
    account: String,
) -> StdResult<QueryWhitelistedBalancesResponse> {
    let mut pagination = None;
    let mut balances = vec![];
    let mut res: QueryWhitelistedBalancesResponse;
    loop {
        let request = QueryWhitelistedBalancesRequest {
            pagination,
            account: account.clone(),
        };
        res = request.query(&deps.querier)?;
        balances.append(&mut res.balances);
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
    let res = QueryWhitelistedBalancesResponse {
        pagination: res.pagination,
        balances,
    };
    Ok(res)
}
