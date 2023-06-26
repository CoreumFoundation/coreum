use schemars::JsonSchema;
use serde::{Serialize, Deserialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[derive(Default)]
#[serde(rename_all = "snake_case")]
pub struct PageRequest {
    pub key: Option<String>,
    pub offset: Option<u64>,
    pub limit: Option<u64>,
    pub count_total: Option<bool>,
    pub reverse: Option<bool>,
}

impl PageRequest {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn key(mut self, key: impl Into<String>) -> Self {
        self.key = Some(key.into());
        self
    }

    pub fn offset(mut self, offset: u64) -> Self {
        self.offset = Some(offset);
        self
    }

    pub fn limit(mut self, limit: u64) -> Self {
        self.limit = Some(limit);
        self
    }

    pub fn count_total(mut self) -> Self {
        self.count_total = Some(true);
        self
    }

    pub fn reverse(mut self) -> Self {
        self.reverse = Some(true);
        self
    }
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub struct PageResponse {
    pub next_key: Option<String>,
    pub total: Option<u64>,
}
