use cosmwasm_std::{entry_point, StdError};
use cosmwasm_std::{BalanceResponse, BankQuery};
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
    _env: Env,
    info: MessageInfo,
    amount: Uint128,
    recipient: String,
) -> CoreumResult<ContractError> {
    // TODO(milad) check that amount is present in the attached funds, and attached funds
    // is enough to cover the transfer.
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

    if let Some(features) = &token.features {
        // TODO(masih):
        // - The user cannot burn the frozen amount if both freezing and burning is enabled.
        // - If either or both of BurnRate and SendCommissionRate are set above zero, then
        // after transfer has taken place and those rates are applied, the sender's balance
        // must not go below the frozen amount. Otherwise the transaction will fail.

        if features.contains(&assetft::FREEZING) {
            assert_freezing(deps.as_ref(), info.sender.as_ref(), &token)?;
        }

        if features.contains(&assetft::WHITELISTING) {
            assert_whitelisting(deps.as_ref(), &recipient, &token, amount)?;
        }

        // TODO remove this if statement.
        // This check is intended for POC testing, it must be replaced with a more
        // meaningful check.
        if amount == Uint128::new(101) {
            let burn_message = CoreumMsg::AssetFT(assetft::Msg::Burn {
                coin: cosmwasm_std::coin(amount.u128(), denom)
            });

            // TODO(masih): Change token.issuer to token.admin
            if !features.contains(&assetft::BURNING) && token.issuer != info.sender.as_ref().to_string() {
                return Err(ContractError::FeatureDisabledError {
                    issuer: token.issuer.to_string(),
                    sender: info.sender.as_ref().to_string()
                });
            }

            return Ok(Response::new()
                .add_attribute("method", "burn")
                .add_message(burn_message));
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
