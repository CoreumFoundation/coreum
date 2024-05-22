# x/asset/ft

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
- IBC transfers
- Clawback
- Extension

## Interaction with bank module, introducing wbank module
Since Coreum is based on Cosmos SDK, We should mention that Cosmos SDK provides the native bank module which is responsible for tracking fungible token creation and balances of each account. But this module does not allow any public to create a fungible token, mint/burn it, and also does not allow for other features such as freezing and whitelisting. To work around this issue we have wrapped the `bank` module into the `wbank` module.

In `wbank` module we wrap all the send related  methods of the `bank` module and intercept them with `BeforeSend` and `BeforeInputOutput` functions provided by `assetft` module. This allows `assetft` module to inject custom logic into interceptor functions and reject some transaction if whitelisting or freezing criteria are not met, or apply other features such as BurnRate.

This structure allows to reuse the code provided by Cosmos SDK, and also reuse the infrastructure that the community provides (e.g explorers and wallets). But it also leads to the fact that some of the information regarding fungible tokens will exist in the `assetft` module and some in the `bank` module. For example, if you want to query for frozen balances of a fungible token, you need to query the `assetft` module but if you want to get the total supply, you must query the bank module.

In a nutshell, `assetft` module interacts with `wbank` which in turn wraps the original `bank` module.

## Token Interactions
### Issue
Coreum provides a decentralized platform which allows everyone to tokenize their assets. Although the functionality of fungible token creation and minting is present in the original `bank` module of Cosmos SDK, it is not exposed to end users, and it is only possible to create new fungible tokens via either the governance or IBC. The Issue method described here, makes it possible for everyone to create a fungible token and manage its supply. When the admin (issuer of the token) issues a token, they specify the initial total supply which will be delivered to the admin account address.

All the information provided at the time of issuance is immutable and cannot be changed later.

### Denom naming, Symbol and Precision
The way that denom is created is that the user provides a name for their subunit, and the denom for the token, which is the main identifier of the token, will be created by joining the subunit and the issuer address separated with a dash (subunit-address). The user also provides the symbol and precision which will only be used for display purposes and will be stored in bank module's metadata field.

For example to represent Bitcoin on Coreum, one could choose satoshi as subunit, BTC as Symbol and 8 as precision. It means that if the issuer address is core1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8 then the denom will be `satoshi-core1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8` and since we have chosen BTC as symbol and 8 as precision, it will follow that (1 BTC = 10^8 `satoshi-devcore1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8`)

### Transferring admin
Each token has an issuer, whose address is a part of the denom forever. The initial admin of the token is the issuer, but the admin role can be transferred to another account.
Then, all the privileges of the previous admin will be transferred to the new admin, such as the ability to mint tokens if minting is enabled. The specific privileges and features will be discussed in the next section.

### Clearing admin
Tokens can also lose their admin forever by clearing admin.
Then, no one will have any more privilege than others.

## Token Features
When issuing a token, the admin must decide which features are enabled on the token. For example if `minting` feature is enabled then it will allow the admin to mint further tokens on top of the initial supply, otherwise no new tokens can be minted. Here is a list of all the features that can be enabled on the token. Each of these features affects how the token will behave and a detailed description will be provided in the dedicated section for each feature.

- minting
- burning
- freezing
- whitelisting
- ibc
- clawback

### Burn Rate
The admin has the option to provide `BurnRate` when issuing a new token. This value is a number between 0 and 1, and if it is above zero, in every transfer, some additional tokens will be burnt on top of the transferred value, from the senders address. The tokens to be burnt are calculated by multiplying the TransferAmount by burn rate, and rounding it up to an integer value.

If IBC feature is enabled for the token then the burn rate is applied to outgoing IBC transfers.

Burn rate is never applied if smart contract is the sender.

### Send Commission Rate
Exactly same as the Burn Rate, but the calculated value will be transferred to the admin account addressed (or the extension address if there is no admin) instead of being burnt.

If IBC feature is enabled for the token then the send commission rate is applied to outgoing IBC transfers.

