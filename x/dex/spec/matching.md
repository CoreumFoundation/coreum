# DEX

The spec describes the coreum DEX specification.

## Order book

The Coreum DEX orders can be created for any-to-any token by any Coreum user. Which meas that the order book
is bidirectional and permissionless. The user place an order and provides the order attributes:

* `order_id` - unique order identifier per account and order book
* `base_denom` - when you buy, you are buying the `base_denom`, when you sell, you are selling the `base_denom`.
* `quote_denom` - when you buy, you are selling the `quote_denom`, when you sell, you are buying the `quote_denom`.
* `price` - value of one unit of the `base_denom` expressed in terms of the `quote_denom`. It indicates how much of the
  `quote_denom` is needed to buy one unit of the `base_denom`.
* `quantity` - is amount of the base `base_denom` being traded.
* `side`
    * `sell` - means that the order is to sell `base_denom` `quantity` with the `price`.
    * `buy` - means that the order is to buy `base_denom` `quantity` with the `price`.

Once an order is placed the DEX will try to match the order with the same order book opposite side orders and opposite
order book same side orders. And depending on order type and settings execute it.

## Matching

### Rounding issue

Let's say we have 2 orders from the opposite order books:

| order_id | base_denom | quote_denom | quantity | price | side |
|----------|------------|-------------|----------|-------|------|
| order1   | AAA        | BBB         | 50000000 | 0.375 | sell |
| order2   | BBB        | AAA         | 10000000 | 2.6   | sell |

The inverse taker price (1/price) of order2 is greater than the price of order1 (~0.3846 > 0.375)hence orders match.
The order2 should be executed with the price of the order1 (taker gets the better price). The exact amount of AAA order2
receives is 10000000 * 1/0.375 = 26_666_666.(6) AAA. The amount we can use must be an integer, so we can't send full
amount without the price violation. The following solution resolves that issue.

### Tick size

