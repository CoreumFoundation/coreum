# Fee Structure
The purpose of this document is to outline the overall design of the fee model that will be used in Coreum. Basic concept of gas is described in [Gas and Fees](https://docs.cosmos.network/v0.45/basics/gas-fees.html) section of Cosmos SDK.

# Gas
Each transaction in Cosmos SDK uses fees to secure the network by paying validators, disincentivizing attacks, etc. The fee is determined by the following formula:

<p align="center">
<i>
fee = gasPrice * gasUsed
</i>
</p>

By providing two of the variables in the formula above, the other one can be determined. Variable gasUsed is invariant for each transaction but gasPrice may change which will result in having different fees at different times. Factors affecting gasPrice are but not limited to network congestion and price of the Coreum token (CORE). Coreum network will enforce a minimum gasPrice specified in units of CORE, [TBD] but users are free to provide a higher fee.

## Gas Used
The amount of gas used will be predetermined for different transaction types used by the pre-provided Cosmos SDK modules. For instance transferring coins will use some amount of gas which is hardcoded into the bank module and cannot be changed. Since at this stage of development we are assuming that we will be using many default Cosmos SDK modules, then when writing our own modules we should consume gas in proportion to the gas used for similar operations in the default modules. This means we need to compile down a table of transaction types and gas used in each one.
## Gas Price
Gas Price is the only variable in the fee formula that we can modify. This allows us to incentivize network users to do more or less transactions at any period of time and help run the network at an optimal state.
There are some criteria that must be met regarding the gas price.
When transactions are low the fees should be in a manner that it will incentivize the validators to keep participating in the network.
The more transactions coming in, the network will reduce the fees up to 50% until the desired tps is reached. This will incentivize the whole community to do more transactions up to the desired tps.
The more transactions coming in after the desired tps, fees will increase exponentially up to the network capacity to avoid network congestion. 

# User Experience 
Here we discuss user's perspective (interacting via the wallet). [TBD]All the user needs to provide is the maxFee that they are willing to pay for this transaction. 

More research needs to be done to figure out how wallets are communicating with the blockchain to figure the required fees.

# Security Concerns 
Fees are one of the mechanisms used to ensure network security. Here we try to provide a list of security concerns regarding the fees.

## DoS Attack Against Memory Pool
### Zero Fee Transactions
An adversary can decide to overload the memory pool of nodes in the network by publishing transactions with zero fees or from wallets that do not contain enough fees. To protect against this nodes must check that the issuer of the transaction has enough fees as is in the current state of the blockchain. This will happen in [**CheckTx**](https://docs.tendermint.com/master/spec/abci/abci.html#checktx) Method from ABCI protocol. [TBD] `CheckTx` maps to [**AnteHandlers**](https://docs.cosmos.network/v0.44/modules/auth/03_antehandlers.html) section of Cosmos SDK.

### Multiple Transactions from Single Account with Minimal Credits 
An adversary may try to overload the memory pool of the nodes in the network by publishing many transactions  from a single address with enough fees in the address to cover only one of those transactions. To protect against this different approaches might be taken:
1. Nodes can check that the issuer of the transaction has enough fees and deduct the fees from a copy of the current state of the blockchain. In this way only one of the transactions will get to the mempool and the others will fail. 
2. Nodes can enforce a maximum allowed transactions per address per a given period. 

These checks can occur in [**CheckTx**](https://docs.tendermint.com/master/spec/abci/abci.html#checktx) Method from ABCI protocol. [TBD] `CheckTx` maps to [**AnteHandlers**](https://docs.cosmos.network/v0.44/modules/auth/03_antehandlers.html) section of Cosmos SDK.

## Validators Artificially Withholding Transactions
It must be checked to ensure that validators cannot make higher profits by purposefully including less transactions in the block. In other words validators must always be incentivized to include more transactions in the block

## Not Enough CORE Tokens Getting Staked
The more token staked by network participants the more secure a POS network will become. In the cosmos hub the inflation rate is used to control how much token is staked by token holders. In other words they manipulate inflation rate to either incentivize or discourage token holders to stake their tokens. Since our system is deflationary, we cannot use inflation and a controlling mechanism and we should be mindful about the implications of our decision.

# Gas Price Proposals
## Proposal 1
We designate some variables defined below to utilize in gas price calculations:
- *BlockLoad*: Is an indicator of what percentage of block capacity is used
- *AverageBlockLoad<sub>n</sub>*: Is the average *BlockLoad* in the previous n blocks 
- *OptimalLoad*: Is the ideal *BlockLoad* that we want the blockchain to operate at. 
- *MaxBlockLoad*: Is the maximum *BlockLoad* blockchain can handle. 
- *MaxDiscount*: The maximum discount applied on type of the Initial Gas Price when the network is operating within the optimal range

The gas price will start at a predefined value called GP<sub>0</sub> if last *BlockLoad* is 0 and will reach to *MaxDiscount* of GP<sub>0</sub> when last *BlockLoad* reaches *AverageTPS<sub>n</sub>* it will remain at *MaxDiscount* of GP<sub>0</sub> until *OptimalTPS* of the network. After that point a fee escalation mechanism will kick in to avoid blockchain overload.

Gas price will be calculated by the following formula: 

IF *AverageBlockLoad<sub>n</sub>* < *BlockLoad*

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
gasPrice = GP<sub>0</sub> * (1 - maxDiscount)<sup>(*BlockLoad*/*AverageBlockLoad<sub>n</sub>*)</sup>

IF *AverageBlockLoad<sub>n</sub>* >= *BlockLoad* >= *OptimalBlockLoad*  

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
gasPrice = GP<sub>0</sub> * *MaxDiscount*

IF *BlockLoad* >= *OptimalBlockLoad*  

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
gasPrice = GP<sub>0</sub> * *MaxDicount* * 
(*MaxBlockLoad* - *OptimalBlockLoad*) / (*MaxBlockLoad* - *BlockLoad*)
