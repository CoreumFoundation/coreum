use crate::error::ContractError;
use crate::msg::{
    DEXOrder, ExecuteMsg, IBCPurpose, InstantiateMsg, QueryIssuanceMsgResponse, QueryMsg, SudoMsg,
    TransferContext,
};
use crate::state::{DENOM, EXTRA_DATA};
use coreum_wasm_sdk::deprecated::core::{CoreumMsg, CoreumResult};
use coreum_wasm_sdk::types::coreum::asset::ft::v1::{
    MsgBurn, MsgMint, QueryTokenRequest, QueryTokenResponse, Token,
};
use coreum_wasm_sdk::types::cosmos::bank::v1beta1::{
    MsgSend,
};
use coreum_wasm_sdk::types::cosmos::base::v1beta1::Coin;
use cosmwasm_std::{entry_point, to_json_binary, CosmosMsg, StdError};
use cosmwasm_std::{Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult, Uint128};
use cw2::set_contract_version;
use std::ops::Div;
use std::string::ToString;
use cosmwasm_schema::schemars::_serde_json::to_string;

// version info for migration info
const CONTRACT_NAME: &str = env!("CARGO_PKG_NAME");
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

const AMOUNT_DISALLOWED_TRIGGER: Uint128 = Uint128::new(7);
const AMOUNT_BURNING_TRIGGER: Uint128 = Uint128::new(101);
const AMOUNT_MINTING_TRIGGER: Uint128 = Uint128::new(105);
const AMOUNT_IGNORE_BURN_RATE_TRIGGER: Uint128 = Uint128::new(108);
const AMOUNT_IGNORE_SEND_COMMISSION_RATE_TRIGGER: Uint128 = Uint128::new(109);
const AMOUNT_BLOCK_IBC_TRIGGER: Uint128 = Uint128::new(110);
const AMOUNT_BLOCK_SMART_CONTRACT_TRIGGER: Uint128 = Uint128::new(111);
const ID_DEX_ORDER_SUFFIX_TRIGGER: &str = "blocked";
const AMOUNT_DEX_EXPECT_TO_SPEND_TRIGGER: Uint128 = Uint128::new(103);
const AMOUNT_DEX_EXPECT_TO_RECEIVE_TRIGGER: Uint128 = Uint128::new(104);

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> CoreumResult<ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    DENOM.save(deps.storage, &msg.denom)?;
    EXTRA_DATA.save(
        deps.storage,
        &msg.issuance_msg.extra_data.unwrap_or_default(),
    )?;

    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("owner", info.sender))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    _deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: ExecuteMsg,
) -> CoreumResult<ContractError> {
    match msg {}
}

#[entry_point]
pub fn sudo(deps: DepsMut, env: Env, msg: SudoMsg) -> CoreumResult<ContractError> {
    match msg {
        SudoMsg::ExtensionTransfer {
            sender,
            recipient,
            transfer_amount,
            commission_amount,
            burn_amount,
            context,
        } => sudo_extension_transfer(
            deps,
            env,
            transfer_amount,
            sender,
            recipient,
            commission_amount,
            burn_amount,
            context,
        ),
        SudoMsg::ExtensionPlaceOrder {
            order,
            expected_to_spend,
            expected_to_receive,
        } => sudo_extension_place_order(order, expected_to_spend, expected_to_receive),
    }
}

