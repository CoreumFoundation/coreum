use coreum_wasm_sdk::core::{CoreumMsg, CoreumResult};
use coreum_wasm_sdk::nft;
use coreum_wasm_sdk::types::cosmos::authz::v1beta1::MsgExec;
use coreum_wasm_sdk::types::cosmos::bank::v1beta1::MsgSend;
use coreum_wasm_sdk::types::cosmos::base::v1beta1::Coin;
use coreum_wasm_sdk::types::cosmos::nft::v1beta1::MsgSend as MsgSendNft;

#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{
    BankMsg, Binary, Coin as CWCoin, CosmosMsg, DepsMut, Env, MessageInfo, Response, Uint128,
};
use cw2::set_contract_version;
use cw_utils::one_coin;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg};
use crate::state::{Offer, GRANTER, NFT_OFFERS};

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
    info: MessageInfo,
    msg: ExecuteMsg,
) -> CoreumResult<ContractError> {
    match msg {
        ExecuteMsg::Transfer {
            address,
            amount,
            denom,
        } => execute_transfer(deps, env, address, amount, denom),
        ExecuteMsg::OfferNft {
            class_id,
            id,
            price,
        } => offer_nft(deps, env, info, class_id, id, price),
        ExecuteMsg::AcceptNftOffer { class_id, id } => accept_offer(deps, info, class_id, id),
        ExecuteMsg::Stargate { type_url, value } => execute_stargate_message(type_url, value),
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
    };

    let exec = MsgExec {
        grantee: env.contract.address.to_string(),
        msgs: vec![send.to_any()],
    };
    let exec_bytes: Vec<u8> = exec.to_proto_bytes();

    let msg = CosmosMsg::Stargate {
        type_url: "/cosmos.authz.v1beta1.MsgExec".to_string(),
        value: Binary::from(exec_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "execute_authz_transfer")
        .add_message(msg))
}

// The contract must have been granted authorization to send the NFT before execution or it will fail.
// We will send the NFT to the contract to be able to sell it when someone provides the price.
fn offer_nft(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    class_id: String,
    id: String,
    price: CWCoin,
) -> CoreumResult<ContractError> {
    let nft_send = MsgSendNft {
        class_id: class_id.clone(),
        id: id.clone(),
        sender: info.sender.to_string(),
        receiver: env.contract.address.to_string(),
    };

    let exec = MsgExec {
        grantee: env.contract.address.to_string(),
        msgs: vec![nft_send.to_any()],
    };
    let exec_bytes: Vec<u8> = exec.to_proto_bytes();

    let msg = CosmosMsg::Stargate {
        type_url: "/cosmos.authz.v1beta1.MsgExec".to_string(),
        value: Binary::from(exec_bytes),
    };

    NFT_OFFERS.save(
        deps.storage,
        (class_id, id),
        &Offer {
            address: info.sender,
            price,
        },
    )?;

    Ok(Response::new()
        .add_attribute("method", "execute_offer_nft_authz")
        .add_message(msg))
}

fn accept_offer(
    deps: DepsMut,
    info: MessageInfo,
    class_id: String,
    id: String,
) -> CoreumResult<ContractError> {
    let offer = NFT_OFFERS.load(deps.storage, (class_id.clone(), id.clone()))?;

    if one_coin(&info)? != offer.price {
        return Err(ContractError::InvalidFundsAmount {});
    }

    let nft_send_msg = CosmosMsg::from(CoreumMsg::NFT(nft::Msg::Send {
        class_id,
        id,
        receiver: info.sender.to_string(),
    }));

    let send_funds_msg = CosmosMsg::Bank(BankMsg::Send {
        to_address: offer.address.to_string(),
        amount: info.funds,
    });

    Ok(Response::new()
        .add_attribute("method", "execute_accept_nft_offer")
        .add_messages([nft_send_msg, send_funds_msg]))
}

pub fn execute_stargate_message(type_url: String, value: Binary) -> CoreumResult<ContractError> {
    let msg = CosmosMsg::Stargate { type_url, value };

    Ok(Response::new()
        .add_attribute("method", "execute_authz_stargate")
        .add_message(msg))
}
