# Coreum DEX Specification

## Overview

The Coreum DEX allows any Coreum user to create orders for trading any token pair. This means the order book is
bidirectional and permissionless, allowing for flexible and open trading.

## Order book

### Order Attributes

Users can place orders with the following attributes:

* `order_id` - unique order identifier of the order.
* `base_denom` - when you buy, you are buying the `base_denom`, when you sell, you are selling the `base_denom`.
* `quote_denom` - when you buy, you are selling the `quote_denom`, when you sell, you are buying the `quote_denom`.
* `price` - value of one unit of the `base_denom` expressed in terms of the `quote_denom`. It indicates how much of the
  `quote_denom` is needed to buy one unit of the `base_denom`.
* `quantity` - is amount of the `base_denom` being traded.
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

## Order placement and matching

Once an order is placed, the DEX attempts to match it with opposite orders in the same order book and with orders in the
corresponding inverse order book. The system ensures execution at the best available price, depending on the order type
and settings.

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

This method is implemented inside `x/dex`. The func name is `func computeMaxExecutionQuantity`

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

### 2-way matching

Coreum DEX uses a 2-way matching system for processing all placed orders. This approach matches orders within its own
order book (the self order book), where the base and quote denominations are identical (e.g., AAA/BBB). Additionally,
it simultaneously checks the corresponding opposite order book (e.g., BBB/AAA) to find the best possible execution
price. This 2-way order book matching ensures that orders are filled at the most favorable rates available,
whether within the same trading pair or its inverse, optimizing the trading experience.

### Price tick and precision