Send commission rate is never applied if smart contract is the sender.

If a token doesn't have admin (ClearAdmin called on it), the commission rate would be set to zero, since there is no account to send the commissions to.
The only exception is tokens that have extension, since the extension can receive the commission.

### Issuance Fee
Whenever a user wants to issue a fungible token, they have to pay some extra money as issuance fee, which is calculated on top of tx execution fee and will be burnt. The amount of the issuance fee is controlled by governance.

### Mint
If the minting feature is enabled, then admin of the token can submit a Mint transaction to add more tokens to the total supply. All the minted tokens will be transferred to the admin account address.

### Burn
The admin of the token can burn the tokens that they hold. If the burning feature is enabled, then every holder of the token can burn the tokens they hold.

### Freeze/Unfreeze
If the freezing feature is enabled on a token, then the admin of the token can freeze an account up to an amount. The frozen amount can be more than what the user currently holds, an works even if the user holds zero. The user can only send the tokens that they hold in excess of the frozen amount.
For example if the admin freezes 1000 ABC tokens on account Y and this account holds 800, then they cannot move any of their tokens, but if the account receives 400 extra ABC tokens, their total balance will become 1200 and then can only spend 200 of it, since 1000 is Frozen.

Here is the description of behavior of the freezing feature:
- The admin can freeze an account up to amount if the freezing feature is enabled.
- The admin can increase the frozen amount by submitting new freeze transaction on top of already frozen account. The frozen amount of the account will be increased by the new value.
- The admin can unfreeze a portion of the frozen amount on an account.
- The admin cannot freeze their own account
- The user can only send their tokens in excess of the frozen amount.
- The user can receive tokens regardless of frozen limitation.
- The user cannot burn the frozen amount if both freezing and burning is enabled.
- Frozen amount cannot be a negative value, it means that amount present in unfreeze transaction cannot be bigger than the current frozen amount
- If either or both of BurnRate and SendCommissionRate are set above zero, then after transfer has taken place and those rates are applied, the sender's balance must not go below the frozen amount. Otherwise the transaction will fail.

Same rules apply to sending tokens over IBC transfer protocol if IBC is enabled for the token.

### Global Freeze/Unfreeze
If the freezing feature is enabled on a token, then the admin of the token can globally freeze that token, which means that nobody except the admin can send that token. In other words, only the admin will be able to send to other accounts. The admin can also globally unfreeze and remove this limitation.

If IBC is enabled for the token and token is globally frozen then only the admin can send them over IBC transfer protocol.

### Whitelist
If the whitelisting feature is enabled, then every account that wishes to receive this token, must first be whitelisted by the admin, otherwise they will not be able to receive that token. This feature allows the admin to set whitelisted limit on any account, and then that account will be able to receive tokens only up to the whitelisted limit. If someone tries to send tokens to an account which will result in the whitelisted amount to be exceeded, the transaction will fail.

Here is the description of behavior of the whitelisting feature:
- The admin can set whitelisted limit on any account except their own.
- The admin can set whitelisted amount higher or lower than what the user currently holds.
- The admin account is whitelisted to infinity by default and cannot be modified.
- The user can receive tokens as long as their total balance, after the transaction execution, will not be higher than their whitelisted amount

Same rules apply to receiving tokens over IBC transfer protocol if IBC is enabled for the token.

### IBC
When token is created, admin decides if users may send and receive it over IBC transfer protocol.
If IBC feature is disabled token can never leave the Coreum chain.

### Clawback
If the clawback feature is enabled on a token, then the admin of the token can confiscate up to the amount an account holds. The clawback amount cannot be more than what the user currently holds.

Here is the description of behavior of the clawback feature:
- The admin can clawback up to the amount an account holds if the clawback feature is enabled.
- The admin cannot clawback from module accounts

Same rules apply to sending tokens over IBC transfer protocol if IBC is enabled for the token.

_**Disclaimer**: if the admin claws back from the escrow address, then it will break the IBC. admins should not do this if they want the IBC to work for their token._

