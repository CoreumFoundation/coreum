# NFT Contract

This contract showcases how to interact with the AssetNFT module using all messages and queries available.

# Instantiation

The contract can be instantiated with the following messages

```
{
    "name": "<NEW_ASSETNFT_NAME>",
    "symbol": "<NEW_ASSETNFT_SYMBOL>",
    "description": "<DESCRIPTION_INFO>",
    "uri": "<NFT_URI>",
    "uri_hash": "<NFT_URI_HASH>",
    "data": "<NFT_DATA_BINARY>",
    "features": "[<FEATURE_1_ID>, <FEATURE_2_ID> ...]",
    "royalty_rate": "<ROYALTY_RATE>",
}
```

The instantiantion of the contract will issue a new AssetNFT (and therefore become the issuer of the asset) with the values provided. The class_id of the new AssetNFT will be generated as {symbol}-{issuer_address}.

Features define what actions can be performed on the new non fungible token (These features are immutable in the future). Available features: Burning (0), Freezing(1), Whitelisting(2), Disable Sending(3).
Royalty rate is a number between 0 and 1 (in String format). Whenever an NFT of this class is traded for a certain amount, the royalty fee will be multiplied by this amount and sent to the issuer of the NFT

For more detailed information of the AssetNFT module and functionality go to [AssetNFT](https://github.com/CoreumFoundation/coreum/tree/master/x/asset/nft/spec)

# Messages

### Mint (id, uri, uri_hash, data)

The contract (issuer) will mint an NFT with the information provided (only id is mandatory).

### Burn (id) [Burning](https://github.com/CoreumFoundation/coreum/tree/master/x/asset/nft/spec#burning)

Burns the NFT with a certain id.

### Freeze (id) [Freezing](https://github.com/CoreumFoundation/coreum/tree/master/x/asset/nft/spec#freezing)

Freezes the NFT with that id.

### Unfreeze (id) 

Unfreezes a frozen NFT.

### AddToWhitelist (id, account) [Whitelisting](https://github.com/CoreumFoundation/coreum/tree/master/x/asset/nft/spec#whitelisting)

Whitelists an address so that it can receive the NFT with that id.

### RemoveFromWhitelist (id, account)

Removes an address from the whitelist so that it can not receive the NFT with that id.

### Send (id, receiver)

Sends an NFT to the address provided.

# Queries (AssetNFT)

### Params

Returns the parameters of the AssetNFT module (the mint fee of an NFT, which is burnt every time an NFT is minted).

### Class

Returns all the information of the NFT Class (the one we provided during instantiation)

### Classes (issuer)

Returns the information of all NFT classes issued by an address.

### Frozen (id)

Checks if an NFT is frozen or not.

### Whitelisted (id, account)

Checks if an account is whitelisted for a particular NFT.

### WhitelistedAccountsForNFT (id)

Returns all accounts that are whitelisted for an NFT

# Queries (NFT)

These queries have nothing to do with the AssetNFT module. They are queries used to test an old NFT module that will eventually be deprecated.

### Balance (owner)

Queries the number of NFTs of a given class owned by the owner, same as balanceOf in ERC721

### Owner (id)

Queries the owner of an NFT.

### Supply

Queries the number of NFTs.

### Nft (id)

Queries the NFT of a given id.

### Nfts (Optional: Owner)

Queries all the NFTs by a ClassID or by an Owner if provided.

### Class

Queries a Class with a given ClassID

### Classes

Queries all the Classes
