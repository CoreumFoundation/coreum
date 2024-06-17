#### Price tick and precision

To provide a better trading experience we define the [price_tick](https://www.investopedia.com/terms/t/tick.asp) for
each pair. The `price_tick` mostly depends on the price of the assets traded, that's why we can define the variable for
a token used to define the pair `price_tick`. This variable is named `unified_ref_price`. `unified_ref_price` for token
represents the amount of the token subunit you need to pay to buy 1 USD dollar. If the token is issued on the Coreum
chain, that variable can be set/updated by the token admin. If the token is IBC token, or the token doesn't have and
admin this variable can be set/updated by the chain governance. If the `unified_ref_price` is not set for a token, the
`unified_ref_price` is equal to 1. 

The formula taken for the price tick is:

```
price_tick(base_denom/quote_denom) = round_to_power_of_ten(price_tick_multiplier * unified_ref_price(quote_denom) / unified_ref_price(base_denom))
```

The `price_tick_multiplier` is the coefficient used to give better price precision for the token pairs. The default
`price_tick_multiplier` is `10^-5`, and can be updated by the governance.

The `round_to_power_of_ten` is the function that finds nearest power of ten value to define the tick. For the rounding
it uses half round up rounding.

Tick size example:

| unified_ref_price(AAA) | unified_ref_price(BBB) | price_tick(AAA/BBB) | price_tick(BBB/AAA) |    
|------------------------|------------------------|---------------------|---------------------|
| 10000.0                | 10000.0                | 10^-5               | 10^-5               | 
| 3000.0                 | 20.0                   | 10^-7               | 10^-3               | 
| 3100000.0              | 8.0                    | 10^-11              | 1                   |
| 0.00017                | 100.0                  | 10                  | 10^-11              |

The update of the `unified_ref_price` doesn't affect the created orders. 

The Coreum DEX price supports up to 36 decimals for the price precision. Which means that max price tick is equal to
the max supported price precision. If the price tick of two traded coins exceeds 36 decimals we use `10^-36` as the 
price tick for such pair.