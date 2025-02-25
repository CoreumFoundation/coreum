# Binance ticks


| Market      | Price (in quote) | Amount (in base)          | Price & ATH(base) | Total (in quote) |
| ----------- | ---------------- | ------------------------- | ----------------- | ---------------- |
| BTC/USDT    | 0.01             | 0.00001=10^-5   (~0.95$)  | 95000 & 110000    | 10^-7            |
| ETH/USDT    | 0.01             | 0.0001=10^-4    (~0.3$)   | 3000  & 4800      | 0.000001=10^-6   |
| ATOM/USDT   | 0.001            | 0.01=10^-2      (~0.04$)  | 4.5   & 40        | 0.00001=10^-5    |
| TRX/USDT    | 0.0001           | 0.1=10^-1       (~0.024$) | 0.24  & 0.45      | 0.00001=10^-5    |
| PEOPLE/USDT | 0.00001          | 0.1=10^-1       (~0.002$) | 0.019  & 0.20     | 0.000001=10^-6   |
| YFI/USDT    | 1                | 0.00001=10^-5   (~0.06$)  | 6000  & 90000     | 0.00001=10^-5    |
| SOL/USDT    | 0.01             | 0.001=10^-3     (~0.15$)  | 150   & 280       | 0.00001=10^-5    |
| TON/USDT    | 0.001            | 0.01=10^-2      (~0.035$) | 3.5   & 8.5       | 0.0001=10^-4     |
| PEPE/USDT   | 0.00000001       | 1               (low)     | 0.00000873 & 3X   | 10^-8            |


ATH price to amount tick rule:

| ATH price     | Amount tick | Markets |
| ------------- | ----------- |---------
| <0.1$         | 1           | PEPE
| 0.1-1$        | 0.1         | TRX,PEOPLE
| 1-10$         | 0.01        | TON
| 10-100$       | 0.01        | ATOM
| 100-1000$     | 0.001       | SOL
| 1000-10000$   | 0.0001      | ETH
| 10000-100000$ | 0.00001     | BTC,YFI

So it seems that reasonable value for Base Amount tick should be equal to 0.01$ = `unified_ref_amount * 0.01`
Quote amount tick should be equal to 10^-6$ = `unified_ref_amount * 10^-6`

quote_amount = base_amount * price => price = quote_amount / base_amount, so 

`price_tick = unified_ref_amount_quote * 10^-6 / unified_ref_amount_base * 10^-2 =`
`10^-4 * unified_ref_amount_quote / unified_ref_amount_base`

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