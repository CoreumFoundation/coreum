use cosmwasm_std::StdError;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("Requested transfer token is frozen.")]
    FreezingError {},

    #[error("Whitelisted limit exceeded. {amount} {bank_balance} {whitelist_balance}")]
    WhitelistingError {
        amount: String,
        bank_balance: String,
        whitelist_balance: String
    },
}
