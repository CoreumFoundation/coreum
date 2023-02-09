# IBC integration plan

## IBC protocol
- we need someone to deeply understand how the IBC protocol works including: communication between relayer and chain
- meaning of all the messages and params in the module, can we define deterministic gas for them?
- check if there is anything related to governance
- understand difference between ordered and unordered channels
- understand data in the exported state of IBC
- inspect all IBC-related modules
- inspect IBC integration inside bank module (and others?)
- when `upgrade` module executes migrations there is some weird logic related to IBC. We need to understand it.

## Relayer
- it is possible to use single relayer for all the channels
- we should run at least two relayers
- we should run two full nodes for each blockchain we want to connect to. We may use external service delivering them (check https://www.zeeve.io/)
- we must monitor the funds of relayer on all the chains (grafana metric)
- if we want other entities to manage relayers we need to incentivize them to do it
- how relayer syncs two chains? how often? what transactions are used? how much gas they take?
- inspect IBC transactions on other chains to check how much gas they take
- check if it's possible to run many relayers serving the same channel
- check how relayer behaves when client, connection, channel or port is closed due to inactivity or any other reason. How can they be restored?
 
### Hermes vs Cosmos relayer
- Cosmos relayer is written in go, Hermes is written in rust
- Cosmos relayer supports many chains in single instance, need to check Hermes
- Hermes has a feature preventing channel from being automatically closed due to inactivity - check this
- We need to test both Hermes and Cosmos, integrate them into znet, test and compare their features and limitations

## FT integration

In general FT integration with IBC works out of the box via bank module. But we need to put special attention to how FT features behave in conjunction with IBC.

### FT IBC transfer
If FT is transferred to another chain it will behave like a standard token. All the limitations like whitelisting, freezing etc. don't apply there and FTs might be freely transferred between all the parties.

Once token is transferred back through the same channel, token becomes our FT back. But if it is transferred through yet another chain like Coreum -> Osmosis -> Cosmos -> Coreum, it will be visible on our chain like a bank token received through IBC - not an FT.

### Whitelisting
- whitelisting does not affect sending FT through IBC, transfer is always accepted.
- FT might be received back only by someone who is whitelisted

### Freezing and global freezing
If someone tries to transfer out amount exceeding unfrozen balance, transaction should be rejected.

### Burning
FTs transferred to other chain might be burned according to the rules implemented on that chain. It means that those FTs will never be transferred back, so they always are visible as existing tokens on Coreum.

### Minting
Minting is not affected by IBC in any way. Theoretically on the other chain, new IBC tokens might be minted intentionally or due to a bug. The question is if those tokens might enter our chain.

### Burn rate
Burn amount is taken from the sender account, so we should be able to charge it when FTs are transferred out. But there is no way to do it when FTs come back. We could charge recipient in this case, but should we?

Obviously burn rate does not apply when FT is transferred between accounts on other chains. Only rules implemented there apply.

### Send commission
Limitations described in "Burn rate" section apply here.

### Blocking IBC transfers for FT

Due to the fact that FTs transferred to othr chain are out of our control, we should add a new feature allowing issuer to block IBC transfers of that token.

### List of receivable tokens

Do we want to define a set of (token, channel) pairs we allow to be received via IBC?
It may reduce the mess but on the other hand will limit possibilities.

### Received tokens

Tokens received through IBC must be treated as regular bank tokens (like CORE), without any additional features provided by FT module.

## NFT integration

There is an ongoing work on NFT+IBC integration: https://github.com/bianjieai/ibc-go/tree/ics-721-nft-transfer
Should be available in ibc v5.

## WASM integration

In theory IBC might be used to transfer any messages between two compatible chains. So it could be used to transfer some data from one smart contract to another.

- Do we want to do it?
- It requires significant effort put on investigation first
- We don't even know at the moment if it really might be done in practice.

## Benchmarking
- write a benchmark integration test to check how IBC transfers behave under load
- check how much resources (disk space, memory, cpu?) each integrated chain takes on our side

## Other modules

Inspect modules in https://github.com/tendermint/spn/tree/main/x.
Some of them are IBC-elated and installed by ignite.