### Extension
Extension is a powerful feature which lets the admin override some functionalities of the token by attaching a smart contract to the token that can administrate it. When a bank transfer is initiated, the smart contract account receives the amount plus any commission or burn amount if they should be applied, then it is called via a sudo call with the information related to the bank transfer and the context information. The smart contract then can decide to do whatever it decides with the transfer which may include overriding of some behaviors for the features explained before.
The sudo call is received in the `pub fn sudo(deps: DepsMut<CoreumQueries>, env: Env, msg: SudoMsg)` entry point of the smart contract and the message would be a `ExtensionTransfer` which is defined as follows.
```rust
#[cw_serde]
pub enum SudoMsg {
    ExtensionTransfer {
        recipient: String,
        sender: String,
        transfer_amount: Uint128,
        commission_amount: Uint128,
        burn_amount: Uint128,
        context: TransferContext,
    },
}

#[cw_serde]
pub struct TransferContext {
    sender_is_smart_contract: bool,
    recipient_is_smart_contract: bool,
    ibc_purpose: IBCPurpose,
}
```

The fields are:
- `recipient`: the account which receives the amount
- `sender`: the account which sends the amount
- `transfer_amount`: the amount to be sent
- `commission_amount`: the amount to be sent times the commission rate of the token (if it is set)
- `burn_amount`: the amount to be sent times the burn rate of the token (if it is set)
- `context`: contextual information regarding the transfer which includes:
  - `sender_is_smart_contract`: indicated whether the transfer is instantiated from a smart contract
  - `recipient_is_smart_contract`: indicated whether the transfer is going to be received by a smart contract
  - `ibc_purpose`: if it is an ibc transfer, indicates whether it's an outgoing, incoming, acknowledged or timed-out transfer

There is a sample implementation of extension in `x/asset/ft/keeper/test-contracts/asset-extension` which can be used to take inspiration from, when implementing other extensions.

## Feature interoperability table

<!-- Original source: https://docs.google.com/spreadsheets/d/1wC51asxQF8gi7Egj0KvzsMf7zko5ojEL6l2CAdb_UNM -->
<!-- Tool to generate table: https://www.tablesgenerator.com/html_tables -->

<table>
<thead>
  <tr>
    <th rowspan="3"></th>
    <th colspan="14">Features</th>
    <th colspan="4">Extensions</th>
  </tr>
  <tr>
    <th colspan="2">Default</th>
    <th colspan="2">Minting</th>
    <th colspan="2">Burning</th>
    <th colspan="2">Freezing</th>
    <th colspan="2">Whitelisting</th>
    <th colspan="2">IBC</th>
    <th colspan="2">Clawing back</th>
    <th colspan="2">Burn rate</th>
    <th colspan="2">Send commission rate</th>
  </tr>
  <tr>
    <th>Admin</th>
    <th>Owner</th>
    <th>Admin</th>
    <th>Owner</th>
    <th>Admin</th>
    <th>Owner</th>
    <th>Admin</th>
    <th>Owner</th>
    <th>Admin</th>
    <th>Recipient</th>
    <th>Admin</th>
    <th>Owner</th>
    <th>Admin</th>
    <th>Owner</th>
    <th>Admin</th>
    <th>Owner</th>
    <th>Admin</th>
    <th>Owner</th>
  </tr>
