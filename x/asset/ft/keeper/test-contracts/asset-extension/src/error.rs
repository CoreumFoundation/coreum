use cosmwasm_std::StdError;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("Requested transfer token is frozen.")]
    FreezingError {},

    #[error("Whitelisted limit exceeded.")]
    WhitelistingError {},

    #[error("Feature disabled. {issuer} {sender}")]
    FeatureDisabledError {
        issuer: String,
        sender: String
    },
}
