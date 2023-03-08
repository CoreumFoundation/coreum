use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct NFT {
    pub class_id: String,
    pub id: String,
    pub uri: Option<String>,
    pub uri_hash: Option<String>,
    pub data: Option<String>,
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
pub enum Msg {
    Send {
        class_id: String,
        id: String,
        receiver: String,
    },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum Query {
    Balance { class_id: String, owner: String },
    Owner { class_id: String, id: String },
    Supply { class_id: String },
    NFT { class_id: String, id: String },
}
