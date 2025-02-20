# Binance ticks


| Market      | Price (in quote) | Amount (in base)    | Total (in quote) |
| ----------- | ---------------- | ------------------- | ---------------- |
| BTC/USDT    | 0.01             | 0.00001   (~0.9$)   | 10^-7            |
| ETH/USDT    | 0.01             | 0.0001    (~0.3$)   | 0.000001=10^-6   |
| ATOM/USDT   | 0.001            | 0.01      (~0.04$)  | 0.00001=10^-5    |
| TRX/USDT    | 0.0001           | 0.1       (~0.024$) | 0.00001=10^-5    |
| PEOPLE/USDT | 0.00001          | 0.1                 | 0.000001=10^-6   |
| YFI/USDT    | 1                | 0.00001             | 0.00001=10^-5    |
|             |                  |                     |                  |
|             |                  |                     |                  |


Optimal total IMO is 10^-6 in USDT

In binance BTC/USDT:

Price USDT for btc 95000,55 -> max 2 decimals. So min tick for price is 0.01

Amount (BTC) 0.00001 -> 10^-5. Which is equal to 0.95 USD

So in binance amount tick is equal to ~1$

While tick price is equal to 0.01$

Also min order value is 5$


# For Coreum DEX it should be:

Amount tick = ref_amount_a

price_tick = 10^(floor(log10((unified_ref_amount(quote_denom) / unified_ref_amount(base_denom)))) + price_tick_exponent) = 0.0001 - we want to have some extra tick comparing to binance.

## Simulation

ref_amount_BTC = 0.00001
ref_amount_USD = 1

market BTC/USD:
Amount_tick = 0.00001 BTC
price_tick = 10^ log10(1/0.00001) - 4 = 10^(5-4) = 10