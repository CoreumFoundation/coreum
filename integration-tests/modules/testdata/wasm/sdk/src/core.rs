use crate::{assetft, assetnft, nft};
use cosmwasm_std::{CosmosMsg, CustomMsg, CustomQuery};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum CoreumMsg {
    AssetFT(assetft::Msg),
    AssetNFT(assetnft::Msg),
    NFT(nft::Msg),
}

impl From<CoreumMsg> for CosmosMsg<CoreumMsg> {
    fn from(val: CoreumMsg) -> Self {
        CosmosMsg::Custom(val)
    }
}

impl CustomMsg for CoreumMsg {}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum CoreumQueries {
    AssetFT(assetft::Query),
    AssetNFT(assetnft::Query),
    NFT(nft::Query),
}

impl CustomQuery for CoreumQueries {}
