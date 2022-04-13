# Token Issuance
##Background
Token issuance is the process of **creating new tokens** that are then **added to the total supply** of the cryptocurrency

Token issuance may also refer to the process of tokenization, in which an asset outside of the cryptocurrency ecosystem is added to the blockchain via a specific crypto token. In such cases, token issuance becomes the process of creating a token, yet not one that belongs to a cryptocurrency, but rather a token that represents an outside asset.

##Cosmos SDK auth module
The auth module exposes the account keeper, which allows other modules to read, write, and modify accounts. 

Accounts contain authentication information for a uniquely identified external user of an SDK blockchain including:
- public key, 
- address
- account number 
- sequence number for replay protection. 
- Balance

`AccountI` is an interface used to store coins at a given address within state.

```go

type AccountI interface {
    proto.Message
    
    GetAddress() sdk.AccAddress
    SetAddress(sdk.AccAddress) error // errors if already set.
    
    GetPubKey() crypto.PubKey // can return nil.
    SetPubKey(crypto.PubKey) error
    
    GetAccountNumber() uint64
    SetAccountNumber(uint64) error
    
    GetSequence() uint64
    SetSequence(uint64) error
    
    // Ensure that account implements stringer
    String() string
}
```
A base account is the simplest and most common account type, which just stores all requisite fields directly in a struct.

`BaseAccount` defines a base account type. It contains all the necessary fields for basic account functionality. 

Any custom account type should extend this type for additional functionality (e.g. vesting).
```go
message BaseAccount {
  string address = 1;
  google.protobuf.Any pub_key = 2;
  uint64 account_number = 3;
  uint64 sequence       = 4;
}
```
