use cosmwasm_schema::QueryResponses;
use cosmwasm_std::Binary;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use crate::pagination::{PageRequest, PageResponse};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct NFT {
    pub class_id: String,
    pub id: String,
    pub uri: Option<String>,
    pub uri_hash: Option<String>,
    pub data: Option<Binary>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct Class {
    pub id: String,
    pub name: Option<String>,
    pub symbol: Option<String>,
    pub description: Option<String>,
    pub uri: Option<String>,
    pub uri_hash: Option<String>,
    pub data: Option<Binary>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct BalanceResponse {
    pub amount: u64,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct OwnerResponse {
    pub owner: String,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct SupplyResponse {
    pub amount: u64,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct NFTResponse {
    pub nft: NFT,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct NFTsResponse {
    pub nfts: Vec<NFT>,
    pub pagination: PageResponse,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct ClassResponse {
    pub class: Class,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct ClassesResponse {
    pub classes: Vec<Class>,
    pub pagination: PageResponse,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum Msg {
    Send {
        class_id: String,
        id: String,
        receiver: String,
    },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema, QueryResponses)]
pub enum Query {
    #[returns(BalanceResponse)]
    Balance { class_id: String, owner: String },

    #[returns(OwnerResponse)]
    Owner { class_id: String, id: String },

    #[returns(SupplyResponse)]
    Supply { class_id: String },

    #[returns(NFTResponse)]
    NFT { class_id: String, id: String },

    #[returns(NFTsResponse)]
    NFTs {
        class_id: Option<String>,
        owner: Option<String>,
        pagination: Option<PageRequest>,
    },

    #[returns(ClassResponse)]
    Class { class_id: String },

    #[returns(ClassesResponse)]
    Classes { pagination: Option<PageRequest> },
}
