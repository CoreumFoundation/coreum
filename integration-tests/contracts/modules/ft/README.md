# FT Contract

This contract showcases how to interact with the AssetFT module using all messages and queries available.

# Instantiation

The contract can be instantiated with the following messages

```
{
    "symbol": "<NEW_ASSETFT_SYMBOL>",
    "subunit": "<NEW_ASSETFT_SUBUNIT>",
    "precision: <NEW_ASSETFT_PRECISION>",
    "description": "<DESCRIPTION_INFO>",
    "features": "[<FEATURE_1_ID>, <FEATURE_2_ID> ...]",
    "burn_rate": "<BURN_RATE>",
    "send_commission_rate": "<SEND_COMMISSION_RATE>"
}
```

The instantiantion of the contract will issue a new AssetFT (and therefore become the issuer of the asset) with the values provided. The denom of the new AssetFT will be generated as {subunit}-{issuer_address}.

Features define what actions can be performed on the new fungible token (These features are immutable in the future). Available features: Minting (0), Burning(1), Freezing(2), Whitelisting(3).
Burn rate and send commission rate are numbers between 0 and 1 (in String format) which will be multiplied by send amount to determine how much is going to be burnt/sent to the token issuer on top of the send amount.

For more detailed information of the AssetFT module and functionality go to [AssetFT](https://github.com/CoreumFoundation/coreum/tree/master/x/asset/ft/spec)

# Messages

### Mint (amount) [Mint](https://github.com/CoreumFoundation/coreum/tree/master/x/asset/ft/spec#mint)

The contract (issuer) will mint the amount of tokens provided.

### Burn (amount) [Burn](https://github.com/CoreumFoundation/coreum/tree/master/x/asset/ft/spec#burn)

Burns the amount of tokens provided.

### Freeze (account, amount) [Freeze](https://github.com/CoreumFoundation/coreum/tree/master/x/asset/ft/spec#freezeunfreeze)

Freezes the amount of tokens of an account.

### Unfreeze (account, amount) [Unfreeze](https://github.com/CoreumFoundation/coreum/tree/master/x/asset/ft/spec#freezeunfreeze)

Unfreezes the amount of tokens of an account.

### GloballyFreeze [Global-freeze](https://github.com/CoreumFoundation/coreum/tree/master/x/asset/ft/spec#global-freezeunfreeze)

Globally freezes the token.

### GloballyUnfreeze [Global-unfreeze](https://github.com/CoreumFoundation/coreum/tree/master/x/asset/ft/spec#global-freezeunfreeze)

Globally unfreezes the token.

### SetWhitelistedLimit (account, amount) [Whitelist](https://github.com/CoreumFoundation/coreum/tree/master/x/asset/ft/spec#whitelist)

Sets a whitelisted limit for an account.

### MintAndSend (account, amount)

Combines the Mint feature described above with a bank transfer for convenience.

# Queries

### Params

Returns the parameters of the AssetFT module (the issue fee of a token).

### Token

Returns all available information of the Token we issued during instantiation.

### Tokens (issuer)

Returns all tokens issued by a specific address.

### Balance (account)

Returns the balance of the token that was issued during instantiation for an account.

### FrozenBalance(account)

Returns the frozen balance of the token issued for an account.

### FrozenBalances (account)

Returns all frozen balances (of all tokens) for an account.

### WhitelistedBalance(account)

Returns the whitelisted balance of the token issued for an account.

### WhitelistedBalances (account)

Returns all whitelisted balances (of all tokens) for an account.
