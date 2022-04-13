
#Whitelisting
To whitelist a user, the token issuer must initiate a transaction with the address of the requester as well as a “limit” amount (Limit indicates how much of this asset the whitelisted user can hold or trade).

**Example:**
- **User A** issues a new token called “ABC” and sets the config in a way that the newly minted tokens are transferred to **User B**.
- **User A** also restricts holding of this token by setting a config setting called “require_auth”. After this transaction, **User B** instantly gets whitelisted and is delivered the “ABC” tokens.
- Now user C wants to receive “ABC” tokens from **User B**. He must now ask user A (creator of the token) to authorize or whitelist them.
- After whitelisting is done, then user C can receive tokens from **User B**.

This ensures that **token creators comply** with their policies such as **regional regulations** (e.g. KYC, AML) before allowing the minted tokens to move around freely.

The creator of the token (user A) **can freeze anyone's “ABC” tokens held in their wallets** (given that they set the configuration that allows them to do so when they minted the token).

The same user (user A) can also set a **global freeze to all holders of the “ABC” token** at any time (again, given the fact that they set this config when minting the token).
The token creator/owner (user A) can also **change anyone's limit at any time** given the new limit is not smaller than the current user balance.
It is important to note that users **can also issue tokens without the whitelisting feature**. This creates tokens that can freely move around the Blockchain, such as USDT, or wrapped BTC.
**Token issuance is a built-in feature of the coreum Blockchain**, this means that minting tokens are not conducted through a smart contract, but rather natively in the Blockchain.


##Possible Implementation

###Messages
###Create token

###Add token to user token's list

###Freeze tokens

###Change limit amount
