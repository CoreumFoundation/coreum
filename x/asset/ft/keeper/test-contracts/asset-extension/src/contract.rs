use cosmwasm_std::{entry_point, StdError};
use cosmwasm_std::{BalanceResponse, BankQuery, ContractInfoResponse, WasmQuery};
use cosmwasm_std::{Binary, Coin, Deps, DepsMut, Env, MessageInfo, Response, StdResult, Uint128};
use cw2::set_contract_version;

use crate::error::ContractError;
use coreum_wasm_sdk::assetft::{
    self, FrozenBalanceResponse, Query, Token, TokenResponse, WhitelistedBalanceResponse,
};
use coreum_wasm_sdk::core::{CoreumMsg, CoreumQueries, CoreumResult};

use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg, SudoMsg};
use crate::state::DENOM;

// version info for migration info
const CONTRACT_NAME: &str = env!("CARGO_PKG_NAME");
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

const AMOUNT_DISALLOWED_TRIGGER: Uint128 = Uint128::new(7);
const AMOUNT_IGNORE_WHITELISTING_TRIGGER: Uint128 = Uint128::new(49);
const AMOUNT_IGNORE_FREEZING_TRIGGER: Uint128 = Uint128::new(79);
const AMOUNT_BURNING_TRIGGER: Uint128 = Uint128::new(101);
const AMOUNT_MINTING_TRIGGER: Uint128 = Uint128::new(105);
const AMOUNT_IGNORE_BURN_RATE_TRIGGER: Uint128 = Uint128::new(108);

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> CoreumResult<ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    DENOM.save(deps.storage, &msg.denom)?;

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
            commission_amount: _,
            burn_amount,
            context: _,
        } => sudo_extension_transfer(deps, env, transfer_amount, sender, recipient, burn_amount),
    }
}

pub fn sudo_extension_transfer(
    deps: DepsMut<CoreumQueries>,
    _env: Env,
    amount: Uint128,
    sender: String,
    recipient: String,
    burn_amount: Uint128,
) -> CoreumResult<ContractError> {
    if amount == AMOUNT_DISALLOWED_TRIGGER {
        return Err(ContractError::Std(StdError::generic_err(
            "7 is not allowed",
        )));
    }

    let denom = DENOM.load(deps.storage)?;

    let token = query_token(deps.as_ref(), &denom)?;

    if let Some(features) = &token.features {
        // TODO(masih): If either or both of BurnRate and SendCommissionRate are set above zero,
        // then after transfer has taken place and those rates are applied, the sender's balance
        // must not go below the frozen amount. Otherwise the transaction will fail.

        if features.contains(&assetft::FREEZING) {
            assert_freezing(deps.as_ref(), sender.as_ref(), &token, amount)?;
        }

        if features.contains(&assetft::WHITELISTING) {
            assert_whitelisting(deps.as_ref(), &recipient, &token, amount)?;
        }

        if features.contains(&assetft::BLOCK_SMART_CONTRACTS) {
            assert_block_smart_contracts(deps.as_ref(), &recipient, &token)?;
        }

        // TODO remove this if statement.
        // This check is intended for POC testing, it must be replaced with a more
        // meaningful check.
        if amount == AMOUNT_BURNING_TRIGGER {
            return assert_burning(amount, &token);
        }

        // TODO remove this if statement.
        // This check is intended for POC testing, it must be replaced with a more
        // meaningful check.
        if amount == AMOUNT_MINTING_TRIGGER {
            return assert_minting(sender.as_ref(), &recipient, amount, &token);
        }
    }

    let transfer_msg = cosmwasm_std::BankMsg::Send {
        to_address: recipient.to_string(),
        amount: vec![Coin { amount, denom }],
    };

    let mut response = Response::new()
        .add_attribute("method", "execute_transfer")
        .add_message(transfer_msg);

    if !burn_amount.is_zero() {
        response = assert_burn_rate(response, sender.as_ref(), amount, &token, burn_amount)?;
    }

    Ok(response)
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(_deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {}
}

fn assert_freezing(
    deps: Deps<CoreumQueries>,
    account: &str,
    token: &Token,
    amount: Uint128,
) -> Result<(), ContractError> {
    // Allow any amount if recipient is admin
    if token.admin == Some(account.to_string()) {
        return Ok(());
    }

    // TODO remove this if statement.
    // This check is intended for POC testing, it must be replaced with a more
    // meaningful check.
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
    deps: Deps<CoreumQueries>,
    account: &str,
    token: &Token,
    amount: Uint128,
) -> Result<(), ContractError> {
    // Allow any amount if recipient is admin
    if token.admin == Some(account.to_string()) {
        return Ok(());
    }

    // TODO remove this if statement.
    // This check is intended for POC testing, it must be replaced with a more
    // meaningful check.
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
    deps: Deps<CoreumQueries>,
    recipient: &str,
    token: &Token,
) -> Result<(), ContractError> {
    if recipient.to_string() == token.issuer
        || Some(recipient.to_string()) == token.extension_cw_address
    {
        return Ok(());
    }

    if is_smart_contract(deps, recipient) {
        return Err(ContractError::SmartContractBlocked {});
    }

    return Ok(());
}

fn assert_burn_rate(
    mut response: Response<CoreumMsg>,
    sender: &str,
    amount: Uint128,
    token: &Token,
    mut burn_amount: Uint128,
) -> CoreumResult<ContractError> {
    if amount == AMOUNT_IGNORE_BURN_RATE_TRIGGER {
        let refund_burn_rate_msg = cosmwasm_std::BankMsg::Send {
            to_address: sender.to_string(),
            amount: vec![Coin {
                amount: burn_amount,
                denom: token.denom.to_string(),
            }],
        };

        burn_amount = Uint128::zero();

        response = response
            .add_attribute("burn_rate_refund", burn_amount.to_string())
            .add_message(refund_burn_rate_msg);
    }

    if !burn_amount.is_zero() {
        let burn_message = CoreumMsg::AssetFT(assetft::Msg::Burn {
            coin: cosmwasm_std::coin(burn_amount.u128(), &token.denom),
        });

        response = response
            .add_attribute("burn_amount", burn_amount)
            .add_message(burn_message);
    }

    Ok(response)
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

fn query_contract_info(
    deps: Deps<CoreumQueries>,
    account: &str,
) -> StdResult<ContractInfoResponse> {
    let contract_info: ContractInfoResponse = deps.querier.query(
        &WasmQuery::ContractInfo {
            contract_addr: account.to_string(),
        }
        .into(),
    )?;

    Ok(contract_info)
}

fn is_smart_contract(deps: Deps<CoreumQueries>, account: &str) -> bool {
    query_contract_info(deps, account).is_ok()
}
