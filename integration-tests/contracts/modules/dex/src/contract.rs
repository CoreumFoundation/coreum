use coreum_wasm_sdk::types::coreum::asset::ft::v1::MsgIssue;
use coreum_wasm_sdk::types::coreum::dex::v1::{
    MsgCancelOrder, MsgCancelOrdersByDenom, MsgPlaceOrder, Order, OrderType,
    QueryAccountDenomOrdersCountRequest, QueryAccountDenomOrdersCountResponse,
    QueryOrderBookOrdersRequest, QueryOrderBookOrdersResponse, QueryOrderBooksRequest,
    QueryOrderBooksResponse, QueryOrderRequest, QueryOrderResponse, QueryOrdersRequest,
    QueryOrdersResponse, QueryParamsRequest, QueryParamsResponse, Side, TimeInForce,
};
use coreum_wasm_sdk::types::cosmos::base::query::v1beta1::PageRequest;
use cosmwasm_std::{entry_point, to_json_binary, Binary, CosmosMsg, Deps, StdError, StdResult};
use cosmwasm_std::{DepsMut, Env, MessageInfo, Response};
use cw2::set_contract_version;
use cw_ownable::initialize_owner;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};

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
        extension_settings: msg.extension_settings,
        dex_settings: msg.dex_settings,
    };

    let denom = format!("{}-{}", msg.subunit, env.contract.address).to_lowercase();

    Ok(Response::new()
        .add_attribute("owner", info.sender)
        .add_attribute("denom", denom)
        .add_message(CosmosMsg::Any(issue.to_any())))
}

// ********** Execute **********

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    _deps: DepsMut,
    env: Env,
    _info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::PlaceOrder { order } => place_order(env, order),
        ExecuteMsg::CancelOrder { order_id } => cancel_order(env, order_id),
        ExecuteMsg::CancelOrdersByDenom { account, denom } => {
            cancel_orders_by_denom(env, account, denom)
        }
    }
}

// ********** Transactions **********

fn place_order(env: Env, order: MsgPlaceOrder) -> Result<Response, ContractError> {
    let place_order = MsgPlaceOrder {
        sender: env.contract.address.to_string(),
        r#type: order.r#type,
        id: order.id.clone(),
        base_denom: order.base_denom.clone(),
        quote_denom: order.quote_denom.clone(),
        price: order.price.clone(),
        quantity: order.quantity.clone(),
        side: order.side,
        good_til: order.good_til.clone(),
        time_in_force: order.time_in_force,
    };

    let mut res = Response::new()
        .add_attribute("method", "place_order")
        .add_attribute("sender", env.contract.address.to_string())
        .add_attribute(
            "type",
            OrderType::try_from(order.r#type)
                .or(Err(ContractError::Std(StdError::generic_err(
                    "wrong order type",
                ))))?
                .as_str_name(),
        )
        .add_attribute("id", order.id)
        .add_attribute("base_denom", order.base_denom)
        .add_attribute("quote_denom", order.quote_denom)
        .add_attribute("price", order.price)
        .add_attribute("quantity", order.quantity)
        .add_attribute(
            "side",
            Side::try_from(order.side)
                .or(Err(ContractError::Std(StdError::generic_err(
                    "wrong order side",
                ))))?
                .as_str_name(),
        )
        .add_attribute(
            "time_in_force",
            TimeInForce::try_from(order.time_in_force)
                .or(Err(ContractError::Std(StdError::generic_err(
                    "wrong order side",
                ))))?
                .as_str_name(),
        );
    if let Some(good_til) = order.good_til {
        res = res.add_attribute(
            "good_til_block_height",
            good_til.good_til_block_height.to_string(),
        );
        if let Some(good_til_block_time) = good_til.good_til_block_time {
            res = res.add_attribute(
                "good_til_block_time_seconds",
                good_til_block_time.seconds.to_string(),
            );
        }
    }
    res = res.add_message(CosmosMsg::Any(place_order.to_any()));
    Ok(res)
}

fn cancel_order(env: Env, order_id: String) -> Result<Response, ContractError> {
    let cancel_order = MsgCancelOrder {
        sender: env.contract.address.to_string(),
        id: order_id.clone(),
    };

    Ok(Response::new()
        .add_attribute("method", "cancel_order")
        .add_attribute("sender", env.contract.address.to_string())
        .add_attribute("order_id", order_id)
        .add_message(CosmosMsg::Any(cancel_order.to_any())))
}

fn cancel_orders_by_denom(
    env: Env,
    account: String,
    denom: String,
) -> Result<Response, ContractError> {
    let cancel_orders_by_denom = MsgCancelOrdersByDenom {
        sender: env.contract.address.to_string(),
        denom: denom.clone(),
        account: account.clone(),
    };

    Ok(Response::new()
        .add_attribute("method", "cancel_orders_by_denom")
        .add_attribute("sender", env.contract.address.to_string())
        .add_attribute("account", account)
        .add_attribute("denom", denom)
        .add_message(CosmosMsg::Any(cancel_orders_by_denom.to_any())))
}

// ********** Queries **********
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Params {} => to_json_binary(&query_params(deps)?),
        QueryMsg::Order { acc, order_id } => to_json_binary(&query_order(deps, acc, order_id)?),
        QueryMsg::Orders { creator } => to_json_binary(&query_orders(deps, creator)?),
        QueryMsg::OrderBooks {} => to_json_binary(&query_order_books(deps)?),
        QueryMsg::OrderBookOrders {
            base_denom,
            quote_denom,
            side,
        } => to_json_binary(&query_order_book_orders(
            deps,
            base_denom,
            quote_denom,
            side,
        )?),
        QueryMsg::AccountDenomOrdersCount { account, denom } => {
            to_json_binary(&query_account_denom_orders_count(deps, account, denom)?)
        }
    }
}

