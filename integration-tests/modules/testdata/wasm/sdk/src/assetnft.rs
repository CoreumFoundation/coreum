use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct Class {
    pub id: String,
    pub issuer: String,
    pub name: String,
    pub symbol: String,
    pub description: Option<String>,
    pub uri: Option<String>,
    pub uri_hash: Option<String>,
    pub data: Option<String>,
    pub features: Option<Vec<u32>>,
    pub royalty_rate: Option<String>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct ClassResponse {
    pub class: Class,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct FrozenResponse {
    pub frozen: bool,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct WhitelistedResponse {
    pub whitelisted: bool,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum Msg {
    IssueClass {
        name: String,
        symbol: String,
        description: Option<String>,
        uri: Option<String>,
        uri_hash: Option<String>,
        data: Option<String>,
        features: Option<Vec<u32>>,
        royalty_rate: Option<String>,
    },
    Mint {
        class_id: String,
        id: String,
        uri: Option<String>,
        uri_hash: Option<String>,
        data: Option<String>,
    },
    Burn {
        class_id: String,
        id: String,
    },
    Freeze {
        class_id: String,
        id: String,
    },
    Unfreeze {
        class_id: String,
        id: String,
    },
    AddToWhitelist {
        class_id: String,
        id: String,
        account: String,
    },
    RemoveFromWhitelist {
        class_id: String,
        id: String,
        account: String,
    },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum Query {
    Class {
        id: String,
    },
    Frozen {
        id: String,
        class_id: String,
    },
    Whitelisted {
        id: String,
        class_id: String,
        account: String,
    },
}
