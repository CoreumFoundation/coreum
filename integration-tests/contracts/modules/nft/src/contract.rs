use coreum_wasm_sdk::types::coreum::asset::nft::v1::{
    DataBytes, DataDynamic, DataDynamicIndexedItem, DataDynamicItem, DataEditor,
    MsgAddToClassWhitelist, MsgAddToWhitelist, MsgBurn, MsgClassFreeze, MsgClassUnfreeze,
    MsgFreeze, MsgIssueClass, MsgMint, MsgRemoveFromClassWhitelist, MsgRemoveFromWhitelist,
    MsgUnfreeze, MsgUpdateData,
};
use coreum_wasm_sdk::types::cosmos::nft::v1beta1::MsgSend;
use cosmwasm_std::{
    entry_point, Binary, CosmosMsg, Deps, DepsMut, Env, MessageInfo, Response, StdResult,
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
        data,
        features: msg.features.unwrap_or_default(),
        royalty_rate: msg.royalty_rate.unwrap_or_default(),
    };

    let issue_bytes = issue.to_proto_bytes();

    let issue_msg = CosmosMsg::Stargate {
        type_url: issue.to_any().type_url,
        value: Binary::from(issue_bytes),
    };

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
        data,
        recipient: recipient.unwrap_or_default(),
    };

    let mint_bytes = mint.to_proto_bytes();

    let msg = CosmosMsg::Stargate {
        type_url: mint.to_any().type_url,
        value: Binary::from(mint_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "mint_immutable")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(msg))
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
        data,
        recipient: recipient.unwrap_or_default(),
    };

    let mint_bytes = mint.to_proto_bytes();

    let msg = CosmosMsg::Stargate {
        type_url: mint.to_any().type_url,
        value: Binary::from(mint_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "mint_mutable")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(msg))
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

    let modify_data_bytes = modify_data.to_proto_bytes();

    let msg = CosmosMsg::Stargate {
        type_url: modify_data.to_any().type_url,
        value: Binary::from(modify_data_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "modify_data")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(msg))
}

fn burn(deps: DepsMut, env: Env, info: MessageInfo, id: String) -> Result<Response, ContractError> {
    assert_owner(deps.storage, &info.sender)?;
    let class_id = CLASS_ID.load(deps.storage)?;

    let burn = MsgBurn {
        sender: env.contract.address.to_string(),
        class_id: class_id.clone(),
        id: id.clone(),
    };

    let burn_bytes = burn.to_proto_bytes();

    let burn_msg = CosmosMsg::Stargate {
        type_url: burn.to_any().type_url,
        value: Binary::from(burn_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "burn")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(burn_msg))
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

    let freeze_bytes = freeze.to_proto_bytes();

    let freeze_msg = CosmosMsg::Stargate {
        type_url: freeze.to_any().type_url,
        value: Binary::from(freeze_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "freeze")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(freeze_msg))
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

    let unfreeze_bytes = unfreeze.to_proto_bytes();

    let unfreeze_msg = CosmosMsg::Stargate {
        type_url: unfreeze.to_any().type_url,
        value: Binary::from(unfreeze_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "unfreeze")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(unfreeze_msg))
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

    let add_to_whitelist_bytes = add_to_whitelist.to_proto_bytes();

    let add_to_whitelist_msg = CosmosMsg::Stargate {
        type_url: add_to_whitelist.to_any().type_url,
        value: Binary::from(add_to_whitelist_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "add_to_white_list")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(add_to_whitelist_msg))
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

    let remove_from_whitelist_bytes = remove_from_whitelist.to_proto_bytes();

    let remove_from_whitelist_msg = CosmosMsg::Stargate {
        type_url: remove_from_whitelist.to_any().type_url,
        value: Binary::from(remove_from_whitelist_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "remove_from_white_list")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(remove_from_whitelist_msg))
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

    let send_bytes = send.to_proto_bytes();

    let send_msg = CosmosMsg::Stargate {
        type_url: send.to_any().type_url,
        value: Binary::from(send_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "send")
        .add_attribute("class_id", class_id)
        .add_attribute("id", id)
        .add_message(send_msg))
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

    let class_freeze_bytes = class_freeze.to_proto_bytes();

    let class_freeze_msg = CosmosMsg::Stargate {
        type_url: class_freeze.to_any().type_url,
        value: Binary::from(class_freeze_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "class_freeze")
        .add_attribute("class_id", class_id)
        .add_attribute("account", account)
        .add_message(class_freeze_msg))
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

    let class_unfreeze_bytes = class_unfreeze.to_proto_bytes();

    let class_unfreeze_msg = CosmosMsg::Stargate {
        type_url: class_unfreeze.to_any().type_url,
        value: Binary::from(class_unfreeze_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "class_unfreeze")
        .add_attribute("class_id", class_id)
        .add_attribute("account", account)
        .add_message(class_unfreeze_msg))
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

    let add_to_class_whitelist_bytes = add_to_class_whitelist.to_proto_bytes();

    let add_to_class_whitelist_msg = CosmosMsg::Stargate {
        type_url: add_to_class_whitelist.to_any().type_url,
        value: Binary::from(add_to_class_whitelist_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "add_to_class_whitelist")
        .add_attribute("class_id", class_id)
        .add_attribute("account", account)
        .add_message(add_to_class_whitelist_msg))
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

    let remove_from_class_whitelist_bytes = remove_from_class_whitelist.to_proto_bytes();

    let remove_from_class_whitelist_msg = CosmosMsg::Stargate {
        type_url: remove_from_class_whitelist.to_any().type_url,
        value: Binary::from(remove_from_class_whitelist_bytes),
    };

    Ok(Response::new()
        .add_attribute("method", "remove_from_class_whitelist")
        .add_attribute("class_id", class_id)
        .add_attribute("account", account)
        .add_message(remove_from_class_whitelist_msg))
}

// ********** Queries **********

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(_deps: Deps, _env: Env, _msg: QueryMsg) -> StdResult<Binary> {
    Ok(Binary::default())
}
