# Matching

The spec describes the orders matching.

## Rounding issue

Let's say we have 2 orders:

| order_id | sell_denom | buy_denom | quantity       | price |
|----------|------------|-----------|----------------|-------|
| order1   | AAA        | BBB       | 50_000_000 AAA | 0.375 | 
| order2   | BBB        | AAA       | 10_000_000 BBB | 2.6   | 

*Note that it doesn't matter whether we define orders as input & output amount or input and price the main rule is to
respect price.*

The `price` here is the amount you want to get for each token you sell.
The `quantity` here is the amount you want to sell.

The rev_price (1/price) of order2 is greater than the price of order1 (~0.3846 > 0.375)
hence orders match. The order2 should be executed with the price of the order1 (taker gets the better price).
The exact amount of AAA order2 receives is 10_000_000 * 1/0.375 = 26_666_666.(6) AAA. The amount we can use
must be an integer, so we can't send full amount without the price violation.

## Solution

### Tick size

To avoid the price violation issue and provide a better trading experience we define
the [tick_size](https://www.investopedia.com/terms/t/tick.asp) for each pair. The pair in the Coreum DEX is
bidirectional, so we need to define the `tick_size` for each pair. `tick_size` mostly depends on the price of the assets
traded, that's why we can define the variable for a token used to define the pair `tick_sizes`. This variable is
named `significant_amount`. `significant_amount` for token represents minimum valuable amount of a token. The
recommended targeting price for the significant amount is 1 USD dollar cent.

The formula for tick size is:

```
tick_size(AAA/BBB) = tick_multiplier * significant_amount(BBB) / significant_amount(AAA)
tick_size(BBB/AAA) = tick_multiplier * significant_amount(AAA) / significant_amount(BBB)
```

The `tick_size_multiplier` is the coefficient used to give better price precision for the token pairs. The default value
of the `tick_multiplier` is  `0.01`, and can be updated by the governance.

Example:

| significant_amount(AAA) | significant_amount(BBB) | tick_size(AAA/BBB) | tick_size(BBB/AAA) |    
|-------------------------|-------------------------|--------------------|--------------------|
| 10_000                  | 10_000                  | 0.01               | 0.01               | 
| 1000                    | 10                      | 0.0001             | 1                  | 
| 1_000_000               | 1                       | 0.00000001         | 10_000             | 

### Matching with max execution quantity

With this strategy, when orders match, we target to fill order `remaining_quantity` fully. In case we can't do it fully
due to the rounding, we find `max_execution_quantity` that we can use to prevent the price violation, and return
the remainder.

The `max_execution_quantity` is the maximum integer value of the `remaining_quantity` (integer) which gives integer when
we multiply it by the execution price. The execution price might be the maker order price, if we try to fill the maker
order `remaining_quantity`. Or one divided by the maker order price, if we try to fill the taker
order `remaining_quantity`.

Let's find the formula for the `max_execution_quantity` :

```
Qa - quantity of token AAA to trade (integer)
P - execution price (decimal)
Qa' - final quantity of token A to be traded (integer)
Qb' - final quantity of token B to be traded (integer)
P = pn / pd, where pn is price numerator (integer) and pd - price denominator (integer)
```

We can define the `Qb'` as:

```
Qb' = Qa' * P = Qa' * pn / pd
```

To make `Qb'` an integer `pd` must be reduced to 1 by being reduced with `pn` or `Qa'`.
Because `pn / pd`, by definition, is an irreducible fraction it means `pn` and `pd` already have no common divider other
than 1.

It means that `pd` must be reduced fully by `Qa'` exclusively.
It means that `Qa'` must be a multiple of pd.

That's why

```
Qa' = floor(Qa / pd) * pd
```

generates the biggest `Qa'` divisible by `pd`.

Which we can re-write as

```
max_execution_quantity = floor(remaining_quantity / price_denominator) * price_denominator
```

Based on the formula we can define the max not filled quantity (the remained). The max not filled quantity depends on
the order we fill fully. If the order is taker order, the max remainder is equal to the maker price numerator, if maker,
the maker price denominator.

#### Matching algorithm

The following algorithm is used to match the orders:

```
if 1/(taker_price) > rev_order_price:
    if maker_remaining_quantity * maker_price > taker_quantity
      fill taker max_execution_quantity of quantity with the 1/(maker_price)
      rev_execution_quantity = max_execution_quantity * 1/(maker_price)
      taker receives:
        taker_quantity - max_execution_quantity (remainder)
        rev_execution_quantity
      maker receives: 
        max_execution_quantity
      taker order is closed  
      maker order is reduced by rev_execution_quantity
    else
      fill maker max_execution_quantity of remaining_quantity with the maker_price
      rev_execution_quantity = max_execution_quantity * maker_price
      maker receives:
        rev_execution_quantity
        maker_remaining_quantity - max_execution_quantity (remainder)
      taker receives: 
        max_execution_quantity
      maker order is closed  
      taker order is reduced by rev_execution_quantity
```

#### Matching example

Let's say we have two tokens AAA and BBB.

```
significant_amount(AAA) = 100
significant_amount(BBB) = 10

tick_size(AAA/BBB) = 0.01 * 10 / 100 = 0.001
tick_size(BBB/AAA) = 0.01 * 100 / 10 = 0.1
```

If `sell_denom` is `AAA` and `buy_denom` is `BBB` the `price` must be multiple of `tick_size(AAA/BBB)` and vice
versa.

* Round 1

Let's say we have 2 orders:

| order_id | account  | sell_denom | buy_denom | quantity       | price | remaining_quantity | 
|----------|----------|------------|-----------|----------------|-------|--------------------|
| order1   | account1 | AAA        | BBB       | 50_000_000 AAA | 0.371 | 50_000_000 AAA     |  
| order2   | account2 | BBB        | AAA       | 10_000_000 BBB | 2.6   | 10_000_000 BBB     | 

The rev_price (1/price) of order2 is greater than the price of order1 (~0.3846 > 0.371)
hence orders match. The order2 should be executed with the price of the order1 (taker gets the better price).
Maker expected amount is 50_000_000 * 0.371 = 18_550_000 BBB. The 18_550_000 > 10_000_000, that's why we fill the
10_000_000 BBB. The `max_execution_quantity` of 10_000_000 with the 1/0.371 price is 9_999_934 (the remainder is 66
BBB).

The result order book and account balances look like:

| order_id | account  | sell_denom | buy_denom | quantity       | price | remaining_quantity                                 |
|----------|----------|------------|-----------|----------------|-------|----------------------------------------------------|
| order1   | account1 | AAA        | BBB       | 50_000_000 AAA | 0.371 | 50_000_000 - (9_999_934 / 0.371) = 23_046_000  AAA |

| account  | balance                |
|----------|------------------------|
| account1 | 9_999_934 BBB,         | 
| account2 | 26_954_000 AAA, 66 BBB |

The account1's order1 was filled partially with the exact price 9_999_934 BBB / 26_954_000 AAA = 0.371

The account2 expected to sell 10_000_000 BBB, receive 26_000_000 AAA, but received more 26_954_000 AAA and
additionally not filled 66 BBB.

* Round 2

Let's say we have the previous partially filled order and new order:

| order_id | account  | sell_denom | buy_denom | quantity       | price | remaining_quantity |
|----------|----------|------------|-----------|----------------|-------|--------------------|
| order1   | account1 | AAA        | BBB       | 50_000_000 AAA | 0.371 | 23_046_000 AAA     |
| order3   | account3 | BBB        | AAA       | 70_000_000 BBB | 2.3   | 70_000_000 BBB     |

The rev_price (1/price) of order3 is greater than the price of order1 (~0.4347 > 0.371)
hence orders match. The order2 should be executed with the price of the order1 (taker gets the better price).
Maker expected amount is 23_046_000 * 0.371 = 8_550_066 BBB. The 8_550_066 < 70_000_000, that's why we fill the
23_046_000 AAA. The `max_execution_quantity` of 23_046_000 with the 0.371 price is 23_046_000 AAA.

The result order book and account balances look like:

| order_id | account  | sell_denom | buy_denom | quantity       | price | remaining_quantity                                 |
|----------|----------|------------|-----------|----------------|-------|----------------------------------------------------|
| order3   | account3 | BBB        | AAA       | 70_000_000 BBB | 2.3   | 70_000_000 - (23_046_000 * 0.371) = 61_449_934 BBB |

| account  | balance                                 |
|----------|-----------------------------------------|
| account1 | 9_999_934 + 8_550_066 = 18_550_000 BBB, | 
| account2 | 26_954_000 AAA, 66 BBB                  |
| account3 | 23_046_000 AAA                          |

As a result the account1's order1 is executed with full match of both quantity and amount
As a result the account3's order3 is executed partially with the better price

* Round 3

Let's say we have the previous partially filled order and new order:

| order_id | account  | sell_denom | buy_denom | quantity        | price | remaining_quantity |
|----------|----------|------------|-----------|-----------------|-------|--------------------|
| order3   | account3 | BBB        | AAA       | 70_000_000 BBB  | 2.3   | 61_449_934 BBB     |
| order4   | account4 | AAA        | BBB       | 220_000_000 AAA | 0.36  | 220_000_000 AAA    |

The rev_price (1/price) of order4 is greater than the price of order3 (~2.7(7) > 2.3)
hence orders match. The order2 should be executed with the price of the order3 (taker gets the better price).
Maker expected amount is 61_449_934 * 2.3 = 141_334_848.2 AAA. The 141_334_848.892 < 220_000_000,
that's why we fill the 61_449_934 BBB. The `max_execution_quantity` of 61_449_934 with the 2.3 price is 61_449_930
BBB
(4 BBB is the remainder).

The result order book and account balances look like:

| order_id | account  | sell_denom | buy_denom | quantity        | price | remaining_quantity                              |
|----------|----------|------------|-----------|-----------------|-------|-------------------------------------------------|
| order4   | account4 | AAA        | BBB       | 220_000_000 AAA | 0.36  | 220_000_000 - (61449930 * 2.3) = 78_665_161 AAA |

| account  | balance                                                  |
|----------|----------------------------------------------------------|
| account1 | 18_550_000 BBB,                                          | 
| account2 | 26_954_000 AAA, 66 BBB                                   |
| account3 | 23_046_000 + (61_449_930 * 2.3) = 164_380_839 AAA, 4 BBB |
| account4 | 61_449_930 BBB                                           |

The order3 is executed fully now with the better price (expected 161_000_000AAA), but received 164_380_839 AAA, 4 BBB