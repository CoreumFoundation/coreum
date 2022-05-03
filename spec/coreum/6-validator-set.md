# Validators
Active validator set consist of 2 validator sets
1. [Super Validators](#super-validator) 
2. [Public Validators](#public-validator)

which will be chosen from the [Candidates](#candidates)

## Candidates
Any Token holder that stakes a minimum amount of CORE token([TBD] e.g 1% of total supply) either directly or via delegation can choose to become a Candidate Validator which lets them to become an either [Super Validator](#super-validator) or [Public Validator](#public-validator).

## Super Validator 
Super Validators are chosen amongst the [Candidates](#candidates) via voting for a [specific amount of time](#super-validator-duration). There are a maximum number of super validators.
Voting for new Super Validator set will start at the start of the current round and must finish before the end of the current round. If there are not enough validators applying for Super Validator position those empty positions will be filled by public validators.

## Public Validator
Public Validators are chosen randomly among the candidates for a [period
of time](#super-validator-duration) . This will help all the candidates to have a 
chance to take part in the validation process.

## Params
##### Super Validator Duration
[TBD] 4 days
##### Super Validator Set Size
[TBD] 4
##### Public Validator Set Size
[TBD] 12
##### Public Validator Duration
[TBD] 4 hours 

# Reward Distribution
- [TBD] Do we want to distribute block rewards according the the stake of validators or evenly? (consider [This](#more-rewards-for-super-validators))
- [TBD] Do we want to manage delegation reward distribution on chain ?

# Concerns
## Questions and Unknowns
- What happens if a super validator is jailed ? do we replace them from public validator set ?

## Staking, Slash, Evidence rewrites
With our proposal we need to either extend current staking module or create our own (more likely to create our own). This will affect slash and evidence modules and we will likely need to rewrite them as well. 

## More Rewards for Super Validators
[Super Validator](#super-validator) will stay among active validators for a much longer time (e.g 20 times more). 
If we distribute the rewards among the active validators according to their stake, 
then Super Validators will yield more (e.g 20 times more) if there are many [Candidates](#candidates). 
This leads to a situation that the number of [Candidates](#candidates) will approach the size of active validator set. One could argue that it is simpler to just take top Public Validators according to the stake and do away with random selection.

## How to Introduce Fairness in Public Validator Selection
#### Proposal 1: Use Previous Block Hash
One proposal is to use the hash of the previous block as seed for a pseudorandom algorithm. But the concern is that the proposer of the last block can manipulation block formation in a way that they are included in the next block. 

#### Proposal 2: Use Large Window to Precalculate Validator Set and Enforce Fairness Rules
Another proposal is to use a large window in which we propose future Validator set and if this window is big enough we can check and see if the Fairness rules are observed (i.e every validator is chosen according to their stake). 

#### Proposal 3: Use a Priority Queue 
A possible good algorithm can be similar to what tendermint is using for proposer selection as described [Here][TendermintProposerSelection]. It is important to note that a large enough Unbounding period must be selected to avoid manipulation of this algorithm.


# Cosmos SDK Considerations
Validator set can be updated by sending validator updates to [EndBlock](https://docs.tendermint.com/master/spec/abci/abci.html#endblock) function of Tendermint, there is function with the same name in AppModule interface in Cosmos SDK. It takes 3 blocks for those updates to take affect as described in tendermint doc. In order to remove a validator its power must be set to zero.

## Tendermint Proposer Selection
To correctly reason about validator set, it is important to fully understand how tendermint operates when it comes to validators. Tendermint Proposer selection algorithm is described [Here][TendermintProposerSelection]. 

[TendermintProposerSelection]: https://docs.tendermint.com/master/spec/consensus/proposer-selection.html