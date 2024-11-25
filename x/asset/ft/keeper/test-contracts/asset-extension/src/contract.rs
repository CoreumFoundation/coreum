use crate::error::ContractError;
use crate::msg::{
    ExecuteMsg, IBCPurpose, InstantiateMsg, QueryIssuanceMsgResponse, QueryMsg, SudoMsg,
    TransferContext,
};
use crate::state::{DENOM, EXTRA_DATA};
use coreum_wasm_sdk::deprecated::assetft::{FREEZING, WHITELISTING};
use coreum_wasm_sdk::deprecated::core::{CoreumMsg, CoreumResult};
use coreum_wasm_sdk::types::coreum::asset::ft::v1::{
    MsgBurn, MsgMint, QueryFrozenBalanceRequest, QueryFrozenBalanceResponse, QueryTokenRequest,
    QueryTokenResponse, QueryWhitelistedBalanceRequest, QueryWhitelistedBalanceResponse, Token,
};
use coreum_wasm_sdk::types::cosmos::bank::v1beta1::{
    MsgSend, QueryBalanceRequest, QueryBalanceResponse,
};
use coreum_wasm_sdk::types::cosmos::base::v1beta1::Coin;
use cosmwasm_std::{entry_point, to_json_binary, CosmosMsg, StdError};
use cosmwasm_std::{Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult, Uint128};
use cw2::set_contract_version;
use std::ops::Div;

// version info for migration info
const CONTRACT_NAME: &str = env!("CARGO_PKG_NAME");
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

const AMOUNT_DISALLOWED_TRIGGER: Uint128 = Uint128::new(7);
const AMOUNT_IGNORE_WHITELISTING_TRIGGER: Uint128 = Uint128::new(49);
const AMOUNT_IGNORE_FREEZING_TRIGGER: Uint128 = Uint128::new(79);
const AMOUNT_BURNING_TRIGGER: Uint128 = Uint128::new(101);
const AMOUNT_MINTING_TRIGGER: Uint128 = Uint128::new(105);
const AMOUNT_IGNORE_BURN_RATE_TRIGGER: Uint128 = Uint128::new(108);
const AMOUNT_IGNORE_SEND_COMMISSION_RATE_TRIGGER: Uint128 = Uint128::new(109);
const AMOUNT_BLOCK_IBC_TRIGGER: Uint128 = Uint128::new(110);
const AMOUNT_BLOCK_SMART_CONTRACT_TRIGGER: Uint128 = Uint128::new(111);

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
        if token.features.contains(&(FREEZING as i32)) {
            assert_freezing(&context, deps.as_ref(), sender.as_ref(), &token, amount)?;
        }

        if token.features.contains(&(WHITELISTING as i32)) {
            assert_whitelisting(&context, deps.as_ref(), &recipient, &token, amount)?;
        }

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

fn assert_freezing(
    context: &TransferContext,
    deps: Deps,
    account: &str,
    token: &Token,
    amount: Uint128,
) -> Result<(), ContractError> {
    // Allow any amount if recipient is admin
    if token.admin == account {
        return Ok(());
    }

    // Ignore freezing if the transfer is an IBC transfer in. In case of IBC transfer coming into the chain
    // source account is the escrow account and since we don't want to allow freeze of every
    // escrow address we ignore freezing for incoming ibc transfers.
    if context.ibc_purpose == IBCPurpose::In {
        return Ok(());
    }

    if amount == AMOUNT_IGNORE_FREEZING_TRIGGER {
        return Ok(());
    }

    if token.globally_frozen {
        return Err(ContractError::FreezingError {});
    }

    let bank_balance = query_bank_balance(deps, account, &token.denom)?;
    let frozen_balance = query_frozen_balance(deps, account, &token.denom)?;

    // the amount is already deducted from the balance, so you can omit it from both sides
    if frozen_balance.amount.parse::<u128>().unwrap() > bank_balance.amount.parse::<u128>().unwrap()
    {
        return Err(ContractError::FreezingError {});
    }

    Ok(())
}

fn assert_whitelisting(
    context: &TransferContext,
    deps: Deps,
    account: &str,
    token: &Token,
    amount: Uint128,
) -> Result<(), ContractError> {
    // Allow any amount if recipient is admin
    if token.admin == account {
        return Ok(());
    }

    // Ignore whitelising if the transfer is an IBC transfer. In case of IBC transfer
    // destination account is the escrow account and since we don't want to whitelist every
    // escrow address we ignore whitelisting for outgoing ibc transfers.
    if context.ibc_purpose == IBCPurpose::Out {
        return Ok(());
    }

    if amount == AMOUNT_IGNORE_WHITELISTING_TRIGGER {
        return Ok(());
    }

    let bank_balance = query_bank_balance(deps, account, &token.denom)?;
    let whitelisted_balance = query_whitelisted_balance(deps, account, &token.denom)?;

    if amount + Uint128::from(bank_balance.amount.parse::<u128>().unwrap())
        > Uint128::from(whitelisted_balance.amount.parse::<u128>().unwrap())
    {
        return Err(ContractError::WhitelistingError {});
    }

    Ok(())
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

fn query_frozen_balance(deps: Deps, account: &str, denom: &str) -> StdResult<Coin> {
    let request = QueryFrozenBalanceRequest {
        account: account.to_string(),
        denom: denom.to_string(),
    };
    let frozen_balance: QueryFrozenBalanceResponse = request.query(&deps.querier)?;
    Ok(frozen_balance.balance.unwrap_or_default())
}

fn query_whitelisted_balance(deps: Deps, account: &str, denom: &str) -> StdResult<Coin> {
    let request = QueryWhitelistedBalanceRequest {
        account: account.to_string(),
        denom: denom.to_string(),
    };
    let whitelisted_balance: QueryWhitelistedBalanceResponse = request.query(&deps.querier)?;
    Ok(whitelisted_balance.balance.unwrap_or_default())
}

fn query_bank_balance(deps: Deps, account: &str, denom: &str) -> StdResult<Coin> {
    let request = QueryBalanceRequest {
        address: account.to_string(),
        denom: denom.to_string(),
    };
    let bank_balance: QueryBalanceResponse = request.query(&deps.querier)?;
    Ok(bank_balance.balance.unwrap_or_default())
}

fn query_token(deps: Deps, denom: &str) -> StdResult<Token> {
    let request = QueryTokenRequest {
        denom: denom.to_string(),
    };
    let token: QueryTokenResponse = request.query(&deps.querier)?;
    Ok(token.token.unwrap_or_default())
}