</thead>
<tbody>
  <tr>
    <td>Mint</td>
    <td></td>
    <td></td>
    <td>➕</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
  </tr>
  <tr>
    <td>Burn</td>
    <td>➕</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td>➕</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
  </tr>
  <tr>
    <td>Freeze</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td>➕</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
  </tr>
  <tr>
    <td>Unfreeze</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td>➕</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
  </tr>
  <tr>
    <td>GloballyFreeze</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td>➕</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
  </tr>
  <tr>
    <td>GloballyUnfreeze</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td>➕</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
  </tr>
  <tr>
    <td>Whitelist</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td>➕</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
  </tr>
  <tr>
    <td>Unwhitelist</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td>➕</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
  </tr>
  <tr>
    <td>Send</td>
    <td>➕</td>
    <td>➕</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td><a href="#freezeunfreeze">ⓘ</a></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td><a href="#burn-rate">ⓘ</a></td>
    <td></td>
    <td><a href="#send-commission-rate">ⓘ</a></td>
  </tr>
  <tr>
    <td>Send to admin</td>
    <td>➕</td>
    <td>➕</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
  </tr>
  <tr>
    <td>Send from smart contract</td>
    <td>➕</td>
    <td>➕</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
  </tr>
  <tr>
    <td>Receive</td>
    <td>➕</td>
    <td>➕</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td><a href="#whitelist">ⓘ</a></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
  </tr>
  <tr>
    <td>IBC Send</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td><a href="#freezeunfreeze">ⓘ</a></td>
    <td></td>
    <td></td>
    <td>➕</td>
    <td>➕</td>
    <td></td>
    <td></td>
    <td></td>
    <td><a href="#burn-rate">ⓘ</a></td>
    <td></td>
    <td><a href="#send-commission-rate">ⓘ</a></td>
  </tr>
  <tr>
    <td>IBC Send to admin</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td>➕</td>
    <td>➕</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
  </tr>
  <tr>
    <td>IBC Receive</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td><a href="#whitelist">ⓘ</a></td>
    <td>➕</td>
    <td>➕</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
  </tr>
  <tr>
    <td>Clawback</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td>➕</td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
    <td></td>
  </tr>
</tbody>
</table>

**Legend**:

* **Admin** : The FT admin.
* **Owner** : The account that owns some FT amount.
* **Recipient** : The FT recipient in the `send` operation.
* **Default** : The **Default** is the state that the FT has without any feature. Adding the **Features** to the FT, you
  can extend or override the **Default** functionality.
* **➕** : Allowing
* **ⓘ** : Custom behaviour

# Upgrading token to v1

In version `v2` of the blockchain we introduced an ability for the tokens to be sent to and received from other Cosmos SDK-based chains using IBC protocol.

Now, whenever new fungible token is created, the admin might decide if IBC transfers are enabled or disabled for that token. The decision might be made once, when token is created, and can never be changed later. It is done by specifying `ibc` feature in the token issuance transaction.

For admins of pre-existing tokens, those created on version `v1` of the blockchain, we open the time-limited opportunity to enable IBC transfers on those tokens.

All the pre-existing tokens are now considered to be of version `v0`. New tokens, created after `v2` blockchain upgrade, will be of version `v1`.

The mechanism described below enables admins to upgrade tokens from version `v0` to `v1` after the blockchain itself is upgraded to version `v2`.

After the blockchain upgrade, the 3-week period starts when admin must make a decision if they want to enable IBC or not. Three possible outcomes exist:
- admin does not take any action during the transition period - after transition period ends, token stays in version `v0`, can never be upgraded to `v1` again, IBC transfers for that token are disabled forever,
- admin decides to disable IBC - token version is upgraded from `v0` to `v1` immediately, IBC transfers are disabled for that token forever, IBC can never be re-enabled even if admin changes his mind before the transition period ends,
- admin decides to enable IBC - for next week nothing happens, it is called a `grace period`, when token holders might recognize the pending upgrade and decide if they still want to keep the token or liquidate it (if they don't support the decision made by the admin). After the grace period, token is automatically upgraded from `v0` to `v1` and IBC is enabled. Admin cannot change the decision even if grace period or transition period is still pending.

## Commands to upgrade the token

To upgrade the token and **enable IBC**, use this command:

```console
cored tx assetft upgrade-v1 [denom] --ibc-enabled=true --from [sender]
```

To upgrade the token and **disable IBC**, use this command:

```console
cored tx assetft upgrade-v1 [denom] --ibc-enabled=false --from [sender]
``` 

## Querying the upgrade status

Everyone is able to check what decision has been made by the admin, by running the query command:

```console
cored q assetft token-upgrade-statuses [denom]
```
