use cosmwasm_schema::cw_serde;

#[cw_serde]
pub struct InstantiateMsg {
    pub count: i32,
}

#[cw_serde]
pub enum ExecuteMsg {
    Increment {},
}

#[cw_serde]
pub enum QueryMsg {
    // GetCount returns the current count as a json-encoded number
    GetCount {},
}

// We define a custom struct for each query response
#[cw_serde]
pub struct CountResponse {
    pub count: i32,
}

#[cw_serde]
pub struct MigrateMsg {
    pub count: i32,
}