To provide a better trading experience we define the [price_tick](https://www.investopedia.com/terms/t/tick.asp) for
each order book. The `price_tick` mostly depends on the price of the assets traded, that's why we can define the
variable for a token used to define the order book `price_tick`. This variable is named `unified_ref_amount`.
`unified_ref_amount` for token represents the amount of the token subunit you need to pay to buy 1 USD dollar. If the
token is issued on the Coreum chain, that variable can be set/updated by the token admin. If the token is IBC token,
or the token doesn't have and admin this variable can be set/updated by the chain governance. If the
`unified_ref_amount` is not set for a token, the `unified_ref_amount` is equal to 10^6.

The formula taken for the price tick is:

```
price_tick(base_denom/quote_denom) = 10^(floor(log10((unified_ref_amount(quote_denom) / unified_ref_amount(base_denom)))) + price_tick_exponent)
```

The `price_tick_exponent` is the coefficient used to give better price precision for the token orders. The default
`price_tick_exponent` is `-8`, and can be updated by the governance.

Tick size example:

| unified_ref_amount(AAA) | unified_ref_amount(BBB) | price_tick(AAA/BBB) | price_tick(BBB/AAA) |    
|-------------------------|-------------------------|---------------------|---------------------|
| 10000.0                 | 10000.0                 | 10^-8               | 10^-8               | 
| 3000.0                  | 20.0                    | 10^-11              | 10^-6               | 
| 3100000.0               | 8.0                     | 10^-14              | 10^-3               |
| 0.00017                 | 100.0                   | 10^-3               | 10^-14              |
| 0.000001                | 10000000                | 10^5                | 10^-21              |

The update of the `unified_ref_amount` doesn't affect the created orders.

#### Price type

The `price` is custom type for the DEX module used to represent the price supporting extremely large or low values.
The `price` uses `{number}e{exponent}` format as a string representation of the price type. The min exponent is -100,
and the max exponent is 100. The max number of digits in number part 19. Zero price is prohibited.

Examples:

* A price like 45000000000000 can be represented as 45e12.
* A price like 0.000123 can be represented as 123e-6.
* A price like 70001 can be represented as is.
* Min value: 1e-100
* Max value: 9999999999999999999e100

The `price` string must be normalized, so you can't provide the same value using different strings. That's why the
`price` must match the regex: `^(([1-9])|([1-9]\d*[1-9]))(e-?[1-9]\d*)?$`.

Examples:

* 10 is invalid, must be 1e1
* 01 is invalid, must be 1
* 1e01 is invalid, must be 1e1
* 1e+1 is invalid, must be 1e1

### Balance locking/freezing/whitelisting/clawback.

When a user places an order we lock the coins in the assetft (similar to freezing). Also, we reserve the expected
receiving amount if whitelisting for the token the user expects to receive is enabled. At the time of the placement we
enforce all assetft rules. If, at the time of matching, the assetft rules for the maker orders are changed, the orders
will be still executed with the amounts in the order book. That's why to avoid unexpected behavior with the
`freezing/whitelisting/clawback` the token admin should [cancel](#Order-cancellation) user's orders before update the
rules.

### Time in force

The `time_in_force` setting specifies the execution conditions of an order, dictating how long it should remain active
or how it should be executed:

* `GTC (Good Til Canceled)`: The order remains active until it is fully executed or manually canceled by the trader.
  There is no time limit for the order to stay in the order book.

* `IOC (Immediate Or Cancel)`: The order must be executed immediately, either in full or partially. Any portion of the
  order that cannot be filled immediately is canceled.

* `FOK (Fill or Kill)`: The order must be executed immediately and completely. If the order cannot be filled in its
  entirety right away, it is canceled in full.

### Good til

The `good_til` setting specifies how long an order remains active based on certain conditions:

* `good_til_block_height`: The order remains active until a specific blockchain block height is reached. Once the
  blockchain reaches the specified block height, the order is automatically canceled if it hasn't been executed.

* `good_til_block_time`: The order stays active until a specified time, based on the blockchain’s timestamp. If the
  order is not executed by this time, it is automatically canceled.

### Order reserve

This feature introduces an order reserve requirement for each order placed on the chain. The reserve acts as a security
deposit that ensures users have a vested interest in their orders, helping to prevent spam and malicious activities.
Once the order is executed that reserve is released.
The default reserve amount is `10 CORE` and can be updated by the governance.

### Max orders limit

The number of active orders a user can have for each denom is limited by a value called `max_orders_per_denom`,
which is determined by DEX governance. The default value is 100.

### Events

The DEX module emits events at the time of the matching to notify the interested parties of the changes caused by the
matching. The events are defined in the [event.proto](../../../proto/coreum/dex/v1/event.proto). The events are emitted
for each order that was placed. The events are:

1. `EventOrderPlaced` is emitted when the order is placed.
2. `EventOrderReduced` is emitted when the order is reduced during the matching.
3. `EventOrderClosed` is emitted when the order is closed during the matching or manually, or because of `good_til` in
   the `begin blocker`, and removed from the order book.
4. `EventOrderCreated` is emitted when the order is saved to the order book.

## Asset FT and DEX

### Unified ref amount

The `unified_ref_amount` used in [price tick and precision](#price-tick-and-precision) can be updated and by the gov
or the token admin. Check the [price tick and precision](#price-tick-and-precision) to get the details.

### Order cancellation

The users can cancel their orders within the DEX. Or The token admin or gov can cancel user orders within the system.
It grants specific administrative or governance roles the power to manage and oversee active orders, providing a
safeguard against potential issues such as erroneous trades, malicious activity, or market manipulation. For the token
admin to cancel the user's orders, the `dex_order_cancellation` feature must be enabled.

### Block DEX

The asset FT token admin at time of issues can set the `block_dex` feature, which will block the DEX completely for the
issued denom. So the DEX won't accept the orders with the denom.

### Global freeze

If the `global_freeze` feature is enabled for either the `base_denom` or `quote_denom` of an order, it is prohibited
to place an order with the corresponding `base_denom` and `quote_denom`.

### Denoms to trade with

This feature introduces an enhancement to the asset FT (`restrict_dex` feature), allowing the asset FT
(a specific fungible token) to be traded only against a predefined set of denoms specified in `denoms_to_trade_with`.
This adds a layer of control over trading pairs, ensuring that denom(asset FT) can only be exchanged with certain
denoms/currencies or assets, as specified by the admin.

### Extension

The `extension` feature integrates the DEX with the smart contract extension capability of asset FT. This integration
enables asset FT extension contracts to define custom extension functions that the DEX invokes before executing an
order. The extension functions can leverage order details, including the amounts expected to be spent and received
post-execution, to implement custom rules. For example, they can validate an order against specific business logic or
constraints.

#### Asset FT features handling

If the extension feature is enabled for a token but no contract extension function is implemented, all order
placements will fail. When multiple asset FT features are enabled, the rules of the asset FT features are validated
first, followed by invoking the extension.

#### Example of WASM Call Implementation in an Extension Smart Contract

Below is an example showcasing how to implement the extension functionality in a smart contract:

##### Code Example

```rust
#[cw_serde]
pub struct DEXOrder {
    pub creator: String,
    #[serde(rename = "type")]
    pub order_type: String,
    pub id: String,
    pub base_denom: String,
    pub quote_denom: String,
    pub price: Option<String>,
    pub quantity: Uint128,
    pub side: String,
}

pub enum SudoMsg {
    ExtensionTransfer {
        recipient: String,
        sender: String,
        transfer_amount: Uint128,
        commission_amount: Uint128,
        burn_amount: Uint128,
        context: TransferContext,
    },
    ExtensionPlaceOrder {
        order: DEXOrder,
        expected_to_spend: Coin,
        expected_to_receive: Coin,
    },
}

#[entry_point]
pub fn sudo(deps: DepsMut, env: Env, msg: SudoMsg) -> CoreumResult<ContractError> {
    match msg {
        SudoMsg::ExtensionTransfer {
            recipient,
            sender,
            transfer_amount,
            commission_amount,
            burn_amount,
            context,
        } => sudo_extension_transfer(
            deps,
            env,
            recipient,
            sender,
            transfer_amount,
            commission_amount,
            burn_amount,
            context,
        ),
        SudoMsg::ExtensionPlaceOrder {
            order,
            expected_to_spend,
            expected_to_receive,
        } => sudo_extension_place_order(
            deps,
            env,
            order,
            expected_to_spend,
            expected_to_receive,
        ),
    }
}
```

**`SudoMsg` Enum**:

- Defines the types of messages the extension supports, such as `ExtensionTransfer` and `ExtensionPlaceOrder`.
- Each variant carries the necessary information to execute specific logic.

**`sudo` Entry Point**:

- Acts as the main function that handles incoming messages.
- Delegates to specific helper functions based on the message type.

