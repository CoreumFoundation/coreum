#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{Addr, Binary, CosmosMsg, DepsMut, Env, MessageInfo, Response};
use cw2::set_contract_version;
use protobuf::Message;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg};
use crate::state::GRANTER;
// Get Protos
include!("protos/mod.rs");
use CosmosAuthz::MsgExec;
use CosmosBankSend::Coin;
use CosmosBankSend::MsgSend;

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

    GRANTER.save(deps.storage, &deps.api.addr_validate(msg.granter.as_ref())?)?;

    Ok(Response::new()
        .add_attribute("contract", CONTRACT_NAME)
        .add_attribute("action", "instantiate")
        .add_attribute("granter", info.sender))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    _info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::Transfer{ address, amount, denom } => execute_transfer(deps, env, address, amount, denom),
    }
}

pub fn execute_transfer(
    deps: DepsMut,
    env: Env,
    address: Addr,
    amount: u64,
    denom: String,
) -> Result<Response, ContractError> {
    deps.api.addr_validate(address.as_ref())?;
    let granter = GRANTER.load(deps.storage)?;

    let mut send = MsgSend::new();
    send.from_address = granter.into_string();
    send.to_address = address.to_string();
    send.amount = vec![];
    let mut coin = Coin::new();
    coin.amount = amount.to_string();
    coin.denom = denom;
    send.amount.push(coin);

    let mut exec = MsgExec::new();
    exec.grantee = env.contract.address.to_string();
    exec.msgs = vec![send.to_any().unwrap()];
    let exec_bytes: Vec<u8> = exec.write_to_bytes().unwrap();

    let msg = CosmosMsg::Stargate {
        type_url: "/cosmos.authz.v1beta1.MsgExec".to_string(),
        value: Binary::from(exec_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "execute_authz_transfer")
        .add_message(msg))
}
