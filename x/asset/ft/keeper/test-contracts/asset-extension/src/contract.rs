use cosmwasm_std::{entry_point, to_json_binary, StdError};
use cosmwasm_std::{BalanceResponse, BankQuery};
use cosmwasm_std::{Binary, Coin, Deps, DepsMut, Env, MessageInfo, Response, StdResult, Uint128};
use cw2::set_contract_version;
use std::ops::Div;

use crate::error::ContractError;
use coreum_wasm_sdk::assetft::{
    self, FrozenBalanceResponse, Query, Token, TokenResponse, WhitelistedBalanceResponse,
};
use coreum_wasm_sdk::core::{CoreumMsg, CoreumQueries, CoreumResult};

use crate::msg::{
    ExecuteMsg, IBCPurpose, InstantiateMsg, QueryIssuanceMsgResponse, QueryMsg, SudoMsg,
    TransferContext,
};
use crate::state::{DENOM, EXTRA_DATA};

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
    _deps: DepsMut<CoreumQueries>,
    _env: Env,
    _info: MessageInfo,
    msg: ExecuteMsg,
) -> CoreumResult<ContractError> {
    match msg {}
}

#[entry_point]
pub fn sudo(deps: DepsMut<CoreumQueries>, env: Env, msg: SudoMsg) -> CoreumResult<ContractError> {
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
    deps: DepsMut<CoreumQueries>,
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
    if recipient == env.contract.address {
        return Ok(rsp.add_attribute("skip_checks", "self_recipient"));
    }

    if amount == AMOUNT_DISALLOWED_TRIGGER {
        return Err(ContractError::Std(StdError::generic_err(
            "7 is not allowed",
        )));
    }

    let denom = DENOM.load(deps.storage)?;

    let token = query_token(deps.as_ref(), &denom)?;

    if let Some(features) = &token.features {
        if features.contains(&assetft::FREEZING) {
            assert_freezing(&context, deps.as_ref(), sender.as_ref(), &token, amount)?;
        }

        if features.contains(&assetft::WHITELISTING) {
            assert_whitelisting(&context, deps.as_ref(), &recipient, &token, amount)?;
        }

        assert_block_smart_contracts(&context, &recipient, &token, amount)?;

        assert_ibc(&context, &recipient, &token, amount)?;

        if amount == AMOUNT_BURNING_TRIGGER {
            return assert_burning(amount, &token);
        }

        if amount == AMOUNT_MINTING_TRIGGER {
            return assert_minting(sender.as_ref(), &recipient, amount, &token);
        }
    }

    let transfer_msg = cosmwasm_std::BankMsg::Send {
        to_address: recipient.to_string(),
        amount: vec![Coin { amount, denom }],
    };

    let mut response = rsp.add_message(transfer_msg);

    if !commission_amount.is_zero() {
        response = assert_send_commission_rate(
            response,
            sender.as_ref(),
            amount,
            &token,
            commission_amount,
        )?;
    }

    if !burn_amount.is_zero() {
        response = assert_burn_rate(response, sender.as_ref(), amount, &token, burn_amount)?;
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
    deps: Deps<CoreumQueries>,
    account: &str,
    token: &Token,
    amount: Uint128,
) -> Result<(), ContractError> {
    // Allow any amount if recipient is admin
    if token.admin == Some(account.to_string()) {
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

    if token.globally_frozen == Some(true) {
        return Err(ContractError::FreezingError {});
    }

    let bank_balance = query_bank_balance(deps, account, &token.denom)?;
    let frozen_balance = query_frozen_balance(deps, account, &token.denom)?;

    // the amount is already deducted from the balance, so you can omit it from both sides
    if frozen_balance.amount > bank_balance.amount {
        return Err(ContractError::FreezingError {});
    }

    Ok(())
}

fn assert_whitelisting(
    context: &TransferContext,
    deps: Deps<CoreumQueries>,
    account: &str,
    token: &Token,
    amount: Uint128,
) -> Result<(), ContractError> {
    // Allow any amount if recipient is admin
    if token.admin == Some(account.to_string()) {
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

    if amount + bank_balance.amount > whitelisted_balance.amount {
        return Err(ContractError::WhitelistingError {});
    }

    Ok(())
}

fn assert_burning(amount: Uint128, token: &Token) -> CoreumResult<ContractError> {
    let burn_message = CoreumMsg::AssetFT(assetft::Msg::Burn {
        coin: cosmwasm_std::coin(amount.u128(), &token.denom),
    });

    return Ok(Response::new()
        .add_attribute("method", "burn")
        .add_message(burn_message));
}

fn assert_minting(
    sender: &str,
    recipient: &str,
    amount: Uint128,
    token: &Token,
) -> CoreumResult<ContractError> {
    let mint_message = CoreumMsg::AssetFT(assetft::Msg::Mint {
        coin: cosmwasm_std::coin(amount.u128(), &token.denom),
        recipient: Some(recipient.to_string()),
    });

    let return_fund_msg = cosmwasm_std::BankMsg::Send {
        to_address: sender.to_string(),
        amount: vec![Coin {
            amount,
            denom: token.denom.clone(),
        }],
    };

    return Ok(Response::new()
        .add_attribute("method", "mint")
        .add_message(mint_message)
        .add_message(return_fund_msg));
}

fn assert_block_smart_contracts(
    context: &TransferContext,
    recipient: &str,
    token: &Token,
    amount: Uint128,
) -> Result<(), ContractError> {
    if recipient.to_string() == token.issuer
        || Some(recipient.to_string()) == token.extension_cw_address
    {
        return Ok(());
    }

    if context.recipient_is_smart_contract && amount == AMOUNT_BLOCK_SMART_CONTRACT_TRIGGER {
        return Err(ContractError::SmartContractBlocked {});
    }

    return Ok(());
}

fn assert_ibc(
    context: &TransferContext,
    recipient: &str,
    token: &Token,
    amount: Uint128,
) -> Result<(), ContractError> {
    if Some(recipient.to_string()) == token.admin
        || Some(recipient.to_string()) == token.extension_cw_address
    {
        return Ok(());
    }

    if context.ibc_purpose == IBCPurpose::Out && amount == AMOUNT_BLOCK_IBC_TRIGGER {
        return Err(ContractError::IBCDisabled {});
    }

    return Ok(());
}

fn assert_send_commission_rate(
    response: Response<CoreumMsg>,
    sender: &str,
    amount: Uint128,
    token: &Token,
    commission_amount: Uint128,
) -> CoreumResult<ContractError> {
    if amount == AMOUNT_IGNORE_SEND_COMMISSION_RATE_TRIGGER {
        let refund_commission_msg = cosmwasm_std::BankMsg::Send {
            to_address: sender.to_string(),
            amount: vec![Coin {
                amount: commission_amount,
                denom: token.denom.to_string(),
            }],
        };

        return Ok(response
            .add_attribute("send_commission_rate_refund", commission_amount.to_string())
            .add_message(refund_commission_msg));
    }

    // if token has an admin, send half of the commission to the admin and let the extension keep
    // the rest of the commission
    if let Some(admin) = &token.admin {
        let admin_commission_amount = commission_amount.div(Uint128::new(2));
        if admin_commission_amount.is_zero() {
            return Ok(response);
        }

        let admin_commission_msg = cosmwasm_std::BankMsg::Send {
            to_address: admin.to_string(),
            amount: vec![Coin {
                amount: admin_commission_amount,
                denom: token.denom.to_string(),
            }],
        };
        return Ok(response
            .add_attribute(
                "admin_send_commission_amount",
                admin_commission_amount.to_string(),
            )
            .add_message(admin_commission_msg));
    }

    // else, let the extension keep all the commission
    Ok(response)
}

fn assert_burn_rate(
    response: Response<CoreumMsg>,
    sender: &str,
    amount: Uint128,
    token: &Token,
    burn_amount: Uint128,
) -> CoreumResult<ContractError> {
    if amount == AMOUNT_IGNORE_BURN_RATE_TRIGGER {
        let refund_burn_rate_msg = cosmwasm_std::BankMsg::Send {
            to_address: sender.to_string(),
            amount: vec![Coin {
                amount: burn_amount,
                denom: token.denom.to_string(),
            }],
        };

        return Ok(response
            .add_attribute("burn_rate_refund", burn_amount.to_string())
            .add_message(refund_burn_rate_msg));
    }

    let burn_message = CoreumMsg::AssetFT(assetft::Msg::Burn {
        coin: cosmwasm_std::coin(burn_amount.u128(), &token.denom),
    });

    Ok(response
        .add_attribute("burn_amount", burn_amount)
        .add_message(burn_message))
}

fn query_frozen_balance(deps: Deps<CoreumQueries>, account: &str, denom: &str) -> StdResult<Coin> {
    let frozen_balance: FrozenBalanceResponse = deps.querier.query(
        &CoreumQueries::AssetFT(Query::FrozenBalance {
            account: account.to_string(),
            denom: denom.to_string(),
        })
        .into(),
    )?;
    Ok(frozen_balance.balance)
}

fn query_whitelisted_balance(
    deps: Deps<CoreumQueries>,
    account: &str,
    denom: &str,
) -> StdResult<Coin> {
    let whitelisted_balance: WhitelistedBalanceResponse = deps.querier.query(
        &CoreumQueries::AssetFT(Query::WhitelistedBalance {
            account: account.to_string(),
            denom: denom.to_string(),
        })
        .into(),
    )?;
    Ok(whitelisted_balance.balance)
}

fn query_bank_balance(deps: Deps<CoreumQueries>, account: &str, denom: &str) -> StdResult<Coin> {
    let bank_balance: BalanceResponse = deps.querier.query(
        &BankQuery::Balance {
            address: account.to_string(),
            denom: denom.to_string(),
        }
        .into(),
    )?;

    Ok(bank_balance.amount)
}

fn query_token(deps: Deps<CoreumQueries>, denom: &str) -> StdResult<Token> {
    let token: TokenResponse = deps.querier.query(
        &CoreumQueries::AssetFT(Query::Token {
            denom: denom.to_string(),
        })
        .into(),
    )?;

    Ok(token.token)
}
