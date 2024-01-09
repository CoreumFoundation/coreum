use cosmwasm_std::StdError;
use cw_utils::PaymentError;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("Payment error: {0}")]
    Payment(#[from] PaymentError),

    #[error("Need to send exactly the NFT price")]
    InvalidFundsAmount {},
}
