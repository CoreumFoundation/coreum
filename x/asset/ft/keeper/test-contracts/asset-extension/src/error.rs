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

    #[error("Feature disabled.")]
    FeatureDisabledError {},

    #[error("Unauthorized.")]
    Unauthorized {},

    #[error("Transferring to or from smart contracts are prohibited.")]
    SmartContractBlocked {},

    #[error("Insufficient funds attached.")]
    InsufficientFunds {},

    // TODO: Delete this one
    #[error("Debugging {a} {b} {c} {d} {e} {f}")]
    Debugging {
        a: String,
        b: String,
        c: String,
        d: String,
        e: String,
        f: String
    },
}
