use coreum_wasm_sdk::types::coreum::asset::ft::v1::{
    MsgBurn, MsgClawback, MsgClearAdmin, MsgFreeze, MsgGloballyFreeze, MsgGloballyUnfreeze, MsgIssue, MsgMint, MsgSetFrozen, MsgSetWhitelistedLimit, MsgTransferAdmin, MsgUnfreeze, MsgUpgradeTokenV1
};
use coreum_wasm_sdk::types::cosmos::base::v1beta1::Coin;
use cosmwasm_std::{entry_point, Binary, CosmosMsg, Deps, StdResult};
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
        extension_settings: msg.extension_settings,
    };

    let issue_bytes = issue.to_proto_bytes();

    let issue_msg = CosmosMsg::Stargate {
        type_url: issue.to_any().type_url,
        value: Binary::from(issue_bytes),
    };

    let denom = format!("{}-{}", msg.subunit, env.contract.address).to_lowercase();

    DENOM.save(deps.storage, &denom)?;

    Ok(Response::new()
        .add_attribute("owner", info.sender)
        .add_attribute("denom", denom)
        .add_message(issue_msg))
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
    amount: u128,
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

    let mint_bytes = mint.to_proto_bytes();

    let mint_msg = CosmosMsg::Stargate {
        type_url: mint.to_any().type_url,
        value: Binary::from(mint_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "mint")
        .add_attribute("denom", denom)
        .add_attribute("amount", amount.to_string())
        .add_message(mint_msg))
}

fn burn(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    amount: u128,
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

    let burn_bytes = burn.to_proto_bytes();

    let burn_msg = CosmosMsg::Stargate {
        type_url: burn.to_any().type_url,
        value: Binary::from(burn_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "burn")
        .add_attribute("denom", denom)
        .add_attribute("amount", amount.to_string())
        .add_message(burn_msg))
}

fn freeze(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    account: String,
    amount: u128,
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

    let freeze_bytes = freeze.to_proto_bytes();

    let freeze_msg = CosmosMsg::Stargate {
        type_url: freeze.to_any().type_url,
        value: Binary::from(freeze_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "freeze")
        .add_attribute("denom", denom)
        .add_attribute("amount", amount.to_string())
        .add_message(freeze_msg))
}

fn unfreeze(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    account: String,
    amount: u128,
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

    let unfreeze_bytes = unfreeze.to_proto_bytes();

    let unfreeze_msg = CosmosMsg::Stargate {
        type_url: unfreeze.to_any().type_url,
        value: Binary::from(unfreeze_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "unfreeze")
        .add_attribute("denom", denom)
        .add_attribute("amount", amount.to_string())
        .add_message(unfreeze_msg))
}

fn set_frozen(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    account: String,
    amount: u128,
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

    let set_frozen_bytes = set_frozen.to_proto_bytes();

    let set_frozen_msg = CosmosMsg::Stargate {
        type_url: set_frozen.to_any().type_url,
        value: Binary::from(set_frozen_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "set_frozen")
        .add_attribute("denom", denom)
        .add_attribute("amount", amount.to_string())
        .add_message(set_frozen_msg))
}

fn globally_freeze(deps: DepsMut, env: Env, info: MessageInfo) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let denom = DENOM.load(deps.storage)?;

    let globally_freeze = MsgGloballyFreeze {
        sender: env.contract.address.to_string(),
        denom: denom.clone(),
    };

    let globally_freeze_bytes = globally_freeze.to_proto_bytes();

    let globally_freeze_msg = CosmosMsg::Stargate {
        type_url: globally_freeze.to_any().type_url,
        value: Binary::from(globally_freeze_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "globally_freeze")
        .add_attribute("denom", denom)
        .add_message(globally_freeze_msg))
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

    let globally_unfreeze_bytes = globally_unfreeze.to_proto_bytes();

    let globally_unfreeze_msg = CosmosMsg::Stargate {
        type_url: globally_unfreeze.to_any().type_url,
        value: Binary::from(globally_unfreeze_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "globally_unfreeze")
        .add_attribute("denom", denom)
        .add_message(globally_unfreeze_msg))
}

fn clawback(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    account: String,
    amount: u128,
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

    let clawback_bytes = clawback.to_proto_bytes();

    let clawback_msg = CosmosMsg::Stargate {
        type_url: clawback.to_any().type_url,
        value: Binary::from(clawback_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "clawback")
        .add_attribute("denom", denom)
        .add_message(clawback_msg))
}

fn set_whitelisted_limit(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    account: String,
    amount: u128,
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

    let set_whitelisted_limit_bytes = set_whitelisted_limit.to_proto_bytes();

    let set_whitelisted_limit_msg = CosmosMsg::Stargate {
        type_url: set_whitelisted_limit.to_any().type_url,
        value: Binary::from(set_whitelisted_limit_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "set_whitelisted_limit")
        .add_attribute("denom", denom)
        .add_attribute("amount", amount.to_string())
        .add_message(set_whitelisted_limit_msg))
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

    let transfer_admin_bytes = transfer_admin.to_proto_bytes();

    let transfer_admin_msg = CosmosMsg::Stargate {
        type_url: transfer_admin.to_any().type_url,
        value: Binary::from(transfer_admin_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "transfer_admin")
        .add_attribute("denom", denom)
        .add_message(transfer_admin_msg))
}

fn clear_admin(deps: DepsMut, env: Env, info: MessageInfo) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let denom = DENOM.load(deps.storage)?;

    let clear_admin = MsgClearAdmin {
        sender: env.contract.address.to_string(),
        denom: denom.clone(),
    };

    let clear_admin_bytes = clear_admin.to_proto_bytes();

    let clear_admin_msg = CosmosMsg::Stargate {
        type_url: clear_admin.to_any().type_url,
        value: Binary::from(clear_admin_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "clear_admin")
        .add_attribute("denom", denom)
        .add_message(clear_admin_msg))
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

    let upgrade_token_v1_bytes = upgrade_token_v1.to_proto_bytes();

    let upgrade_token_v1_msg = CosmosMsg::Stargate {
        type_url: upgrade_token_v1.to_any().type_url,
        value: Binary::from(upgrade_token_v1_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "upgrade_token_v1")
        .add_attribute("denom", denom)
        .add_attribute("ibc_enabled", ibc_enabled.to_string())
        .add_message(upgrade_token_v1_msg))
}

// ********** Queries **********
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(_deps: Deps, _env: Env, _msg: QueryMsg) -> StdResult<Binary> {
    Ok(Binary::default())
}