To avoid significant filling quantity violation and provide a better trading experience we define
the [tick_size](https://www.investopedia.com/terms/t/tick.asp) for each pair.  `tick_size` mostly depends on the price
of the assets traded, that's why we can define the variable for a token used to define the pair `tick_sizes`. This
variable is named `significant_amount`. `significant_amount` for token represents minimum valuable amount of a token.
The recommended targeting price for the significant amount is 1 USD dollar cent.

The formula for tick size is:

```
tick_size(base_denom/quote_denom) = tick_size_multiplier * significant_amount(quote_denom) / significant_amount(base_denom)
```

The `tick_size_multiplier` is the coefficient used to give better price precision for the token pairs. The default value
of the `tick_size_multiplier` is  `0.01`, and can be updated by the governance.

Tick size example:

| significant_amount(AAA) | significant_amount(BBB) | tick_size(AAA/BBB) | tick_size(BBB/AAA) |    
|-------------------------|-------------------------|--------------------|--------------------|
| 10000                   | 10000                   | 0.01               | 0.01               | 
| 1000                    | 10                      | 1                  | 0.0001             | 
| 1000000                 | 1                       | 10000              | 0.00000001         | 

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

#### Matching algorithm

The following algorithm is used to match the orders (for simplicity `max_execution_quantity` and
`opposite_execution_quantity` are assigned directly to the variables used in the algorithm):

```
var (order_to_close, order_to_reduce)
if taker_order(base_denom/quote_denom) == buy
  if taker_order(base_denom/quote_denom).price >= maker_order(base_denom/quote_denom|sell).price 
    maker_order = maker_order(base_denom/quote_denom|sell)
  inv_maker_price = 1/maker_order(quote_denom/base_denom|buy).price  
  if taker_order(base_denom/quote_denom).price >= inv_maker_price
    // comapre self and opposite order books orders
    if maker_order.price > inv_maker_price
      maker_order = maker_order(quote_denom/base_denom|buy)
  if maker_order is maker_order(base_denom/quote_denom|sell)
    if taker_order.remaining_quantity >= maker_order.remaining_quantity
      order_to_close = maker_order
      order_to_reduce = taker_order
      
      execution_price = maker_order.price
      n = floor(order_to_close.remaining_quantity / execution_price.denominator)
      order_to_reduce_filled_quantity = n * execution_price.denominator
      order_to_reduce_used_balance = n * execution_price.numerator  
      
      order_to_close.account receives:
        order_to_reduce_used_balance|order_to_close.quote_denom
        order_to_close.remaining_balance - order_to_reduce_filled_quantity|order_to_close.base_denom
      order_to_reduce.account receives:    
        order_to_reduce_filled_quantity|order_to_reduce.base_denom
             
    else (taker_order.remaining_quantity < maker_order.remaining_quantity)
      order_to_close = taker_order
      order_to_reduce = maker_order
      
      execution_price = maker_order.price
      n = floor(order_to_close.remaining_quantity / execution_price.denominator)
      order_to_reduce_filled_quantity = n * execution_price.denominator
      order_to_reduce_used_balance = order_to_reduce_filled_quantity
      order_to_close_used_balance = n * execution_price.numerator
      
      order_to_close.account receives:
        order_to_reduce_used_balance|order_to_close.base_denom
        order_to_close.remaining_balance - order_to_close_used_balance|order_to_close.quote_denom
      order_to_reduce.account receives:    
        order_to_close_used_balance|order_to_reduce.quote_denom
     
  else (maker_order(quote_denom/base_denom|buy))
    if taker_order.remaining_quantity * 1/maker_order.price >= maker_order.remaining_quantity
      order_to_close = maker_order
      order_to_reduce = taker_order
      
      execution_price = maker_order.price
      n = floor(order_to_close.remaining_quantity / execution_price.denominator)
      order_to_reduce_filled_quantity = n * execution_price.numerator
      order_to_reduce_used_balance = n * execution_price.denominator  
      
      order_to_close.account receives:
        order_to_reduce_used_balance|order_to_close.base_denom
        order_to_close.remaining_balance - order_to_reduce_filled_quantity|order_to_close.quote_denom
      order_to_reduce.account receives:    
        order_to_reduce_filled_quantity|order_to_reduce.base_denom
     
    else (taker_order.remaining_quantity * 1/maker_order.price < maker_order.remaining_quantity)
      order_to_close = taker_order
      order_to_reduce = maker_order

      execution_price = 1/maker_order.price
      n = floor(order_to_close.remaining_quantity / execution_price.denominator)
      order_to_reduce_filled_quantity = n * execution_price.numerator
      order_to_reduce_used_balance = n * execution_price.denominator  
      
      order_to_close.account receives:
        order_to_reduce_used_balance|order_to_close.base_denom
        order_to_close.remaining_balance - order_to_reduce_filled_quantity|order_to_close.quote_denom
      order_to_reduce.account receives:    
        order_to_reduce_filled_quantity|order_to_reduce.base_denom

if taker_order(base_denom/quote_denom) == sell
  if taker_order(base_denom/quote_denom).price < maker_order(base_denom/quote_denom|buy).price 
    maker_order = maker_order(base_denom/quote_denom|buy)
  inv_maker_price = 1/maker_order(quote_denom/base_denom|sell).price  
  if taker_order(base_denom/quote_denom).price < inv_maker_price
    // comapre self and opposite order books orders
    if maker_order.price < inv_maker_price
      maker_order = maker_order(quote_denom/base_denom|sell)
  if maker_order is maker_order(base_denom/quote_denom|buy)
    if taker_order.remaining_quantity >= maker_order.remaining_quantity
      order_to_close = maker_order
      order_to_reduce = taker_order
      
      execution_price = maker_order.price
      n = floor(order_to_close.remaining_quantity / execution_price.denominator)
      order_to_reduce_filled_quantity = n * execution_price.denominator
      order_to_reduce_used_balance = order_to_reduce_filled_quantity
      order_to_close_used_balance = n * execution_price.numerator
      
      order_to_close.account receives:
        order_to_reduce_used_balance|order_to_close.quote_denom
        order_to_close.remaining_balance - order_to_close_used_balance|order_to_close.base_denom
      order_to_reduce.account receives:    
        order_to_close_used_balance|order_to_reduce.quote_denom
             
    else (taker_order.remaining_quantity < maker_order.remaining_quantity)
      order_to_close = taker_order
      order_to_reduce = maker_order

      execution_price = maker_order.price
      n = floor(order_to_close.remaining_quantity / execution_price.denominator)
      order_to_reduce_filled_quantity = n * execution_price.denominator 
      order_to_reduce_used_balance =  n * execution_price.numerator
      
      order_to_close.account receives:
        order_to_reduce_used_balance|order_to_close.quote_denom
        order_to_close.remaining_balance - order_to_reduce_filled_quantity|order_to_close.base_denom
      order_to_reduce.account receives:    
        order_to_reduce_filled_quantity|order_to_reduce.base_denom
           
  else (maker_order(quote_denom/base_denom|sell))     
    if taker_order.remaining_quantity * 1/maker_order.price >= maker_order.remaining_quantity
        order_to_close = maker_order
        order_to_reduce = taker_order
        
        execution_price = maker_order.price
        n = floor(order_to_close.remaining_quantity / execution_price.denominator)
        order_to_reduce_filled_quantity = n * execution_price.denominator
        order_to_reduce_used_balance = order_to_reduce_filled_quantity
        order_to_close_used_balance = n * execution_price.numerator
        
        order_to_close.account receives:
          order_to_reduce_used_balance|order_to_close.quote_denom
          order_to_close.remaining_balance - order_to_close_used_balance|order_to_close.base_denom
        order_to_reduce.account receives:    
          order_to_close_used_balance|order_to_reduce.quote_denom
       
    else (taker_order.remaining_quantity * 1/maker_order.price < maker_order.remaining_quantity)
      order_to_close = taker_order
      order_to_reduce = maker_order

      execution_price = 1/maker_order.price
      n = floor(order_to_close.remaining_quantity / execution_price.denominator)
      order_to_reduce_filled_quantity = n * execution_price.numerator
      order_to_reduce_used_balance = order_to_reduce_filled_quantity
      order_to_close_used_balance = n * execution_price.denominator
      
      order_to_close.account receives:
        order_to_reduce_used_balance|order_to_close.quote_denom
        order_to_close.remaining_balance - order_to_close_used_balance|order_to_close.base_denom
      order_to_reduce.account receives:    
        order_to_close_used_balance|order_to_reduce.quote_denom
        
order_to_reduce.remaining_quantity -= order_to_reduce_filled_quantity 
order_to_reduce.remaining_balance -= order_to_reduce_used_balance
          
close order_to_close
update order_to_reduce (or close if filled)
```

#### Matching example

Let's say we have two tokens AAA and BBB.

```
significant_amount(AAA) = 100
significant_amount(BBB) = 10

tick_size(AAA/BBB) = 0.01 * 10 / 100 = 0.001
tick_size(BBB/AAA) = 0.01 * 100 / 10 = 0.1
```

If `base_denom` is `AAA` and `quote_denom` is `BBB` the `price` must be multiple of `tick_size(AAA/BBB)` and vice
versa.

##### Round 1 (close AAA-BBB sell maker with AAA-BBB buy taker)

Let's say we have 2 orders:

| order_id | account  | base_denom | quote_denom | side | remaining_quantity | remaining_balance | price |
|----------|----------|------------|-------------|------|--------------------|-------------------|-------|
| order1   | account1 | AAA        | BBB         | sell | 50000000 AAA       | 50000000 AAA      | 0.371 |
| order2   | account2 | AAA        | BBB         | buy  | 60000000 AAA       | 22320000 BBB      | 0.372 |

Matching

```
taker_order(base_denom/quote_denom) == buy
  0.372 >= 0.371
  maker_order is maker_order(base_denom/quote_denom|sell)
    60000000 >= 50000000
      order_to_close = order1
      order_to_reduce = order2
     
      execution_price = 0.371 = 371/1000
      n = floor(50000000 / 1000) = 50000
      order_to_reduce_filled_quantity = 50000 * 1000 = 50000000
      order_to_reduce_used_balance = 50000 * 371 = 18550000
      
      account1 receives:
        18550000|BBB
        50000000 - 50000000 = 0|AAA
      account2 receives:    
        50000000|AAA
    
order2.remaining_quantity = 60000000 - 50000000 = 10000000
order2.remaining_balance = 22320000 - 18550000 = 3770000 
```

The result order book and account balances look like:

| order_id | account  | base_denom | quote_denom | side | remaining_quantity | remaining_balance | price |
|----------|----------|------------|-------------|------|--------------------|-------------------|-------|
| order2   | account2 | AAA        | BBB         | buy  | 10000000 AAA       | 3770000 BBB       | 0.372 |

| account  | balance      |
|----------|--------------|
| account1 | 18550000 BBB | 
| account2 | 50000000 AAA |

* order1 is executed fully, with the exact price.
* order2 is executed partially with better price.

##### Round 2 (close AAA-BBB buy maker with BBB-AAA buy taker)

Let's say we have the previous partially filled order and new order:

| order_id | account  | base_denom | quote_denom | side | remaining_quantity | remaining_balance | price |
|----------|----------|------------|-------------|------|--------------------|-------------------|-------|
| order2   | account2 | AAA        | BBB         | buy  | 10000000 AAA       | 3770000 BBB       | 0.372 |
| order3   | account3 | BBB        | AAA         | buy  | 33300000 BBB       | 89910000 AAA      | 2.7   |

Matching

```
taker_order(base_denom/quote_denom) == buy
  inv_maker_price = 1/0.372 ~= 2.6881
  2.7 >= 2.6881
  maker_order(quote_denom/base_denom|buy)
    33300000 * 1/0.372 ~= 89516129.0322 >= 10000000
      order_to_close = order2
      order_to_reduce = order3
      
      execution_price = 0.372 = 372 / 1000
      n = floor(10000000 / 1000) = 10000
      order_to_reduce_filled_quantity = 10000 * 372 = 3720000
      order_to_reduce_used_balance = 10000 * 1000 = 10000000  
      
      account2 receives:
        10000000|AAA
        3770000 - 3720000 = 50000|BBB
      account3 receives:    
        3720000|BBB

order_to_reduce.remaining_quantity = 33300000 - 3720000 = 29580000
order_to_reduce.remaining_balance = 89910000 - 10000000 = 79910000   
```

The result order book and account balances look like:

| order_id | account  | base_denom | quote_denom | side | remaining_quantity | remaining_balance | price |
|----------|----------|------------|-------------|------|--------------------|-------------------|-------|
| order3   | account3 | BBB        | AAA         | buy  | 29580000 BBB       | 79910000 AAA      | 2.7   |

| account  | balance                                       |
|----------|-----------------------------------------------|
| account1 | 18550000 BBB                                  | 
| account2 | 50000000 + 10000000 = 60000000 AAA, 50000 BBB |
| account3 | 3720000 BBB                                   |

* order2 received full expected 60000000 AAA amount and additionaly the remainder of 50000 BBB.
* order3 is executed partially with better price.

##### Round 3 (close AAA-BBB buy taker with BBB-AAA buy maker)

Let's say we have the previous partially filled order and new order:

| order_id | account  | base_denom | quote_denom | side | remaining_quantity | remaining_balance | price |
|----------|----------|------------|-------------|------|--------------------|-------------------|-------|
| order3   | account3 | BBB        | AAA         | buy  | 29580000 BBB       | 79910000 AAA      | 2.7   |
| order4   | account4 | AAA        | BBB         | buy  | 10000 AAA          | 3800 BBB          | 0.38  |

Matching

```
taker_order(base_denom/quote_denom) == buy
  inv_maker_price = 1/2.7 (~=0.3703)
  0.38 >= 0.3703
  maker_order(quote_denom/base_denom|buy)
    10000 * 1/2.7 ~= 3703.70 < 29580000
      order_to_close = order4
      order_to_reduce = order3
    
      execution_price = 1/2.7 = 10/27
      n = floor(10000 / 27) = 370
      order_to_reduce_filled_quantity = 370 * 10 = 3700
      order_to_reduce_used_balance = 370 * 27 = 9990
      
      account4 receives:
          9990|AAA
          3800 - 3700 = 100|BBB
      account3 receives:    
          3700|BBB
        
order_to_reduce.remaining_quantity = 29580000 - 3700 = 29576300
order_to_reduce.remaining_balance = 79910000 - 9990 = 79900010   
```

The result order book and account balances look like:

| order_id | account  | base_denom | quote_denom | side | remaining_quantity | remaining_balance | price |
|----------|----------|------------|-------------|------|--------------------|-------------------|-------|
| order3   | account3 | BBB        | AAA         | buy  | 29576300 BBB       | 79900010 AAA      | 2.7   |

| account  | balance                      |
|----------|------------------------------|
| account1 | 18550000 BBB                 | 
| account2 | 60000000 AAA, 50000 BBB      |
| account3 | 3720000 + 3700 = 3723700 BBB |
| account4 | 9990 AAA, 100 BBB            |

* order3 is executed partially with exact price.
* order4 is executed partially but with better price, and received 100 BBB.

##### Round 4 (close BBB-AAA buy maker with BBB-AAA sell taker)

Let's say we have the previous partially filled order and new order:

| order_id | account  | base_denom | quote_denom | side | remaining_quantity | remaining_balance | price |
|----------|----------|------------|-------------|------|--------------------|-------------------|-------|
| order3   | account3 | BBB        | AAA         | buy  | 29576300 BBB       | 79900010 AAA      | 2.7   |
| order5   | account5 | BBB        | AAA         | sell | 100000000 BBB      | 100000000 BBB     | 2.6   |

Matching

```
taker_order(base_denom/quote_denom) == sell
  2.6 < 2.7
  maker_order(base_denom/quote_denom|buy)  
    100000000 > 29576300
      order_to_close = order3
      order_to_reduce = order5
      
      execution_price = 2.7 = 27/10
      n = floor(29576300 / 10) = 2957630
      order_to_reduce_filled_quantity = 2957630 * 10 = 29576300
      order_to_reduce_used_balance = 29576300
      order_to_close_used_balance = 2957630 * 27 = 79856010
    
      account3 receives:
        29576300|BBB
        79900010 - 79856010 = 44000|AAA
      account5 receives:    
        79856010|AAA
    
order5.remaining_quantity = 100000000 - 29576300 = 70423700
order5.remaining_balance = 100000000 - 29576300 =70423700

```

The result order book and account balances look like:

| order_id | account  | base_denom | quote_denom | side | remaining_quantity | remaining_balance | price |
|----------|----------|------------|-------------|------|--------------------|-------------------|-------|
| order5   | account5 | BBB        | AAA         | sell | 70423700 BBB       | 70423700 BBB      | 2.6   |

| account  | balance                                      |
|----------|----------------------------------------------|
| account1 | 18550000 BBB                                 | 
| account2 | 60000000 AAA, 50000 BBB                      |
| account3 | 44000 AAA, 3723700 + 29576300 = 33300000 BBB |
| account4 | 9990 AAA, 100 BBB                            |
| account5 | 79856010 AAA                                 |

* order3 is executed fully, and bought 33300000 BBB, as expected with the better price, since 44000 AAA returned.
* order5 is executed partially but with better price.

##### Round 5 (close BBB-AAA sell maker with AAA-BBB sell taker)

Let's say we have the previous partially filled order and new order:

| order_id | account  | base_denom | quote_denom | side | remaining_quantity | remaining_balance | price |
|----------|----------|------------|-------------|------|--------------------|-------------------|-------|
| order5   | account5 | BBB        | AAA         | sell | 70423700 BBB       | 70423700 BBB      | 2.6   |
| order6   | account6 | AAA        | BBB         | sell | 1000000000 AAA     | 1000000000 AAA    | 0.383 |

Matching

```
taker_order(base_denom/quote_denom) == sell
  inv_maker_price = 1/2.6 ~= 0.3846
  0.383 < 0.3846
  maker_order(quote_denom/base_denom|sell)
    1000000000 * 1/2.6 = 384615384 > 70423700
      order_to_close = order5
      order_to_reduce = order6
      
      execution_price = 2.6 = 26 / 10
      n = floor(70423700 / 10)
      order_to_reduce_filled_quantity = 7042370 * 26 = 183101620
      order_to_reduce_used_balance = 183101620
      order_to_close_used_balance = 7042370 * 10 = 70423700
      
      account5 receives:
        183101620|AAA
        70423700 - 70423700 = 0|BBB
      account6 receives:    
        70423700|BBB
    
order_to_reduce.remaining_quantity -= 1000000000 - 183101620 = 816898380
order_to_reduce.remaining_balance -= 1000000000 - 183101620 = 816898380

```

The result order book and account balances look like:

| order_id | account  | base_denom | quote_denom | side | remaining_quantity | remaining_balance | price |
|----------|----------|------------|-------------|------|--------------------|-------------------|-------|
| order6   | account6 | AAA        | BBB         | sell | 816898380 AAA      | 816898380 AAA     | 0.383 |

| account  | balance                              |
|----------|--------------------------------------|
| account1 | 18550000 BBB                         | 
| account2 | 60000000 AAA, 50000 BBB              |
| account3 | 44000 AAA, 33300000 BBB              |
| account4 | 9990 AAA, 100 BBB                    |
| account5 | 79856010 + 183101620 = 262957630 AAA |
| account6 | 70423700 BBB                         |

* order5 is executed fully, and sold all 100000000 BBB, as received more than expected 260000000 AAA, the 262957630 AAA.
* order6 is executed partially but with better price.

##### Round 6 (close BBB-AAA sell taker with AAA-BBB sell maker)

Let's say we have the previous partially filled order and new order:

| order_id | account  | base_denom | quote_denom | side | remaining_quantity | remaining_balance | price |
|----------|----------|------------|-------------|------|--------------------|-------------------|-------|
| order6   | account6 | AAA        | BBB         | sell | 816898380 AAA      | 816898380 AAA     | 0.383 |
| order7   | account7 | BBB        | AAA         | sell | 100000 BBB         | 100000 BBB        | 2.6   |

Matching:

```
taker_order(base_denom/quote_denom) == sell
  inv_maker_price = 1/2.6 ~= 0.3846
  0.383 < 0.3846
  maker_order(quote_denom/base_denom|sell)
    100000 < 1/2.6 * 816898380 ~= 314191684
      order_to_close = order7
      order_to_reduce = order6
      
      execution_price = 1 / 0.383 = 1000 / 383  
      n = floor(100000 / 383) = 261
      order_to_reduce_filled_quantity = 261 * 1000 = 261000
      order_to_reduce_used_balance = 261000
      order_to_close_used_balance = 261 * 383 = 99963
    
    account7 receives:
      261000|AAA
      100000 - 99963 = 37|BBB
    account6 receives:    
      99963|BBB
 
order_to_reduce.remaining_quantity -= 816898380 - 261000 = 816637380
order_to_reduce.remaining_balance -= 816898380 - 261000 = 816637380

```

The result order book and account balances look like:

| order_id | account  | base_denom | quote_denom | side | remaining_quantity | remaining_balance | price |
|----------|----------|------------|-------------|------|--------------------|-------------------|-------|
| order6   | account6 | AAA        | BBB         | sell | 816637380 AAA      | 816637380 AAA     | 0.383 |

| account  | balance                              |
|----------|--------------------------------------|
| account1 | 18550000 BBB                         | 
| account2 | 60000000 AAA, 50000 BBB              |
| account3 | 44000 AAA, 33300000 BBB              |
| account4 | 9990 AAA, 100 BBB                    |
| account5 | 79856010 + 183101620 = 262957630 AAA |
| account6 | 70423700 + 99963 = 70523663 BBB      |
| account7 | 261000 AAA,  37 BBB                  |

* order6 is executed partially with exact price.
* order7 is executed fully, and sold 100000 - 37 BBB, 261000 AAA, with better price.

##### Round 7 (close AAA-BBB buy taker with AAA-BBB sell maker)

Let's say we have the previous partially filled order and new order:

| order_id | account  | base_denom | quote_denom | side | remaining_quantity | remaining_balance | price |
|----------|----------|------------|-------------|------|--------------------|-------------------|-------|
| order6   | account6 | AAA        | BBB         | sell | 816637380 AAA      | 816637380 AAA     | 0.383 |
| order8   | account8 | AAA        | BBB         | buy  | 100000 AAA         | 38500 BBB         | 0.385 |

Matching:

```
taker_order(base_denom/quote_denom) == buy
  0.385 > 0.383
  maker_order(base_denom/quote_denom|sell)
    100000 < 816637380
      order_to_close = order8
      order_to_reduce = order6
    
      execution_price = 383 / 1000
      n = floor(100000 / 1000) = 100
      order_to_reduce_filled_quantity = 100 * 1000 = 100000
      order_to_reduce_used_balance = 100000
      order_to_close_used_balance = 100 * 383 = 38300
      
      account8 receives:
        100000|AAA
        38500 - 38300 = 200|BBB
      order6 receives:    
        38300|BBB

order_to_reduce.remaining_quantity = 816637380 - 100000 = 816537380
order_to_reduce.remaining_balance = 816637380 -  100000 = 816537380
```

The result order book and account balances look like:

| order_id | account  | base_denom | quote_denom | side | remaining_quantity | remaining_balance | price |
|----------|----------|------------|-------------|------|--------------------|-------------------|-------|
| order6   | account6 | AAA        | BBB         | sell | 816537380 AAA      | 816537380 AAA     | 0.383 |

| account  | balance                         |
|----------|---------------------------------|
| account1 | 18550000 BBB                    | 
| account2 | 60000000 AAA, 50000 BBB         |
| account3 | 44000 AAA, 33300000 BBB         |
| account4 | 9990 AAA, 100 BBB               |
| account5 | 262957630 AAA                   |
| account6 | 70523663 + 38300 = 70561963 BBB |
| account7 | 261000 AAA,  37 BBB             |
| account8 | 100000 AAA,  200 BBB            |

* order6 is executed partially with exact price.
* order8 is executed fully, bought exact 100000 AAA, and remainder 200 BBB, with better price.

##### Round 8 (close AAA-BBB sell maker with AAA-BBB buy taker, not unique, but required for the next round)

Let's say we have the previous partially filled order and new order:

| order_id | account  | base_denom | quote_denom | side | remaining_quantity | remaining_balance | price |
|----------|----------|------------|-------------|------|--------------------|-------------------|-------|
| order6   | account6 | AAA        | BBB         | sell | 816537380 AAA      | 816537380 AAA     | 0.383 |
| order9   | account9 | AAA        | BBB         | buy  | 7000000000 AAA     | 2695000000 BBB    | 0.385 |

Matching

```
taker_order(base_denom/quote_denom) == buy
  0.385 >= 0.383
  maker_order is maker_order(base_denom/quote_denom|sell)
    7000000000 >= 816537380
      order_to_close = order6
      order_to_reduce = order9
     
      execution_price = 0.383 = 383 / 1000
      n = floor(816537380 / 1000) = 816537
      order_to_reduce_filled_quantity = 816537 * 1000 = 816537000
      order_to_reduce_used_balance = 816537 * 383 = 816537 * 383 = 312733671
      
      account6 receives:
        312733671|BBB
        816537380 - 816537000 = 380|AAA
      account9 receives:    
        816537000|AAA

order9.remaining_quantity = 7000000000 - 816537000 = 6183463000
order9.remaining_balance = 2695000000 - 312733671 = 2382266329 
```

The result order book and account balances look like:

| order_id | account  | base_denom | quote_denom | side | remaining_quantity | remaining_balance | price |
|----------|----------|------------|-------------|------|--------------------|-------------------|-------|
| order9   | account9 | AAA        | BBB         | buy  | 6183463000 AAA     | 2382266329 BBB    | 0.385 |

| account  | balance                                       |
|----------|-----------------------------------------------|
| account1 | 18550000 BBB                                  | 
| account2 | 60000000 AAA, 50000 BBB                       |
| account3 | 44000 AAA, 33300000 BBB                       |
| account4 | 9990 AAA, 100 BBB                             |
| account5 | 262957630 AAA                                 |
| account6 | 380 AAA, 70561963 + 312733671 = 383295634 BBB |
| account7 | 261000 AAA,  37 BBB                           |
| account8 | 100000 AAA,  200 BBB                          |
| account9 | 816537000 AAA                                 |

* order6 is executed fully, expected to receive 383000000 BBB, but received 383295634 BBB, and remainder 200 BBB.
* order9 is executed partially with exact price.

##### Round 9 (close AAA-BBB sell taker with AAA-BBB buy maker)

Let's say we have the previous partially filled order and new order:

| order_id | account   | base_denom | quote_denom | side | remaining_quantity | remaining_balance | price |
|----------|-----------|------------|-------------|------|--------------------|-------------------|-------|
| order9   | account9  | AAA        | BBB         | buy  | 6183463000 AAA     | 2382266329 BBB    | 0.385 |
| order10  | account10 | AAA        | BBB         | sell | 550000000 AAA      | 550000000 AAA     | 0.382 |

Matching

```
taker_order(base_denom/quote_denom) == sell
  0.382 < 0.385
  maker_order is maker_order(base_denom/quote_denom|buy)
    550000000 < 6183463000
      order_to_close = taker_order
      order_to_reduce = maker_order      
      
      execution_price = 0.385 = 385 / 1000
      n = floor(550000000 / 1000) = 550000
      order_to_reduce_filled_quantity = 550000 * 1000 = 550000000
      order_to_reduce_used_balance = 550000 * 385 = 211750000

      account10 receives:
        211750000|BBB
        550000000 - 550000000|AAA
      account9 receives:    
        550000000|AAA

order9.remaining_quantity = 6183463000 - 550000000 = 5633463000
order9.remaining_balance = 2382266329 - 211750000 = 2170516329 
```

The result order book and account balances look like:

| order_id | account  | base_denom | quote_denom | side | remaining_quantity | remaining_balance | price |
|----------|----------|------------|-------------|------|--------------------|-------------------|-------|
| order9   | account9 | AAA        | BBB         | buy  | 5633463000 AAA     | 2170516329 BBB    | 0.385 |

| account   | balance                                |
|-----------|----------------------------------------|
| account1  | 18550000 BBB                           | 
| account2  | 60000000 AAA, 50000 BBB                |
| account3  | 44000 AAA, 33300000 BBB                |
| account4  | 9990 AAA, 100 BBB                      |
| account5  | 262957630 AAA                          |
| account6  | 380 AAA, 383295634 BBB                 |
| account7  | 261000 AAA,  37 BBB                    |
| account8  | 100000 AAA,  200 BBB                   |
| account9  | 816537000 + 550000000 = 1366537000 AAA |
| account10 | 211750000 BBB                          |

* order9 is executed partially with exact price.
* order10 is executed partially with better price.