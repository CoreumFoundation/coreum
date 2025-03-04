# Price Tick Size  

To provide a better trading experience, avoid rounding issues, and minimize remainders during order execution, we use [`price_tick_size`](https://www.investopedia.com/terms/t/tick.asp). A tick is the minimum price movement an asset can make, either upward or downward.

Since tick size is set by the exchange where an asset is traded and is primarily based on its price (though it also depends on asset type and market conditions), we can derive a formula to calculate `price_tick_size` for any market. This calculation is based on the relationship of both assets against a common instrument. For example, to determine the `price_tick_size` for the **ETH/BTC** pair, we can use the prices of **ETH/USD** and **BTC/USD**.  

To define an asset’s price on-chain, we introduce a parameter called `unified_ref_amount`. This represents the amount of the token’s **subunit** required to equal **1 USD**. For example, if BTC is priced at **$90,000**, then `unified_ref_amount` should be set to **0.0000111** (since **1 / 90,000 = 0.0000111**).  

## How `unified_ref_amount` is Defined  

1. **Coreum Native Assets**: If the token is issued on the Coreum chain, this variable can be set or updated by the token admin.  
2. **IBC Tokens & Admin-less Tokens**: If the token is an IBC token or does not have an admin, this variable can be set or updated through chain governance.  
3. **Default Value**: If `unified_ref_amount` is not explicitly set for a token, a default value of **10^6** is used.  

## Formula  

Let's say we have two assets: **AAA** and **BBB**.  
To calculate `price_tick_size` for the **AAA/BBB** market, we use the following formula:  

```
price_tick_size(AAA/BBB) = 10^price_tick_exponent * round_pow10(unified_ref_amount(AAA)/unified_ref_amount(BBB))
```

Where:  
- `price_tick_exponent` is a coefficient that controls the price precision for a market. The current value of  `price_tick_exponent` is `-6`, but it can be changed through governance.
- `round_pow10(x)` rounds `x` to the **closest power of 10** (positive or negative).  

### `round_pow10` Behavior:  
| Input | Output |
| ----- | ------ |
| 0.333 | 0.1    |
| 0.55  | 1.0    |
| 0.5   | 1.0    |
| 0.499 | 0.1    |

For more details on the logic behind this formula and the constants used, refer to the [Research](#research) section.  

# Base Amount Step  

To ensure smooth order execution and prevent excessively small trade sizes, we introduce the `base_amount_step`. This defines the smallest allowable step for the base asset inside a market. It prevents rounding issues, partial order cancellations, and improves consistency across markets. Unlike `price_tick_size`, which is defined per market, `base_amount_step` is defined per asset.  

## Formula

The **Base Amount Step** for a given asset **AAA** is calculated using the following formula:  

```
base_amount_step(AAA) = 10^base_amount_exponent * round_pow10(unified_ref_amount(AAA))
```

Where:  
- `base_amount_exponent` is a coefficient that controls the granularity of the base asset amount step.  
  The current value of `base_amount_exponent` is `-2`, but it can be changed through governance.  
- `round_pow10(unified_ref_amount(AAA))` ensures that the step size aligns with the asset’s magnitude.  

## Example Calculations  

| Asset Price (USD) | `unified_ref_amount(AAA)` | `base_amount_step` |
| ----------------- | ------------------------- | ------------------ |
| $100              | 0.01                      | 0.0001             |
| $5,000            | 0.0002                    | 0.000001           |
| $0.05             | 20                        | 0.1                |

This approach ensures that minimum trade sizes scale appropriately with asset value while maintaining a consistent precision level.  

For more details on the logic behind this formula and the constants used, refer to the [Research](#research) section.  

## Important

**Changes to `unified_ref_amount`, `price_tick_exponent` or `base_amount_exponent` **do not** affect existing orders.**


# Research

## CEX ticks

| Market      | Price (in quote)  | Amount (in base)          | Price & ATH(base) | Total (in quote) |
| ----------- | ----------------- | ------------------------- | ----------------- | ---------------- |
|             |                   |                           |                   |                  |
| Binance     |                   |                           |                   |                  |
| BTC/USDT    | 0.01              | 0.00001=10^-5   (~0.95$)  | 95000 & 110000    | 10^-7            |
| ETH/USDT    | 0.01              | 0.0001=10^-4    (~0.3$)   | 3000  & 4800      | 0.000001=10^-6   |
| ATOM/USDT   | 0.001             | 0.01=10^-2      (~0.04$)  | 4.5   & 40        | 0.00001=10^-5    |
| TRX/USDT    | 0.0001            | 0.1=10^-1       (~0.024$) | 0.24  & 0.45      | 0.00001=10^-5    |
| PEOPLE/USDT | 0.00001           | 0.1=10^-1       (~0.002$) | 0.019  & 0.20     | 0.000001=10^-6   |
| YFI/USDT    | 1                 | 0.00001=10^-5   (~0.06$)  | 6000  & 90000     | 0.00001=10^-5    |
| SOL/USDT    | 0.01              | 0.001=10^-3     (~0.15$)  | 150   & 280       | 0.00001=10^-5    |
| TON/USDT    | 0.001             | 0.01=10^-2      (~0.035$) | 3.5   & 8.5       | 0.0001=10^-5     |
| PEPE/USDT   | 0.00000001=10^-8  | 1               (~0)      | 0.00000873 & 3X   | 10^-8            |
|             |                   |                           |                   |                  |
| OKX         |                   |                           |                   |                  |
| BTC/USDT    | 0.1               | 0.00000001=10^-8          | 95000 & 110000    | 10^-9            |
| ETH/USDT    | 0.01              | 0.000001=10^-6            | 3000  & 4800      | 10^-8            |
| TON/USDT    | 0.001             | 0.0001=10^-4              | 3.5   & 8.5       | 10^-7            |
| PEPE/USDT   | 0.000000001=10^-9 | 1                         | 0.00000873 & 3X   | 10^-9            |
|             |                   |                           |                   |                  |
| ByBit       |                   |                           |                   |                  |
| BTC/USDT    | 0.01              | 0.000001=10^-6            | 95000 & 110000    | 10^-8            |
| ETH/USDT    | 0.01              | 0.00001=10^-5             | 3000  & 4800      | 10^-7            |
| TON/USDT    | 0.001             | 0.01=10^-2                | 3.5   & 8.5       | 10^-5            |
| PEPE/USDT   | 0.00000001=10^-8  | 1                         | 0.00000873 & 3X   | 10^-8            |
|             |                   |                           |                   |                  |
| HyperLiquid |                   |                           |                   |                  |
| BTC/USDT    | 1                 | 0.00001=10^-5             | 95000 & 110000    | 10^-5            |
| ETH/USDT    | 0.1               | 0.0001=10^-4              | 3000  & 4800      | 10^-5            |

### Avg price to amount tick rule (Binance):

| Avg price     | Amount tick | Markets    |
| ------------- | ----------- | ---------- |
| <0.1$         | 1           | PEPE       |
| 0.1-1$        | 0.1         | TRX,PEOPLE |
| 1-10$         | 0.01        | TON        |
| 10-100$       | 0.01        | ATOM       |
| 100-1000$     | 0.001       | SOL        |
| 1000-10000$   | 0.0001      | ETH        |
| 10000-100000$ | 0.00001     | BTC,YFI    |

### TODO: Discuss gap here.

### Non USDT pairs:

| Market  | Price (in quote) | Amount (in base)       | Total (in quote) |
| ------- | ---------------- | ---------------------- | ---------------- |
| Binance |                  |                        |                  |
| ETH/BTC | 0.00001=10^-5    | 0.0001=10^-3   (~0.3$) | 10^-8            |
|         |                  |                        |                  |
| OKX     |                  |                        |                  |
| ETH/BTC | 0.00001=10^-5    | 0.000001=10^-6         | 10^-11           |
|         |                  |                        |                  |
| ByBit   |                  |                        |                  |
| ETH/BTC | 0.000001=10^-6   | 0.00001=10^-5          | 10^-11           |

## Coreum ticks proposal

Proposed Base Amount tick range in USD   `[0.01,0.1]`    => `round_up_pow10(ura_base * 0.01)=0.01*round_up_pow10(ura_base)`
Proposed Quote Amount tick range in USD  `[10^-8,10^-7]` => `round_up_pow10(ura_quote * 10^-8)=10^-8*round_up_pow10(ura_quote)`

However we don't use Quote Amount tick directly, we use price tick instead.
`quote_tick = base_tick * price_tick` => `price_tick = quote_tick / base_tick`
`price_tick = 10^-8*round_up_pow10(ura_quote) / 0.01*round_up_pow10(ura_base)`

To decrease rounding we can do it after all calculations, so
`price_tick = 10^-6 * round_up_pow10(ura_quote/ura_base)=10^-6 * round_up_pow10(ura_quote/ura_base)`

## Coreum ticks simulation

| Base      | Quote | ura_base | ura_quote | amount_tick                     | price_tick                                    |
| --------- | ----- | -------- | --------- | ------------------------------- | --------------------------------------------- |
| BTC       | USDT  | 0.000011 | 1.0       | 0.01*(0.000011)=~0.000001=10^-6 | `10^-6*(1.0/0.000011)=0.1`                    |
| ETH       | USDT  | 0.000333 | 1.0       | 0.01*(0.000333)=~0.00001=10^-5  | `10^-6*(1.0/0.000333)=0.01`                   |
| TRX       | USDT  | 4.5      | 1.0       | 0.01*(4.5)=~0.1=10^-1           | `10^-6*(1.0/4.5)=0.000001`                    |
| PEPE(c.p) | USDT  | 80000    | 1.0       | 0.01*(80000)=1000               | `10^-6*(1.0/80000)=10^-6*0.0001=10^-10`       |
| Non-USDT  |       |          |           |                                 |                                               |
| ETH       | BTC   | 0.000333 | 0.000011  | 0.01*(0.000333)=~0.00001=10^-5  | `10^-6*(0.000011/0.000333)=10^-6*0.1=0.00001` |

### TODO: Discuss:
- potential issue caused by amount_tick for PEPE. We can change formula to use min(amount_tick,1) but it causes quote_tick to not be respected and more rounding
- another approach we can use is to hardcode ticks, discuss gap in `Avg price to amount tick rule (Binance)`
- to be implemented: allow setting of price tick per pair
- potentially discrepancy is because CEXes don't want to decrease tick (e.g for price)
### TODO: discuss:
- discuss ranges. Ranges define which % of price change we consider significant. We can potentially increase 0.01 & 10^-8
- which method of rounding we want to use (round up, round down, ceil, floor,bankers' rounding etc)

References:
- https://www.investopedia.com/terms/t/tick.asp
- https://www.binance.us/trade-limits