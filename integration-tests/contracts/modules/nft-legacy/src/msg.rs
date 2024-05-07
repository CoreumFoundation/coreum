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
    pub features: Option<Vec<u32>>,
    pub royalty_rate: Option<String>,
}

#[cw_serde]
pub enum ExecuteMsg {
    Mint{
        id: String,
        uri: Option<String>,
        uri_hash: Option<String>,
        data: Option<Binary>,
        recipient: Option<String>,
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
pub enum QueryMsg {
    Params {},
    Class {},
    Classes { issuer: String },
    Frozen { id: String },
    ClassFrozen { account: String },
    ClassFrozenAccounts {},
    Whitelisted { id: String, account: String },
    WhitelistedAccountsForNft { id: String },
    ClassWhitelistedAccounts {},
    Balance { owner: String },
    Owner { id: String },
    Supply {},
    Nft { id: String }, // we use Nft not NFT since NFT is decoded as n_f_t
    Nfts { owner: Option<String> }, // we use Nfts not NFTs since NFTs is decoded as n_f_ts
    ClassNft {}, // we use ClassNft instead of Class because there is already a Class query being used
    ClassesNft {}, // we use ClassesNft instead of Class because there is already a Classes query being used
    BurntNft { nft_id: String },
    BurntNftsInClass {},
    // Check that we can query NFTs that were not issued with the handler and thus might have DataDynamic
    ExternalNft { class_id: String, id: String },
}
