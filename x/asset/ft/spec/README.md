# Abstract
This document specifies `assetft` module, which allows public users of the blockchain to create fungible tokens on Coreum blockchain.

# Concepts
In this section we will provide the detailed behavior of fungible token creation and management.

Here is the list of functionalities provided by this module, we will examine each of them separately. 
 - Issue
 - Mint
 - Burn
 - Freeze
 - Global Freeze
 - Whitelist

## Interaction with bank module, introducing wbank module
Since Coreum is based on Cosmos SDK, We should mention that Cosmos SDK provides the native bank module which is responsible for tracking fungible token creation and balances of each account. But this module does not allow any public to create a fungible token, mint/burn it, and also does not allow for other features such as freezing and whitelisting. To work around this issue we have wrapped the `bank` module into the `wbank` module. 

In `wbank` module we wrap all the send related  methods of the `bank` module and intercept them with `BeforeSend` and `BeforeInputOutput` functions provided by `assetft` module. This allows `assetft` module to inject custom logic into interceptor functions and reject some transaction if whitelisting or freezing criteria are not met, or apply other features such as BurnRate. 

This structure allows to reuse the code provided by Cosmos SDK, and also reuse the infrastructure that the community provides (e.g explorers and wallets). But it also leads to the fact that some of the information regarding fungible tokens will exist in the `assetft` module and some in the `bank` module. For example, if you want to query for frozen balances of a fungible token, you need to query the `assetft` module but if you want to get the total supply, you must query the bank module.

In a nutshell, `assetft` module interacts with `wbank` which in turn wraps the original `bank` module.

## Token Interactions
### Issue
Coreum provides a decentralized platform which allows everyone to tokenize their assets. Although the functionality of fungible token creation and minting is present in the original `bank` module of Cosmos SDK, it is not exposed to end users, and it is only possible to create new fungible tokens via either the governance or IBC. The Issue method described here, makes it possible for everyone to create a fungible token and manage its supply. When the issuer issues a token, they specify the initial total supply which will be delivered to the issuer's account address.

All the information provided at the time of issuance is immutable and cannot be changed later.

#### Denom naming, Symbol and Precision
The way that denom is created is that the user provides a name for their subunit, and the denom for the token, which is the main identifier of the token, will be created by joining the subunit and the issuer address separated with a dash (subunit-address). The user also provides the symbol and precision which will only be used for display purposes and will be stored in bank module's metadata field.

For example to represent Bitcoin on Coreum, one could choose satoshi as subunit, BTC as Symbol and 8 as precision. It means that if the issuer address is core1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8 then the denom will be `satoshi-core1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8` and since we have chosen BTC as symbol and 8 as precision, it will follow that (1 BTC = 10^8 `satoshi-devcore1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8`)

#### Token Features
When issuing a token, the issuer must decide which features are enabled on the token. For example if `minting` feature is enabled then it will allow the issuer to mint further tokens on top of the initial supply, otherwise no new tokens can be minted. Here is a list of all the features that can be enabled on the token. Each of these features affects how the token will behave and a detailed description will be provided in the dedicated section for each feature.

- minting
- burning
- freezing
- whitelisting

#### Burn Rate 
The issuer has the option to provide `BurnRate` when issuing a new token. This value is a number between 0 and 1, and if it is above zero, in every transfer, some additional tokens will be burnt on top of the transferred value, from the senders address. The tokens to be burnt are calculated by multiplying the TransferAmount by burn rate, and rounding it up to an integer value.

#### Send Commission Rate
Exactly same as the Burn Rate, but the calculated value will be transferred to the issuer's account addressed instead of being burnt.

#### Issuance Fee
Whenever a user wants to issue a fungible token, they have to pay some extra money as issuance fee, which is calculated on top of tx execution fee and will be burnt. The amount of the issuance fee is controlled by governance.

### Mint
If the minting feature is enabled, then issuer of the token can submit a Mint transaction to add more tokens to the total supply. All the minted tokens will be transferred to the issuer's account address.

### Burn
The issuer of the token can burn the tokens that they hold. If the burning feature is enabled, then every holder of the token can burn the tokens they hold.

### Freeze/Unfreeze
If the freezing feature is enabled on a token, then the issuer of the token can freeze an account up to an amount. The frozen amount can be more than what the user currently holds, an works even if the user holds zero. The user can only send the tokens that they hold in excess of the frozen amount. 
For example if the issuer freezes 1000 ABC tokens on account Y and this account holds 800, then they cannot move any of their tokens, but if the account receives 400 extra ABC tokens, their total balance will become 1200 and then can only spend 200 of it, since 1000 is Frozen.

Here is the description of behavior of the freezing feature:
- The issuer can freeze an account up to amount if the freezing feature is enabled.
- The issuer can increase the frozen amount by submitting new freeze transaction on top of already frozen account. The frozen amount of the account will be increased by the new value.
- The issuer can unfreeze a portion of the frozen amount on an account.
- The issuer cannot freeze their own account
- The user can only send their tokens in excess of the frozen amount.
- The user can receive tokens regardless of frozen limitation.
- The user cannot burn the frozen amount if both freezing and burning is enabled.
- Issuer cannot freeze their own account
- Frozen amount cannot be a negative value, it means that amount present in unfreeze transaction cannot be bigger than the current frozen amount
- If either or both of BurnRate and SendCommissionRate are set above zero, then after transfer has taken place and those rates are applied, the sender's balance must not go below the frozen amount. Otherwise the transaction will fail.

### Global Freeze/Unfreeze
If the freezing feature is enabled on a token, then the issuer of the token can globally freeze that token, which means that nobody except the issuer can send that token. In other words, only the issuer will be able to send to other accounts. The issuer can also globally unfreeze and remove this limitation. 

### Whitelist
If the whitelisting feature is enabled, then every account that wishes to receive this token, must first be whitelisted by the issuer, otherwise they will not be able to receive that token. This feature allows the issuer to set whitelisted limit on any account, and then that account will be able to receive tokens only up to the whitelisted limit. If someone tries to send tokens to an account which will result in the whitelisted amount to be exceeded, the transaction will fail.

Here is the description of behavior of the whitelisting feature:
- The issuer can set whitelisted limit on any account except their own.
- The issuer can set whitelisted amount higher or lower than what the user currently holds.
- The issuer account is whitelisted to infinity by default and cannot be modified.
- The user can receive tokens as long as their total balance, after the transaction execution, will not be higher than their whitelisted amount
