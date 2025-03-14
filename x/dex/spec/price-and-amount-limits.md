# Problem

In our system, all balances are stored and operated as integers. This means that during order matching and execution, all calculations must be performed using integer values. This approach is not unique to our system and centralized exchanges (CEXes), traditional finance exchanges also use integers as the underlying type for calculations to ensure precise execution and avoid rounding errors.

If we were to rely solely on this requirement, order creation would only need to follow two simple rules:
- `base_quantity` must be an integer.
- `quote_quantity = price * base_quantity` must also be an integer.

However, this introduces a significant problem: the granularity of price and base amount become dependent on each other.
- A **larger `base_quantity`** allows specifying a more precise price.
- A **higher price** enables a more precise `base_quantity`.

This relationship can be illustrated as follows:

| base_quantity | min price | quote_quantity |
| ------------- | --------- | -------------- |
| 1             | 1         | 1              |
| 10            | 0.1       | 1              |
| 100           | 0.01      | 1              |
| ...           | ...       | ...            |
| 1000000       | 0.000001  | 1              |

This means that price precision can increase infinitely simply by increasing `base_quantity`, leading to inconsistencies and execution issues.

Consider the following two orders (assuming that only integer amounts are allowed for simplicity):

- **Order 1:** Sell **1,000,000 BTC** at a price of **90,000.111111 USDT** per BTC. (The quote amount is an integer.)
- **Order 2:** Buy **0.1 BTC** at a price of **91,000 USDT** per BTC. (The quote amount is an integer.)

For Order 2 to be valid, the buyer must pay:
`0.1 BTC * 90,000.111111 = 9,000.0111111 USDT`

Since balances are stored as integers, this fractional amount cannot be accurately represented. The mismatch between price granularity and the required integer representation makes it impossible to execute certain trades consistently.

### Solution: Introducing `price_tick_size`, `quantity_step`, and `quote_quantity_step`
To resolve these issues, we introduce three parameters: `price_tick_size`, `quantity_step`, and `quote_quantity_step`. However, only two of these are required to determine the third.

- **`price_tick_size`** ensures that prices have a fixed minimum increment, preventing arbitrary fractional values.
- **`quantity_step`** enforces a minimum step size for base asset quantities, ensuring consistent calculations.
- **`quote_quantity_step`** provides a consistent step size for quote amounts, derived from the other two parameters.

By enforcing these constraints, we eliminate precision mismatches, simplify order execution, and align with standard practices.

# Price Tick

