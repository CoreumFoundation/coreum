use cosmwasm_std::StdError;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("Unauthorized.")]
    Unauthorized {},

    #[error("Transferring to or from smart contracts are prohibited.")]
    SmartContractBlocked {},

    #[error("Insufficient funds attached.")]
    InsufficientFunds {},

    #[error("IBC feature is disabled.")]
    IBCDisabled {},

    #[error("Invalid amount.")]
    InvalidAmountError {},

    #[error("DEX order placement is failed.")]
    DEXOrderPlacementError {},
}
