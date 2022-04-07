
#Whitelisting
To whitelist a user, the token issuer must initiate a transaction with the address of the requester as well as a “limit” amount (Limit indicates how much of this asset the whitelisted user can hold or trade).

**Example:**
- **Bob** issues a new token called “ABC” and sets the config in a way that the newly minted tokens are transferred to **Alice**.
- **Bob** also restricts holding of this token by setting a config setting called “require_auth”. After this transaction, **Alice** instantly gets whitelisted and is delivered the “ABC” tokens.
- Now user C wants to receive “ABC” tokens from **Alice**. He must now ask user A (creator of the token) to authorize or whitelist them.
- After whitelisting is done, then user C can receive tokens from **Alice**.

This ensures that **token creators comply** with their policies such as **regional regulations** (e.g. KYC, AML) before allowing the minted tokens to move around freely.

##Possible Implementation
The Cosmos SDK [authz](https://docs.cosmos.network/v0.44/modules/authz/)  allows granting arbitrary privileges from one account (the granter) to another account (the grantee)