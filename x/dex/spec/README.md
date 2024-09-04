# DEX

The spec describes the coreum DEX specification.

## Order book

The Coreum DEX orders can be created for any-to-any token by any Coreum user. Which meas that the order book
is bidirectional and permissionless.

The user can place an order with the attributes:

* `order_id` - unique order identifier of the order.
* `base_denom` - when you buy, you are buying the `base_denom`, when you sell, you are selling the `base_denom`.
* `quote_denom` - when you buy, you are selling the `quote_denom`, when you sell, you are buying the `quote_denom`.
* `price` - value of one unit of the `base_denom` expressed in terms of the `quote_denom`. It indicates how much of the
  `quote_denom` is needed to buy one unit of the `base_denom`.
* `quantity` - is amount of the base `base_denom` being traded.
* `side`
    * `sell` - means that the order is to sell `base_denom` `quantity` with the `price`.
    * `buy` - means that the order is to buy `base_denom` `quantity` with the `price`.
* `time_in_force` - how long an order will remain active before it is executed or expires, based on matching state.
    * `GTC` - Good Til Canceled
    * `IOC` - Immediate Or Cancel
    * `FOK` - Fill or Kill
* `good_til` - how long an order will remain active before it is executed or expires, based height or time.
    * `good_til_block_height` - max block height to execute the order, or it will be canceled.
    * `good_til_block_time` - max block time to execute the order, or it will be canceled.

Once an order is placed, the DEX will try to match the order with the same order book opposite side orders and opposite
order book same side orders. And depending on order type and settings execute it.

## Order placement and matching

### Rounding issue

Let's say we have 2 orders from the opposite order books:

| order_id | base_denom | quote_denom | side | remaining_quantity | price |
|----------|------------|-------------|------|--------------------|-------|
| order1   | AAA        | BBB         | sell | 500000000          | 0.375 |
| order2   | BBB        | AAA         | sell | 10000000           | 2.6   |

The inverse taker price (1/price) of order2 is greater than the price of order1 (~0.3846 > 0.375) hence orders match.
The order2 should be executed with the price of the order1 (taker gets the better price). The exact amount of AAA order2
receives is 10000000 * 1/0.375 = 26666666.(6) AAA. The amount we can use must be an integer, so we can't send full
amount without the price violation. The following solution resolves that issue.

### Matching with max execution quantity

With this strategy, when orders match, we target to fill the order `remaining_quantity` with lower volume (the one which
will be closed) fully. In case we can't do it fully due to the rounding, we find `max_execution_quantity` that we can
use to prevent the price violation, and return the remainder to the user's balance.

The `max_execution_quantity` is the maximum integer value that is less or equal to the `remaining_quantity` (integer)
which gives integer when we multiply it by the execution price. The execution price might be the maker order price,
if we try to fill the maker order `remaining_quantity`. Or one divided by the maker order price, if we try to fill the
taker order `remaining_quantity`.

Let's find the formula for the `max_execution_quantity` :

```
Qa - quantity of token AAA to trade (integer)
P - execution price (decimal)
Qa' - final quantity of token A to be traded (integer)
Qb' - final quantity of token B to be traded (integer)
P = pn / pd, where pn is price numerator (integer), pd - price denominator (integer), pd/pn is an irreducible fraction
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

And the `Qb'` can be written as:

```
Qb' = floor(Qa / pd) * pn
```

The `Qb'` is the opposite order's `max_execution_quantity` (the amount by which the opposite order is reduced)

```
opposite_execution_quantity = floor(remaining_quantity / price_denominator) * price_numerator
```

Let's return back to the example with the [Rounding issue](#Rounding-issue) :

| order_id | base_denom | quote_denom | side | remaining_quantity | price |
|----------|------------|-------------|------|--------------------|-------|
| order1   | AAA        | BBB         | sell | 10000000           | 0.375 |
| order2   | BBB        | AAA         | sell | 500000000          | 2.6   |

The inverse taker price (1/price) of order2 is greater than the price of order1 (~0.3846 > 0.375) hence orders match.
The order2 should be executed with the price of the order1 (taker gets the better price).

The `max_execution_quantity` of 10000000 with the 1/0.375 price is 9999750 (the remainder is 250).
The `opposite_execution_quantity` is 26666000.

As a result:

The order1 is filled partially with the exact price 9999750 BBB / 26666000 AAA = 0.375.
The order2 expected to sell 10000000 BBB, and receive 26000000 AAA, but received more 26666000 AAA and
additionally not filled 250 BBB.

#### 2-way matching

Coreum DEX supports 2-way matching for all placed orders. This means that it matches the order with the `self` order
book with the same base and quote denom (e.g. AAA/BBB) and with the opposite order book as well (e.g. BBB/AAA) to find
the best price to execute the order.

### Price tick and precision

To provide a better trading experience we define the [price_tick](https://www.investopedia.com/terms/t/tick.asp) for
each order book. The `price_tick` mostly depends on the price of the assets traded, that's why we can define the
variable for a token used to define the order book `price_tick`. This variable is named `unified_ref_amount`.
`unified_ref_amount`for token represents the amount of the token subunit you need to pay to buy 1 USD dollar. If the
token is issued on the Coreum chain, that variable can be set/updated by the token admin. If the token is IBC token,
or the token doesn't have and admin this variable can be set/updated by the chain governance. If the
`unified_ref_amount` is not set for a token, the `unified_ref_amount` is equal to 10^6.

The formula taken for the price tick is:

```
price_tick(base_denom/quote_denom) = 10^(floor(log10((unified_ref_amount(quote_denom) / unified_ref_amount(base_denom)))) + price_tick_exponent)
```

The `price_tick_exponent` is the coefficient used to give better price precision for the token orders. The default
`price_tick_exponent` is `-5`, and can be updated by the governance.

Tick size example:

| unified_ref_amount(AAA) | unified_ref_amount(BBB) | price_tick(AAA/BBB) | price_tick(BBB/AAA) |    
|-------------------------|-------------------------|---------------------|---------------------|
| 10000.0                 | 10000.0                 | 10^-5               | 10^-5               | 
| 3000.0                  | 20.0                    | 10^-8               | 10^-3               | 
| 3100000.0               | 8.0                     | 10^-11              | 1                   |
| 0.00017                 | 100.0                   | 1                   | 10^-11              |
| 0.000001                | 10000000                | 10^8                | 10^-18              |

The update of the `unified_ref_amount` doesn't affect the created orders.

### Balance locking

When a user places an order we lock the coins in the assetft (similar to freezing), both assetft and native coins.
At the time of the placement we enforce all assetft rules. For extensions, we expect a specific interface to be
implemented in the extensions smart contract, which let the DEX understand whether an order placement is allowed or not.
If such an interface is not implemented we don't allow the order placement.
If, at the time of matching, the assetft rules for the maker orders are changed, the orders will be still executed with
the amounts in the order book.