fn query_params(deps: Deps) -> StdResult<QueryParamsResponse> {
    let request = QueryParamsRequest {};
    request.query(&deps.querier)
}

fn query_order(deps: Deps, acc: String, order_id: String) -> StdResult<QueryOrderResponse> {
    let request = QueryOrderRequest {
        creator: acc,
        id: order_id,
    };
    request.query(&deps.querier)
}

fn query_orders(deps: Deps, creator: String) -> StdResult<QueryOrdersResponse> {
    let mut pagination = None;
    let mut orders = vec![];
    let mut res: QueryOrdersResponse;
    loop {
        let request = QueryOrdersRequest {
            pagination,
            creator: creator.clone(),
        };
        res = request.query(&deps.querier)?;
        orders.append(&mut res.orders);
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
    let res = QueryOrdersResponse {
        pagination: res.pagination,
        orders,
    };
    Ok(res)
}

fn query_order_books(deps: Deps) -> StdResult<QueryOrderBooksResponse> {
    let mut pagination = None;
    let mut order_books = vec![];
    let mut res: QueryOrderBooksResponse;
    loop {
        let request = QueryOrderBooksRequest { pagination };
        res = request.query(&deps.querier)?;
        order_books.append(&mut res.order_books);
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
    let res = QueryOrderBooksResponse {
        pagination: res.pagination,
        order_books,
    };
    Ok(res)
}

fn query_order_book_orders(
    deps: Deps,
    base_denom: String,
    quote_denom: String,
    side: i32,
) -> StdResult<QueryOrderBookOrdersResponse> {
    let mut pagination = None;
    let mut orders = vec![];
    let mut res: QueryOrderBookOrdersResponse;
    loop {
        let request = QueryOrderBookOrdersRequest {
            pagination,
            base_denom: base_denom.clone(),
            quote_denom: quote_denom.clone(),
            side,
        };
        res = request.query(&deps.querier)?;
        orders.append(&mut res.orders);
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
    let res = QueryOrderBookOrdersResponse {
        pagination: res.pagination,
        orders,
    };
    Ok(res)
}

fn query_account_denom_orders_count(
    deps: Deps,
    account: String,
    denom: String,
) -> StdResult<QueryAccountDenomOrdersCountResponse> {
    let request = QueryAccountDenomOrdersCountRequest { account, denom };
    request.query(&deps.querier)
}
