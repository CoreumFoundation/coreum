# x/asset/nft

This document specifies `assetnft` module, which allows public users of the blockchain to create non-fungible tokens on the Coreum blockchain.

# Concepts
This module provides transactions and queries which allows public users of the blockchain to issue non-fungible
tokens. The information for the NFTs themselves are stored in the `nft` module developed by Cosmos team,
but that module does not allow public users to issue NFT classes or mint NFTs, and that's where
this module comes in. The interaction between the two modules is described [here](#interaction-with-nft-module-introducing-wnft-module). This module also introduces `features` that defines specific behavior for the nft (described [here](#token-features)).

## Interaction with nft module, introducing wnft module
The Cosmos team has developed the `nft` module (which we hereby refer to as the `original nft module`),
which can be used to store the information about NFTs, their classes, their ownership, etc. But as mentioned earlier this  module does not provide any functionalities to public users to create their own NFTs, or define
custom behavior for transferring tokens. Because of this reason we have wrapped the `original nft module` into
the `wnft` module, which allows injecting custom logic into the transfer method of the `original nft module`.
We have also created the `assetnft` module (this module), to allow public users to create their own NFTs with
their own custom behavior.

In other words the `assetnft` module defines the custom behavior for NFTs, enforces that behavior by injecting
custom logic into `wnft` module, and keeps most NFT related information on the `original nft moduel`.

This design means that some portion of data relating to NFTs will live in this module, and some will live in the
`original nft module`, so to get the final NFT functionality one should be aware and understand that they should
make some of the queries to the `original nft module`.
## Token Features
NFT tokens come with a set of features that the issuer can specify at the time of issuing a class, and then in some cases configured on each NFT level later.

Here is the list of features:
- burning
- freezing
- whitelisting
- disable sending
- royalty rate

We will discuss each feature separately.

### Burning
If this feature is enabled, it allows the holders of the token to burn the tokens they hold.
It should be noted here that the issuer can burn their token regardless of this feature.

### Freezing
If this feature is enabled, it allows the issuer of the class to freeze any NFT token in that class.
A frozen token cannot be transferred until it is unfrozen by the issuer.

### Whitelisting
If this feature is enabled, then for any user to receive any NFT of that class, they must be whitelisted to
receive that specific NFT. It follows that this feature allows the issuer of the class to whitelist an
account to hold an specific NFT of that class, or remove an account from whitelisted accounts for that NFT.

### Disable Sending
If this feature is enabled, then the NFT cannot be directly transferred between users, meaning that user A cannot
send the tokens they hold directly to user B. This feature opens up the door for different use cases, one of which is that it might be used to force transfer of ownership to go via DEX, so that the royalty fee is applied and the creator of the NFT always gets a royalty fee.

### Royalty Rate
This feature is related to the DEX, and if it is enabled, every time that an NFT is traded on the DEX, a percentage of the the traded value is sent to the issuer as royalty fee.