To provide a better trading experience, avoid rounding issues, and minimize remainders during order execution, we use [`price_tick`](https://www.investopedia.com/terms/t/tick.asp). A tick is the minimum price movement an asset price can make, either upward or downward.

Since tick size is set by the exchange where an asset is traded and is primarily based on its price (though it also depends on asset type and market conditions), we can derive a formula to calculate `price_tick` for any market. This calculation is based on the relationship of both assets against a common instrument. For example, to determine the `price_tick` for the **ETH/BTC** pair, we can use the prices of **ETH/USD** and **BTC/USD**.

To define an asset’s price on-chain, we introduce a parameter called `unified_ref_amount`. This represents the quantity of the token’s **subunit** that corresponds to **1 USD**.

For instance, if BTC is issued on Coreum with satoshi as its subunit (where **1 BTC = 100,000,000 satoshis**) and its market price is **$90,000**, then `unified_ref_amount` should be **0.0000111 BTC (or 1110 satoshis)**, since **1 BTC / 90,000 = 0.0000111 BTC**, which approximates **1 USD** in satoshi terms.


## How `unified_ref_amount` is Defined

1. **Coreum Native Assets**: If the token is issued on the Coreum chain, this variable can be set or updated by the token admin.
2. **IBC Tokens & Admin-less Tokens**: If the token is an IBC token or does not have an admin, this variable can be set or updated through chain governance.
3. **Default Value**: If `unified_ref_amount` is not explicitly set for a token, a default value of **10^6** is used.

## Formula

Let's say we have two assets: **AAA** and **BBB**.
To calculate `price_tick` for the **AAA/BBB** market, we use the following formula:

```
price_tick(AAA/BBB) = 10^(price_tick_exponent + ceil(log10(unified_ref_amount(BBB)/unified_ref_amount(AAA))))
```

Where:
- `price_tick_exponent` is a coefficient that controls the price precision for a market. The current value of  `price_tick_exponent` is `-6`, but it can be changed through governance.
- `ceil(log10(unified_ref_amount(BBB)/unified_ref_amount(AAA)))` ensures that price step size takes into account both asset magnitude

For more details on the logic behind this formula and the constants used, refer to the [Theoretical Justification](#Theoretical-Justification) section.

# Quantity Step (Base Quantity Step)

To ensure smooth order execution and prevent excessively small trade sizes, we introduce the `quantity_step`. This defines the smallest allowable step for the **base** asset inside a market. It prevents rounding issues, partial order cancellations during execution, and improves consistency across markets. Unlike `price_tick`, which is defined per market, `quantity_step` is defined per asset.

## Formula

The **Quantity Step** for a given asset **AAA** is calculated using the following formula:

// TODO: Quantity step can't be less than 1 (at least)
```
quantity_step(AAA) = max(1, 10^(quantity_step_exponent + ceil(log10(unified_ref_amount(AAA)))))
```

Where:
- `quantity_step_exponent` is a coefficient that controls the granularity of the quantity of a base asset in market.
  The current value of `quantity_step_exponent` is `-2`, but it can be changed through governance.
- `ceil(log10(unified_ref_amount(AAA)))` ensures that the step size aligns with the asset’s magnitude.

This approach ensures that minimum trade sizes scale appropriately with asset value while maintaining a consistent precision level.

For more details on the logic behind this formula and the constants used, refer to the [Theoretical Justification](#Theoretical-Justification) section.

## **Important**

- Changes to `unified_ref_amount`, `price_tick_exponent` or `quantity_step_exponent` **do not** affect existing orders.
- `price_tick` and `quantity_step` represent hard backend boundaries and may have more granular precision than end users need. Depending on the use case, applications can use less granular values to provide a better user experience.
- `quantity_step` could be greater than 1 in some cases, and front-end applications should handle this properly. For example, a market may be represented where the quoted price is for a specific multiple of the base asset (e.g., **kPEPE/USDT** instead of **PEPE/USDT**, kPEPE means 1000PEPE, [reference](https://app.hyperliquid.xyz/trade/kPEPE)). This improves readability and prevents excessive decimal precision.

# Theoretical Justification

## Other Exchange Ticks

Below is a table containing trading configuration details for Binance across different markets.

| Market      | Price Tick Size    | Base Quantity Step        | Price & ATH (base) | Quote Quantity Step |
| ----------- | ------------------ | ------------------------- | ------------------ | ------------------- |
| BTC/USDT    | 0.01               | 0.00001 = 10^-5 (~0.95$)  | 95000 & 110000     | 10^-7               |
| ETH/USDT    | 0.01               | 0.0001 = 10^-4  (~0.3$)   | 3000  & 4800       | 10^-6               |
| ATOM/USDT   | 0.001              | 0.01 = 10^-2    (~0.04$)  | 4.5   & 40         | 10^-5               |
| TRX/USDT    | 0.0001             | 0.1 = 10^-1     (~0.024$) | 0.24  & 0.45       | 10^-5               |
| PEOPLE/USDT | 0.00001            | 0.1 = 10^-1     (~0.002$) | 0.019  & 0.20      | 10^-6               |
| YFI/USDT    | 1                  | 0.00001 = 10^-5 (~0.06$)  | 6000  & 90000      | 10^-5               |
| SOL/USDT    | 0.01               | 0.001 = 10^-3   (~0.15$)  | 150   & 280        | 10^-5               |
| TON/USDT    | 0.001              | 0.01 = 10^-2    (~0.035$) | 3.5   & 8.5        | 10^-5               |
| PEPE/USDT   | 0.00000001 = 10^-8 | 1               (~0)      | 0.00000873 & 3X    | 10^-8               |

It appears that Base Quantity Step is highly correlated with asset price ranges. Below is a table showing the relationship between the average price and Base Quantity Step.

| Avg Price       | Base Quantity Step | Markets     |
| --------------- | ------------------ | ----------- |
| <0.1$           | 1                  | PEPE        |
| 0.1 - 1$        | 0.1                | TRX, PEOPLE |
| 1 - 10$         | 0.01               | TON         |
| 10 - 100$       | 0.01               | ATOM        |
| 100 - 1000$     | 0.001              | SOL         |
| 1000 - 10000$   | 0.0001             | ETH         |
| 10000 - 100000$ | 0.00001            | BTC, YFI    |

As seen above, `quantity_step` tends to be **inversely proportional** to the asset’s price, with some exceptions.

### **Comparison Across Exchanges**

To further analyze how these values vary between exchanges, we present a cross-platform comparison:

| Market      | Price Tick Size     | Base Quantity Step | Quote Quantity Step |
| ----------- | ------------------- | ------------------ | ------------------- |
| Binance     |                     |                    |                     |
| BTC/USDT    | 0.01                | 0.00001 = 10^-5    | 10^-7               |
| ETH/USDT    | 0.01                | 0.0001 = 10^-4     | 10^-6               |
| TON/USDT    | 0.001               | 0.01 = 10^-2       | 10^-5               |
| PEPE/USDT   | 0.00000001 = 10^-8  | 1                  | 10^-8               |
|             |                     |                    |                     |
| OKX         |                     |                    |                     |
| BTC/USDT    | 0.1                 | 0.00000001 = 10^-8 | 10^-9               |
| ETH/USDT    | 0.01                | 0.000001 = 10^-6   | 10^-8               |
| TON/USDT    | 0.001               | 0.0001 = 10^-4     | 10^-7               |
| PEPE/USDT   | 0.000000001 = 10^-9 | 1                  | 10^-9               |
|             |                     |                    |                     |
| ByBit       |                     |                    |                     |
| BTC/USDT    | 0.01                | 0.000001 = 10^-6   | 10^-8               |
| ETH/USDT    | 0.01                | 0.00001 = 10^-5    | 10^-7               |
| TON/USDT    | 0.001               | 0.01 = 10^-2       | 10^-5               |
| PEPE/USDT   | 0.00000001 = 10^-8  | 1                  | 10^-8               |
|             |                     |                    |                     |
| HyperLiquid |                     |                    |                     |
| BTC/USDT    | 0.1                 | 0.00001 = 10^-5    | 10^-6               |
| ETH/USDT    | 0.01                | 0.0001 = 10^-4     | 10^-6               |
| TON/USDT    | 0.00001             | 0.1 = 10^-1        | 10^-6               |
| kPEPE/USDT  | 0.000001            | 1 = 10^0           | 10^-6               |

### **Key Observations**
- There is **no universal standard** for Price Tick and Base Quantity across exchanges.
- Base Quantity Step typically falls between **0.01$ and 1$**, except in rare cases.
- Quote Quantity Step generally ranges from **10^-9 to 10^-6**, depending on market demand and exchange rules.
- For HyperLiquid Min Total is always equal to 10^-6, which is similar to approach we want to have

## Breaking Down the Math

Based on values we see in other exchanges, the following values are proposed:

- **Base Quantity Step** should be in the range: `0.01$ ≤ base_quantity_step ≤ 0.1$`
- **Quote Quantity Step** should be in the range: `10^-8$ ≤ quote_quantity_step ≤ 10^-7$`

### Quantity Step

Since `unified_ref_amount` is equal to `1$`, we should multiply it by a value between `0.01` and `0.1` to conform to the specified range. Additionally, we want our ticks to be powers of 10 either positive or zero e.g 1,10,100,1000...

The formula that respects all rules: `base_quantity_step(AAA) = max(1, 10^base_quantity_step_exponent * round_up_pow10(unified_ref_amount(AAA)))`

Given that 1$ <= `round_up_pow10(unified_ref_amount(AAA))` <= 10$ to achieve desired range we set `10^base_quantity_step_exponent = 0.01 → base_quantity_step_exponent = -2`.

Refinements:
1. Instead of `round_up_pow10(AAA)` use `10^ceil(log10(AAA))`
2. Shorten `base_quantity_step_exponent` to `quantity_step_exponent` since `quote_quantity_step_exponent` one is not used

Final formula:
```
quantity_step = max(1, 10^(quantity_step_exponent + ceil(log10(unified_ref_amount))))
```

### Price Tick

Since it is more convenient for users to work with `price_tick` instead of `quote_quantity_step`, we can derive `price_tick` using:

`quote_quantity=base_quantity*price -> quote_tick = base_tick * price_tick -> price_tick = quote_tick / base_tick`

Substituting our definitions:
`price_tick=10^base_quantity_step_exponent * round_up_pow10(unified_ref_amount(BBB)) / 10^quote_quantity_exponent * round_up_pow10(unified_ref_amount(AAA)) = 10^(base_quantity_step_exponent-quote_quantity_step_exponent) * round_up_pow10(unified_ref_amount(BBB) / round_up_pow10(unified_ref_amount(AAA)`

Refinements:
1. Minimize error → Perform rounding once after division.
2. Instead of `round_up_pow10` use `10^ceil(log10(AAA))`
3. Introduce `price_tick_exponent` as a new parameter. `price_tick_exponent = base_quantity_exponent - quote_quantity_exponent`


Final formula:
```
price_tick = 10^(price_tick_exponent + ceil(log10(unified_ref_amount(BBB)/unified_ref_amount(AAA))))
```

## Examples

| Base     | Quote | ura_base | ura_quote | quantity_step                       | price_tick                                    |
| -------- | ----- | -------- | --------- | ----------------------------------- | --------------------------------------------- |
| BTC      | USDT  | 0.000011 | 1.0       | 10^(-2+ceil(log10(0.000011)))=10^-6 | 10^(-6+ceil(log10(1/0.000011)))=10^-1=0.1     |
| ETH      | USDT  | 0.000333 | 1.0       | 10^(-2+ceil(log10(0.000333)))=10^-5 | 10^(-6+ceil(log10(1/0.000333)))=10^-2=0.01    |
| TRX      | USDT  | 4.5      | 1.0       | 10^(-2+ceil(log10(4.5)))=10^-1      | 10^(-6+ceil(log10(1/4.5)))=0.000001           |
| PEPE     | USDT  | 80000    | 1.0       | 10^(-2+ceil(log10(80000)))=10^3     | 10^-6*round_up_pow10(1.0/80000)=10^-10        |
| Non-USDT |       |          |           |                                     |                                               |
| ETH      | BTC   | 0.000333 | 0.000011  | 0.01*(0.000333)=~0.00001=10^-5      | 10^-6*round_up_pow10(0.000011/0.000333)=10^-7 |

As seen, these values largely align with or extend beyond the ranges observed on other exchanges.

### Comparison with Other Exchanges for Non-USDT Pairs

To validate the proposed tick sizes, let's compare with other exchanges.

| Exchange   | Market  | Price Tick Size     | Base Quantity Step | Quote Quantity Step |
| ---------- | ------- | ------------------- | ------------------ | ------------------- |
| Binance    | ETH/BTC | `0.00001 = 10^-5`   | `0.0001 = 10^-3`   | `10^-8`             |
| OKX        | ETH/BTC | `0.00001 = 10^-5`   | `0.000001 = 10^-6` | `10^-11`            |
| ByBit      | ETH/BTC | `0.000001 = 10^-6`  | `0.00001 = 10^-5`  | `10^-11`            |
| Coreum DEX | ETH/BTC | `0.0000001 = 10^-7` | `0.00001 = 10^-5`  | `10^-12`            |

## References:
- [Investopedia: Price Tick](https://www.investopedia.com/terms/t/tick.asp)
- [Binance Trading Limits](https://www.binance.us/trade-limits)