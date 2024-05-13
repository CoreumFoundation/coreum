Article about ticks -

### Problem

Lets say we have 2 orders:
order1: 50_000_000 AAA (price 0.375) -> 18_750_000 BBB.
order2: 10_000_000 BBB (price 2.631) -> 26_310_000 AAA. (reversed price: ~0.38)

Note that it doesn't matter either we define orders as input & output amount or input and price the main rule is to respect price.

Since the reversed price of order2 is bigger than the price of order1 (0.38 > 0.375) these orders match.
But if we want to calculate the exact amount of AAA order2 receives it results into: 26666666.66(6)
which is not okay since we can only do calculations with integers.
Rounding up or down is also not acceptable since it violates price expectation of one or another order.
Another option is to find the biggest amount less than 26666666 which matches exactly
and cancel a reminder.
In this case, it could be 26666664 since 26666664*0.375=9_999_999 so both numbers are integers.
But such amount reduces order1 to 40_000_001 AAA (price 0.375) -> 15_000_000.375 BBB.
As you can see it results into decimal number of BBB in order1 which will cause issues during matching of the next order.

### Solution

Exchanges to avoid this issue define [ticks size](https://www.investopedia.com/terms/t/tick.asp) for each pair.
But since inside our DEX implementation we want to be able to trade anything
to anything number of pairs increase exponentially with addition of each token.
What is more trading should be implemented in permission-less manner, so none can configure each pair specifically.

Tick size mostly depends on the price of the asset traded
(could also depend on an asset type but this will be addressed later).
But since our DEX allows users to trade anything to anything (not only to fiat)
it means that tick size should depend on price of both assets to satisfy all possible pairs.
As result mathematical function for tick_size for pair AAA/BBB would look like this:
`f(price_A,price_B)` and for BBB/AAA `f(price_B,price_A)`.
However, using prices to define ticks has some drawbacks:
1. In crypto, prices are very volatile as result tick size might change quite frequently.
While we want to make this happen rarely to avoid heavy computations and to not spoil user experience.
2. Some special assets might need to have less or more precise tick size depending not only on price but also asset type.

The proposed solution is for each asset to introduce an attribute named `min_amount_increment`.
`min_amount_increment` for token represents minimum amount of a token which could take part in trading.
Tick size for a trading pair will be calculated based on `min_amount_increment` of two assets.

In our DEX both AAA/BBB and BBB/AAA pairs are allowed,
so from definition of `min_amount_increment` we can have 2 math equations:
```
min_amount_increment(A) * tick_size(AAA->BBB) >= min_amount_increment(B)
min_amount_increment(B) * tick_size(BBB->AAA) >= min_amount_increment(A)
```

which results into

```
tick_size(AAA->BBB) >= min_amount_increment(B) / min_amount_increment(A)
tick_size(BBB->AAA) >= min_amount_increment(A) / min_amount_increment(B)
```

The proposed formula for tick size is:
```
tick_size(AAA->BBB) = tick_size_coeficient * min_amount_increment(B) / min_amount_increment(A)
Where tick_size_coeficient <= 1.0
```

Empirically (by experimenting with different assets) `0.01`
has been chosen as value for `tick_size_coeficient` but it could be easily changed in the future if needed.

### Examples

| min_amount_increment(A)         | min_amount_increment(B) | tick_size(A->B)                | tick_size(B->A)              |
|---------------------------------|-------------------------|--------------------------------|------------------------------|
| 10_000 (10_000ucore = 0.01CORE) | 10_000                  | 0.01 (1_000_000A*0.01=10_000B) | 0.01(1_000_000B*0.01=10_000A |
| 1000 (1000satoshi = 0.00001BTC) | 10                      | 0.0001 (100_000A*0.0001=10B)   | 1 (1_000B*1=1000A )          |

[//]: # (| 1 &#40;1DOGE=10^x&#41;) TODO: finish
