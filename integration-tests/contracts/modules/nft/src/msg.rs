use cosmwasm_schema::cw_serde;
use cosmwasm_std::Binary;

#[cw_serde]
pub struct InstantiateMsg {
    pub name: String,
    pub symbol: String,
    pub description: Option<String>,
    pub uri: Option<String>,
    pub uri_hash: Option<String>,
    pub data: Option<Binary>,
    pub features: Option<Vec<i32>>,
    pub royalty_rate: Option<String>,
}

#[cw_serde]
pub enum ExecuteMsg {
    MintMutable {
        id: String,
        uri: Option<String>,
        uri_hash: Option<String>,
        data: Option<Binary>,
        recipient: Option<String>,
    },
    MintImmutable {
        id: String,
        uri: Option<String>,
        uri_hash: Option<String>,
        data: Option<Binary>,
        recipient: Option<String>,
    },
    ModifyData {
        id: String,
        data: Binary,
    },
    Burn {
        id: String,
    },
    Freeze {
        id: String,
    },
    Unfreeze {
        id: String,
    },
    ClassFreeze {
        account: String,
    },
    ClassUnfreeze {
        account: String,
    },
    AddToWhitelist {
        id: String,
        account: String,
    },
    RemoveFromWhitelist {
        id: String,
        account: String,
    },
    AddToClassWhitelist {
        account: String,
    },
    RemoveFromClassWhitelist {
        account: String,
    },
    Send {
        id: String,
        receiver: String,
    },
}

#[cw_serde]
pub enum QueryMsg {}