pub fn sudo_extension_transfer(
    deps: DepsMut,
    env: Env,
    amount: Uint128,
    sender: String,
    recipient: String,
    commission_amount: Uint128,
    burn_amount: Uint128,
    context: TransferContext,
) -> CoreumResult<ContractError> {
    if amount.is_zero() {
        return Err(ContractError::InvalidAmountError {});
    }

    let rsp = Response::new().add_attribute("method", "execute_transfer");
    if recipient == env.contract.address.as_str() {
        return Ok(rsp.add_attribute("skip_checks", "self_recipient"));
    }

    if amount == AMOUNT_DISALLOWED_TRIGGER {
        return Err(ContractError::Std(StdError::generic_err(
            "7 is not allowed",
        )));
    }

    let denom = DENOM.load(deps.storage)?;

    let token = query_token(deps.as_ref(), &denom)?;

    if !&token.features.is_empty() {
        assert_block_smart_contracts(&context, &recipient, &token, amount)?;

        assert_ibc(&context, &recipient, &token, amount)?;

        if amount == AMOUNT_BURNING_TRIGGER {
            return assert_burning(env.contract.address.as_str(), amount, &token);
        }

        if amount == AMOUNT_MINTING_TRIGGER {
            return assert_minting(
                env.contract.address.as_str(),
                sender.as_ref(),
                &recipient,
                amount,
                &token,
            );
        }
    }

    let transfer_msg = MsgSend {
        from_address: env.contract.address.to_string(),
        to_address: recipient.to_string(),
        amount: [Coin {
            denom: token.denom.to_string(),
            amount: amount.to_string(),
        }]
            .to_vec(),
    };

    let mut response = rsp.add_message(CosmosMsg::Any(transfer_msg.to_any()));

    if !commission_amount.is_zero() {
        response = assert_send_commission_rate(
            env.contract.address.as_str(),
            response,
            sender.as_ref(),
            amount,
            &token,
            commission_amount,
        )?;
    }

    if !burn_amount.is_zero() {
        response = assert_burn_rate(
            env.contract.address.as_str(),
            response,
            sender.as_ref(),
            amount,
            &token,
            burn_amount,
        )?;
    }

    Ok(response)
}

pub fn sudo_extension_place_order(
    order: DEXOrder,
    expected_to_spend: Coin,
    expected_to_receive: Coin,
) -> CoreumResult<ContractError> {
    if order.id.ends_with(ID_DEX_ORDER_SUFFIX_TRIGGER)
        || expected_to_spend.amount == AMOUNT_DEX_EXPECT_TO_SPEND_TRIGGER.to_string()
        || expected_to_receive.amount == AMOUNT_DEX_EXPECT_TO_RECEIVE_TRIGGER.to_string()
    {
        return Err(ContractError::DEXOrderPlacementError {});
    }

    let order_data = to_string(&order).
        map_err(|_| ContractError::Std(StdError::generic_err("failed to serialize order to json string")))?;

    Ok(
        Response::new()
            .add_attribute("method", "extension_place_order")
            .add_attribute("order_data", order_data)
    )
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::QueryIssuanceMsg {} => query_issuance_msg(deps),
    }
}

fn query_issuance_msg(deps: Deps) -> StdResult<Binary> {
    let test = EXTRA_DATA.load(deps.storage).ok();
    let resp = QueryIssuanceMsgResponse { test };
    to_json_binary(&resp)
}

fn assert_burning(contract: &str, amount: Uint128, token: &Token) -> CoreumResult<ContractError> {
    let burn_message = MsgBurn {
        sender: contract.to_string(),
        coin: Some(Coin {
            denom: token.denom.to_string(),
            amount: amount.to_string(),
        }),
    };

    Ok(Response::new()
        .add_attribute("method", "burn")
        .add_message(CosmosMsg::Any(burn_message.to_any())))
}

fn assert_minting(
    contract: &str,
    sender: &str,
    recipient: &str,
    amount: Uint128,
    token: &Token,
) -> CoreumResult<ContractError> {
    let mint_message = MsgMint {
        sender: contract.to_string(),
        coin: Some(Coin {
            denom: token.denom.to_string(),
            amount: amount.to_string(),
        }),
        recipient: recipient.to_string(),
    };

    let return_fund_msg = MsgSend {
        from_address: contract.to_string(),
        to_address: sender.to_string(),
        amount: [Coin {
            denom: token.denom.to_string(),
            amount: amount.to_string(),
        }]
            .to_vec(),
    };

    Ok(Response::new()
        .add_attribute("method", "mint")
        .add_message(CosmosMsg::Any(mint_message.to_any()))
        .add_message(CosmosMsg::Any(return_fund_msg.to_any())))
}

