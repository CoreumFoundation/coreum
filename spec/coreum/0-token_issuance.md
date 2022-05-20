
# Token Issuance

In order for a user to create their own token on **Coreum**, first they would have to initiate a `create token` transaction with the specification of the token as well as the address as to where the newly minted tokens will be transferred.
This operation mints specified tokens instantly after the execution of the transaction. 

### Minting and burning 
Depending on the configuration, a user can specify whether this token can have "mint", "burn" functionality in the future. If the mintable flag is set to true, then the user can mint new tokens using another transaction at any time.
Users can burn amounts of the issued token given the token is burnable. 

At the time of token issuance, the user can also specify whether holding (sending/receiving) and trading of this token can be allowed by anyone or should be restricted to users who are whitelisted.
The token issuer can set a "transaction fee" and "burn fee" in percentage so that for every movement (either through trading or sending/receiving tokens) this amount is deducted from the sender.

## Fungible vs Non-fungible tokens
Upon token issuance, users will be able to indicate whether the token will be **fungible or not**, together with the maximum supply amount.

NFTs represent ownership over a given set of digital or physical assets.

Some examples are the following:

- Physical property — land, houses, artwork
- Virtual collectables — unique monkey faces, avatars
- Financial assets — car loans, mortgages, etc

## Possible implementation

### Messages

#### Create token message
The issuer can issue new tokens using a `MsgIssueTokens`

```go
// MsgIssueTokens message type used to issue tokens
type MsgIssueTokens struct {
  Sender   sdk.AccAddress 
  Tokens   sdk.Coin 
  Receiver sdk.AccAddress
  Permission courem.PermissionSet //Burnable, mintable 
  Fees //Burn, transaction
  fungible boolean
}
```



