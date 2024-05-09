use cosmwasm_std::{entry_point, StdError};
use cosmwasm_std::{BalanceResponse, BankQuery, WasmQuery, ContractInfoResponse};
use cosmwasm_std::{Binary, Coin, Deps, DepsMut, Env, MessageInfo, Response, StdResult, Uint128};
use cw2::set_contract_version;

use crate::error::ContractError;
use coreum_wasm_sdk::assetft::{FrozenBalanceResponse, Query, Token, TokenResponse, WhitelistedBalanceResponse, self};
use coreum_wasm_sdk::core::{CoreumMsg, CoreumQueries, CoreumResult};

use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::DENOM;

// version info for migration info
const CONTRACT_NAME: &str = env!("CARGO_PKG_NAME");
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

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
    deps: DepsMut<CoreumQueries>,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> CoreumResult<ContractError> {
    match msg {
        ExecuteMsg::ExtensionTransfer { amount, recipient } => {
            execute_extension_transfer(deps, env, info, amount, recipient)
        }
    }
}

pub fn execute_extension_transfer(
    deps: DepsMut<CoreumQueries>,
    env: Env,
    info: MessageInfo,
    amount: Uint128,
    recipient: String,
) -> CoreumResult<ContractError> {
    // TODO remove this if statement.
    // This check is intended for POC testing, it must be replaced with a more
    // meaningful check.
    if amount == Uint128::new(7) {
        return Err(ContractError::Std(StdError::generic_err(
            "7 is not allowed",
        )));
    }

    let denom = DENOM.load(deps.storage)?;

    let token = query_token(deps.as_ref(), &denom)?;

    // check that amount is present in the attached funds, and attached funds is enough
    // to cover the transfer.
    let has_sufficient_funds = info.funds.iter().any(
        |coin| coin.denom == denom && coin.amount >= amount
    );
    if !has_sufficient_funds {
        return Err(ContractError::InsufficientFunds {})
    }

    if let Some(features) = &token.features {
        // TODO(masih): If either or both of BurnRate and SendCommissionRate are set above zero,
        // then after transfer has taken place and those rates are applied, the sender's balance
        // must not go below the frozen amount. Otherwise the transaction will fail.

        if features.contains(&assetft::FREEZING) {
            assert_freezing(deps.as_ref(), info.sender.as_ref(), &token)?;
        }

        if features.contains(&assetft::WHITELISTING) {
            assert_whitelisting(deps.as_ref(), &recipient, &token, amount)?;
        }

        if features.contains(&assetft::BLOCK_SMART_CONTRACTS) {
            assert_block_smart_contracts(
                deps.as_ref(), info.sender.as_ref(), &recipient, &token
            )?;
        }

        // TODO remove this if statement.
        // This check is intended for POC testing, it must be replaced with a more
        // meaningful check.
        if amount == Uint128::new(101) {
            return assert_burning(
                info.sender.as_ref(), amount, &token, features.contains(&assetft::BURNING)
            );
        }

        // TODO remove this if statement.
        // This check is intended for POC testing, it must be replaced with a more
        // meaningful check.
        if amount == Uint128::new(105) {
            return assert_minting(
                info.sender.as_ref(), &recipient, amount, &token, features.contains(&assetft::MINTING)
            );
        }
    }

    let transfer_msg = cosmwasm_std::BankMsg::Send {
        to_address: recipient,
        amount: vec![Coin {amount, denom}],
    };

    Ok(Response::new()
        .add_attribute("method", "execute_transfer")
        .add_message(transfer_msg))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(_deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {}
}

fn assert_freezing(
    deps: Deps<CoreumQueries>,
    account: &str,
    token: &Token,
) -> Result<(), ContractError> {
    // Allow any amount if recipient is admin
    // TODO(masih): Change it to admin
    if token.issuer == account.to_string() {
        return Ok(());
    }

    // TODO(masih): Uncomment after updating the SDK
    // if token.globally_frozen {
    //     return Err(ContractError::FreezingError {});
    // }

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
    // TODO(masih): Change it to admin
    if token.issuer == account.to_string() {
        return Ok(());
    }

    let bank_balance = query_bank_balance(deps, account, &token.denom)?;
    let whitelisted_balance = query_whitelisted_balance(deps, account, &token.denom)?;

    if amount + bank_balance.amount > whitelisted_balance.amount {
        return Err(ContractError::WhitelistingError {});
    }

    Ok(())
}

fn assert_burning(
    sender: &str,
    amount: Uint128,
    token: &Token,
    burning_enabled: bool
) -> CoreumResult<ContractError> {
    let burn_message = CoreumMsg::AssetFT(assetft::Msg::Burn {
        coin: cosmwasm_std::coin(amount.u128(), &token.denom)
    });

    // TODO(masih): Change token.issuer to token.admin
    if !burning_enabled && sender != token.issuer {
        return Err(ContractError::FeatureDisabledError {});
    }

    return Ok(Response::new()
        .add_attribute("method", "burn")
        .add_message(burn_message));
}

fn assert_minting(
    sender: &str,
    recipient: &str,
    amount: Uint128,
    token: &Token,
    minting_enabled: bool
) -> CoreumResult<ContractError> {
    let mint_message = CoreumMsg::AssetFT(assetft::Msg::Mint {
        coin: cosmwasm_std::coin(amount.u128(), &token.denom),
        recipient: Some(recipient.to_string()),
    });

    if !minting_enabled {
        return Err(ContractError::FeatureDisabledError {});
    }

    // TODO(masih): Change token.issuer to token.admin
    if sender != token.issuer {
        return Err(ContractError::Unauthorized {});
    }

    let return_fund_msg = cosmwasm_std::BankMsg::Send {
        to_address: sender.to_string(),
        amount: vec![Coin {amount, denom: token.denom.clone()}],
    };

    return Ok(Response::new()
        .add_attribute("method", "mint")
        .add_message(mint_message)
        .add_message(return_fund_msg));
}

fn assert_block_smart_contracts(
    deps: Deps<CoreumQueries>,
    sender: &str,
    recipient: &str,
    token: &Token,
) -> Result<(), ContractError> {
    // TODO: Do we need this?
    let issued_from_smart_contract = is_smart_contract(deps, &token.issuer);
    if issued_from_smart_contract &&
        (sender.to_string() == token.issuer || recipient.to_string() == token.issuer) {
        return Ok(())
    }

    if is_smart_contract(deps, sender) {
        return Err(ContractError::SmartContractBlocked {});
    }

    if is_smart_contract(deps, recipient) {
        return Err(ContractError::SmartContractBlocked {});
    }

    return Ok(());
}

fn query_frozen_balance(
    deps: Deps<CoreumQueries>,
    account: &str,
    denom: &str,
) -> StdResult<Coin> {
    let frozen_balance: FrozenBalanceResponse = deps.querier.query(
        &CoreumQueries::AssetFT(
            Query::FrozenBalance {
                account: account.to_string(),
                denom: denom.to_string(),
            }
        ).into()
    )?;
    Ok(frozen_balance.balance)
}

fn query_whitelisted_balance(
    deps: Deps<CoreumQueries>,
    account: &str,
    denom: &str,
) -> StdResult<Coin> {
    let whitelisted_balance: WhitelistedBalanceResponse = deps.querier.query(
        &CoreumQueries::AssetFT(
            Query::WhitelistedBalance {
                account: account.to_string(),
                denom: denom.to_string(),
            }
        ).into()
    )?;
    Ok(whitelisted_balance.balance)
}

fn query_bank_balance(
    deps: Deps<CoreumQueries>,
    account: &str,
    denom: &str,
) -> StdResult<Coin> {
    let bank_balance: BalanceResponse = deps.querier.query(
        &BankQuery::Balance {
            address: account.to_string(),
            denom: denom.to_string(),
        }
            .into(),
    )?;

    Ok(bank_balance.amount)
}

fn query_token(
    deps: Deps<CoreumQueries>,
    denom: &str,
) -> StdResult<Token> {
    let token: TokenResponse = deps.querier.query(
        &CoreumQueries::AssetFT(
            Query::Token { denom: denom.to_string() }
        )
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

fn is_smart_contract(
    deps: Deps<CoreumQueries>,
    account: &str,
) -> bool {
    query_contract_info(deps, account).is_ok()
}