fn assert_block_smart_contracts(
    context: &TransferContext,
    recipient: &str,
    token: &Token,
    amount: Uint128,
) -> Result<(), ContractError> {
    if recipient.to_string() == token.issuer || recipient == token.extension_cw_address {
        return Ok(());
    }

    if context.recipient_is_smart_contract && amount == AMOUNT_BLOCK_SMART_CONTRACT_TRIGGER {
        return Err(ContractError::SmartContractBlocked {});
    }

    Ok(())
}

fn assert_ibc(
    context: &TransferContext,
    recipient: &str,
    token: &Token,
    amount: Uint128,
) -> Result<(), ContractError> {
    if recipient == token.admin || recipient == token.extension_cw_address {
        return Ok(());
    }

    if context.ibc_purpose == IBCPurpose::Out && amount == AMOUNT_BLOCK_IBC_TRIGGER {
        return Err(ContractError::IBCDisabled {});
    }

    Ok(())
}

fn assert_send_commission_rate(
    contract: &str,
    response: Response<CoreumMsg>,
    sender: &str,
    amount: Uint128,
    token: &Token,
    commission_amount: Uint128,
) -> CoreumResult<ContractError> {
    if amount == AMOUNT_IGNORE_SEND_COMMISSION_RATE_TRIGGER {
        let refund_commission_msg = MsgSend {
            from_address: contract.to_string(),
            to_address: sender.to_string(),
            amount: [Coin {
                denom: token.denom.to_string(),
                amount: commission_amount.to_string(),
            }]
                .to_vec(),
        };

        return Ok(response
            .add_attribute("send_commission_rate_refund", commission_amount.to_string())
            .add_message(CosmosMsg::Any(refund_commission_msg.to_any())));
    }

    // if token has an admin, send half of the commission to the admin and let the extension keep
    // the rest of the commission
    if !token.admin.is_empty() {
        let admin_commission_amount = commission_amount.div(Uint128::new(2));
        if admin_commission_amount.is_zero() {
            return Ok(response);
        }

        let admin_commission_msg = MsgSend {
            from_address: contract.to_string(),
            to_address: token.admin.to_string(),
            amount: [Coin {
                denom: token.denom.to_string(),
                amount: admin_commission_amount.to_string(),
            }]
                .to_vec(),
        };

        return Ok(response
            .add_attribute(
                "admin_send_commission_amount",
                admin_commission_amount.to_string(),
            )
            .add_message(CosmosMsg::Any(admin_commission_msg.to_any())));
    }

    // else, let the extension keep all the commission
    Ok(response)
}

fn assert_burn_rate(
    contract: &str,
    response: Response<CoreumMsg>,
    sender: &str,
    amount: Uint128,
    token: &Token,
    burn_amount: Uint128,
) -> CoreumResult<ContractError> {
    if amount == AMOUNT_IGNORE_BURN_RATE_TRIGGER {
        let refund_burn_rate_msg = MsgSend {
            from_address: contract.to_string(),
            to_address: sender.to_string(),
            amount: [Coin {
                denom: token.denom.to_string(),
                amount: burn_amount.to_string(),
            }]
                .to_vec(),
        };

        return Ok(response
            .add_attribute("burn_rate_refund", burn_amount.to_string())
            .add_message(CosmosMsg::Any(refund_burn_rate_msg.to_any())));
    }

    let burn_message = MsgBurn {
        sender: contract.to_string(),
        coin: Some(Coin {
            denom: token.denom.to_string(),
            amount: burn_amount.to_string(),
        }),
    };

    Ok(response
        .add_attribute("burn_amount", burn_amount)
        .add_message(CosmosMsg::Any(burn_message.to_any())))
}

fn query_token(deps: Deps, denom: &str) -> StdResult<Token> {
    let request = QueryTokenRequest {
        denom: denom.to_string(),
    };
    let token: QueryTokenResponse = request.query(&deps.querier)?;
    Ok(token.token.unwrap_or_default())
}
