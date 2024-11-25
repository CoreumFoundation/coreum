use coreum_wasm_sdk::deprecated::core::CoreumResult;
use coreum_wasm_sdk::shim;
use coreum_wasm_sdk::types::cosmos::authz::v1beta1::MsgExec;
use coreum_wasm_sdk::types::cosmos::bank::v1beta1::MsgSend;
use coreum_wasm_sdk::types::cosmos::base::v1beta1::Coin;

#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{CosmosMsg, DepsMut, Env, MessageInfo, Response, Uint128};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg};
use crate::state::GRANTER;

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
) -> CoreumResult<ContractError> {
    match msg {
        ExecuteMsg::Transfer {
            address,
            amount,
            denom,
        } => execute_transfer(deps, env, address, amount, denom),
    }
}

fn execute_transfer(
    deps: DepsMut,
    env: Env,
    address: String,
    amount: Uint128,
    denom: String,
) -> CoreumResult<ContractError> {
    deps.api.addr_validate(address.as_ref())?;
    let granter = GRANTER.load(deps.storage)?;

    let send = MsgSend {
        from_address: granter.to_string(),
        to_address: address,
        amount: vec![Coin {
            denom,
            amount: amount.to_string(),
        }],
    }
    .to_any();

    let exec = MsgExec {
        grantee: env.contract.address.to_string(),
        msgs: vec![shim::Any {
            type_url: send.type_url,
            value: send.value.to_vec(),
        }],
    };

    Ok(Response::new()
        .add_attribute("method", "execute_authz_transfer")
        .add_message(CosmosMsg::Any(exec.to_any())))
}
