use coreum_wasm_sdk::core::CoreumResult;
use coreum_wasm_sdk::shim;
use coreum_wasm_sdk::types::cosmos::authz::v1beta1::MsgExec;
use coreum_wasm_sdk::types::cosmos::bank::v1beta1::MsgSend as BankMsg;
use coreum_wasm_sdk::types::cosmos::base::v1beta1::Coin;
use coreum_wasm_sdk::types::cosmos::nft::v1beta1::MsgSend;

#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{CosmosMsg, DepsMut, Env, MessageInfo, Response};
use cw2::set_contract_version;
use cw_utils::one_coin;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg};
use crate::state::{Offer, NFT_OFFERS};

const CONTRACT_NAME: &str = env!("CARGO_PKG_NAME");
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    Ok(Response::new()
        .add_attribute("contract", CONTRACT_NAME)
        .add_attribute("action", "instantiate"))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> CoreumResult<ContractError> {
    match msg {
        ExecuteMsg::OfferNft {
            class_id,
            id,
            price,
        } => offer_nft(
            deps,
            env,
            info,
            class_id,
            id,
            Coin {
                denom: price.denom.to_string(),
                amount: price.amount.to_string(),
            },
        ),
        ExecuteMsg::AcceptNftOffer { class_id, id } => accept_offer(deps, env, info, class_id, id),
    }
}

// The contract must have been granted authorization to send the NFT before execution or it will fail.
// We will send the NFT to the contract to be able to sell it when someone provides the price.
fn offer_nft(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    class_id: String,
    id: String,
    price: Coin,
) -> CoreumResult<ContractError> {
    let nft_send = MsgSend {
        class_id: class_id.clone(),
        id: id.clone(),
        sender: info.sender.to_string(),
        receiver: env.contract.address.to_string(),
    }
    .to_any();

    let exec = MsgExec {
        grantee: env.contract.address.to_string(),
        msgs: vec![shim::Any {
            type_url: nft_send.type_url,
            value: nft_send.value.to_vec(),
        }],
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
        .add_message(CosmosMsg::Any(exec.to_any())))
}

fn accept_offer(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    class_id: String,
    id: String,
) -> CoreumResult<ContractError> {
    let offer = NFT_OFFERS.load(deps.storage, (class_id.clone(), id.clone()))?;

    let coin = one_coin(&info)?;
    let coin = Coin {
        denom: coin.denom.to_string(),
        amount: coin.amount.to_string(),
    };
    if coin != offer.price {
        return Err(ContractError::InvalidFundsAmount {});
    }

    let nft_send = MsgSend {
        class_id,
        id,
        sender: env.contract.address.to_string(),
        receiver: info.sender.to_string(),
    };

    let send_funds = BankMsg {
        from_address: env.contract.address.to_string(),
        to_address: offer.address.to_string(),
        amount: info
            .funds
            .iter()
            .map(|coin| Coin {
                denom: coin.denom.clone(),
                amount: coin.amount.to_string(),
            })
            .collect::<Vec<Coin>>(),
    };

    Ok(Response::new()
        .add_attribute("method", "execute_accept_nft_offer")
        .add_messages([
            CosmosMsg::Any(nft_send.to_any()),
            CosmosMsg::Any(send_funds.to_any()),
        ]))
}
