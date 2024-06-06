# DEX

The spec describes the coreum DEX specification.

## Order book

The Coreum DEX orders can be created for any-to-any token by any Coreum user. Which meas that the order book
is bidirectional and permissionless. The user place an order and provides the order attributes:

* `order_id` - unique order identifier per account and order book
* `price` - amount the user wants to get for each token the user sells
* `quantity` - amount the user want to sell
* `fill_side`
    * `sell` - means that the `remaining_quantity` you expect to fill is equal to `quantity` of the `sell_denom`.
    * `buy`  - means that the `remaining_quantity` you expect to fill is equal to `quantity` * `price` of
      the `buy_denom`.

Once an order is placed the DEX will try to match the order with the orders from the opposite order book and depending
on order type and settings execute it.

## Matching

### Rounding issue

Let's say we have 2 orders:

| order_id | sell_denom | buy_denom | remaining_quantity | price | fill_side | remaining_quantity |
|----------|------------|-----------|--------------------|-------|-----------|--------------------|
| order1   | AAA        | BBB       | 50_000_000         | 0.375 | sell      | 50_000_000 AAA     |
| order2   | BBB        | AAA       | 10_000_000         | 2.6   | sell      | 10_000_000 BBB     |

The inverse taker price (1/price) of order2 is greater than the price of order1 (~0.3846 > 0.375)
hence orders match. The order2 should be executed with the price of the order1 (taker gets the better price).
The exact amount of AAA order2 receives is 10_000_000 * 1/0.375 = 26_666_666.(6) AAA. The amount we can use
must be an integer, so we can't send full amount without the price violation. The following solution fixes that issue.

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
| 1_000                   | 10                      | 0.0001             | 1                  | 
| 1_000_000               | 1                       | 0.00000001         | 10_000             | 

### Matching with max execution quantity

With this strategy, when orders match, we target to fill order `remaining_quantity` fully. In case we can't do it fully
due to the rounding, we find `max_execution_quantity` that we can use to prevent the price violation, and return
the remainder.

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

The following algorithm is used to match the orders:

```
if 1/taker_order.price > maker_order.price:
    var maker_expected_not_filled_quantity
    if maker_order.fill_side != taker_order.fill_side // orders not_filled_quantit use same denom already
        maker_expected_not_filled_quantity = maker_order.not_filled_quantity
    else
        if maker_order.fill_side == sell
            maker_expected_not_filled_quantity = maker_order.not_filled_quantity * maker_order.price
        if maker_order.fill_side == buy
            maker_expected_not_filled_quantity = maker_order.not_filled_quantity / maker_order.price
    var (
        order_to_close
        order_to_reduce
        
        order_to_close_used_quantity
        order_to_reduce_used_quantity
    ) 
    if maker_expected_not_filled_quantity > taker_order.not_filled_quantity
        order_to_close = taker_order
        order_to_reduce = maker_order
        
        if order_to_close.fill_side == sell
            execution_price = 1/maker_order.price
            n = floor(order_to_close.not_filled_quantity / execution_price.denominator)
            order_to_close_used_quantity = n * execution_price.denominator      
            order_to_reduce_used_quantity = n * execution_price.nominator            
        
        else (order_to_close.fill_side == buy)
           execution_price = maker_order.price
           // here the order_to_reduce_used_quantity is based on the order_to_close.not_filled_quantity
           n = floor(order_to_close.not_filled_quantity / execution_price.denominator)
           order_to_reduce_used_quantity = n * execution_price.denominator      
           order_to_close_used_quantity = n * execution_price.nominator
                     
    else (maker_expected_not_filled_quantity <= taker_order.not_filled_quantity)  
        order_to_close = maker_order
        order_to_reduce = taker_order
        
        if order_to_close.fill_side == sell
            execution_price = maker_order.price
            n = floor(order_to_close.not_filled_quantity / execution_price.denominator)
            order_to_close_used_quantity = n * execution_price.denominator      
            order_to_reduce_used_quantity = n * execution_price.nominator
                                           
        else (order_to_close.fill_side == buy)
           execution_price = 1/maker_order.price
           // here the order_to_reduce_used_quantity is based on the order_to_close.not_filled_quantity
           n = floor(order_to_close.not_filled_quantity / execution_price.denominator)
           order_to_reduce_used_quantity = n * execution_price.denominator      
           order_to_close_used_quantity = n * execution_price.nominator
   
    var order_to_reduce_filled_quantity // quantity to sub from the not_filled_quantity  
    if order_to_reduce.fill_side == sell
        order_to_reduce_filled_quantity = order_to_reduce_used_quantity    
    else (order_to_reduce.fill_side == buy)
        order_to_reduce_filled_quantity = order_to_close_used_quantity  
   
    order_to_close.account receives:
        order_to_reduce_used_quantity|order_to_close.buy_denom
        order_to_close.remaining_quantity - order_to_close_used_quantity|order_to_close.sell_denom
    
    order_to_reduce.account receives:    
        order_to_close_used_quantity|order_to_reduce.buy_denom

    order_to_reduce.remaining_quantity -= order_to_reduce_used_quantity 
    order_to_reduce.not_filled_quantity -= order_to_reduce_filled_quantity
      
    close order_to_close
    update order_to_reduce
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

| order_id | account  | sell_denom | buy_denom | remaining_quantity | price | fill_side | not_filled_quantity | 
|----------|----------|------------|-----------|--------------------|-------|-----------|---------------------|
| order1   | account1 | AAA        | BBB       | 50_000_000 AAA     | 0.371 | sell      | 50_000_000 AAA      |  
| order2   | account2 | BBB        | AAA       | 10_000_000 BBB     | 2.6   | sell      | 10_000_000 BBB      | 

Matching

```
1/2.6 > 0.371 (~0.3846 > 0.371)

sell == sell
    maker_order.fill_side == sell  
        maker_expected_not_filled_quantity = 50_000_000 * 0.371 = 18_550_000
 
18_550_000 > 10_000_000
    order_to_close = order2
    order_to_reduce = order1
    
    order_to_close.fill_side == sell
        execution_price = 1 / 0.371 = 1_000 / 371
        n = floor(10_000_000 / 371) = 26_954
        order_to_close_used_quantity = 26_954 * 371 = 9_999_934     
        order_to_reduce_used_quantity = 26_954 * 1_000 = 26_954_000
        
    order_to_reduce.fill_side == sell
        order_to_reduce_filled_quantity = 26_954_000

    account2 receives:
        26_954_000|AAA
        10_000_000 - 9_999_934 = 66|BBB
    
    account1 receives:    
        9_999_934|BBB

    order1.remaining_quantity = 50_000_000 - 26_954_000 = 23_046_000
    order1.not_filled_quantity = 50_000_000 - 26_954_000 = 23_046_000
```

The result order book and account balances look like:

| order_id | account  | sell_denom | buy_denom | remaining_quantity | price | fill_side | not_filled_quantity |
|----------|----------|------------|-----------|--------------------|-------|-----------|---------------------|
| order1   | account1 | AAA        | BBB       | 23_046_000  AAA    | 0.371 | sell      | 23_046_000  AAA     |

| account  | balance                |
|----------|------------------------|
| account1 | 9_999_934 BBB          | 
| account2 | 26_954_000 AAA, 66 BBB |

* order1 was filled partially with the exact price 9_999_934 BBB / 26_954_000 AAA = 0.371
* order2 expected to sell 10_000_000 BBB, receive 26_000_000 AAA, but received more, 26_954_000 AAA and
  additionally not filled 66 BBB.

* Round 2

Let's say we have the previous partially filled order and new order:

| order_id | account  | sell_denom | buy_denom | remaining_quantity | price | fill_side | not_filled_quantity |
|----------|----------|------------|-----------|--------------------|-------|-----------|---------------------|
| order1   | account1 | AAA        | BBB       | 23_046_000  AAA    | 0.371 | sell      | 23_046_000  AAA     |
| order3   | account3 | BBB        | AAA       | 70_000_000 BBB     | 2.3   | sell      | 70_000_000 BBB      |

Matching

```
1/2.3 > 0.371 (~0.4347 > 0.371)

sell == sell
  maker_order.fill_side == sell    
    maker_expected_not_filled_quantity = 23_046_000 * 0.371 = 8_550_066

8_550_066 < 70_000_000
    order_to_close = order1
    order_to_reduce = order3
    order_to_close.fill_side == sell
        execution_price = 0.371 = 371 / 1_000
        n = floor(23_046_000 / 1_000) = 23046
        order_to_close_used_quantity = 23046 * 1_000 = 23_046_000     
        order_to_reduce_used_quantity = 23046 * 371 = 8_550_066
        
    order_to_reduce.fill_side == sell
        order_to_reduce_filled_quantity = 8_550_066
                
    account1 receives:
        8_550_066|BBB
        23_046_000 - 23_046_000 = 0|AAA
    
    account3 receives:    
        23_046_000|AAA

    order3.remaining_quantity = 70_000_000 - 8_550_066 
    order3.not_filled_quantity = 70_000_000 - 8_550_066
```

The result order book and account balances look like:

| order_id | account  | sell_denom | buy_denom | remaining_quantity | price | fill_side | not_filled_quantity |
|----------|----------|------------|-----------|--------------------|-------|-----------|---------------------|
| order3   | account3 | BBB        | AAA       | 61_449_934 BBB     | 2.3   | sell      | 61_449_934 BBB      |

| account  | balance                                |
|----------|----------------------------------------|
| account1 | 9_999_934 + 8_550_066 = 18_550_000 BBB | 
| account2 | 26_954_000 AAA, 66 BBB                 |
| account3 | 23_046_000 AAA                         |

* order1 is executed with full match of both quantity and amount
* order3 is executed partially with the better price

* Round 3

Let's say we have the previous partially filled order and new order:

| order_id | account  | sell_denom | buy_denom | remaining_quantity | price | fill_side | not_filled_quantity |
|----------|----------|------------|-----------|--------------------|-------|-----------|---------------------|
| order3   | account3 | BBB        | AAA       | 61_449_934 BBB     | 2.3   | sell      | 61_449_934 BBB      |                     
| order4   | account4 | AAA        | BBB       | 220_000_000 AAA    | 0.36  | buy       | 79_200_000 BBB      | 

Matching

```
1/0.36 > 2.3 (~2.7777 > 2.3)
    
sell != buy
  maker_expected_not_filled_quantity = 61_449_934

61_449_934 < 79_200_000
    order_to_close = order3
    order_to_reduce = order4

    order_to_close.fill_side == sell
        execution_price = 2.3 = 230 / 100
        n = floor(61_449_934 / 100) = 614_499
        order_to_close_used_quantity = 614_499 * 100 = 61_449_900    
        order_to_reduce_used_quantity = 614_499 * 230 = 141_334_770
        
    order_to_reduce.fill_side == buy
        order_to_reduce_filled_quantity = 61_449_900
                      
    account3 receives:
        141_334_770|AAA
        61_449_934 - 61_449_900 = 34|BBB
    
    account4 receives:    
        61_449_900|BBB

    order_to_reduce.remaining_quantity = 220_000_000 - 141_334_770 = 78_665_230
    order_to_reduce.not_filled_quantity = 79_200_000 - 61_449_900 = 17_750_100
```

The result order book and account balances look like:

| order_id | account  | sell_denom | buy_denom | remaining_quantity | price | fill_side | not_filled_quantity |
|----------|----------|------------|-----------|--------------------|-------|-----------|---------------------|
| order4   | account4 | AAA        | BBB       | 78_665_230 AAA     | 0.36  | buy       | 17_750_100 BBB      | 

| account  | balance                                            |
|----------|----------------------------------------------------|
| account1 | 18_550_000 BBB                                     | 
| account2 | 26_954_000 AAA, 66 BBB                             |
| account3 | 23_046_000 + 141_334_770 = 164_380_770 AAA, 34 BBB |
| account4 | 61_449_900 BBB                                     |

* order3 is executed with full match of both quantity and amount, it expected to receive 161_000_000 AAA and
  sell 70_000_000 BBB, but received more 164_380_770 AAA and sold 70_000_000 - 34 = 69_999_966 BBB.
* order4 is executed partially with the better price.

* Round 4

Let's say we have the previous partially filled order and new order:

| order_id | account  | sell_denom | buy_denom | remaining_quantity | price | fill_side | not_filled_quantity |
|----------|----------|------------|-----------|--------------------|-------|-----------|---------------------|
| order4   | account4 | AAA        | BBB       | 78_665_230 AAA     | 0.36  | buy       | 17_750_100 BBB      | 
| order5   | account5 | BBB        | AAA       | 6_000_000 BBB      | 2.7   | buy       | 16_200_000 AAA      | 

Matching

```
1/2.7 > 0.36 (~0.3703 > 0.36)
    
buy == buy
  maker_expected_not_filled_quantity = 17_750_100 / 0.36 ~= 49305833.3333

49305833.3333 > 16_200_000
    order_to_close = order5
    order_to_reduce = order4
    
    order_to_close.fill_side == buy
        execution_price = 0.36 = 36 / 100
        n = floor(16_200_000 / 100) = 162_000
        order_to_reduce_used_quantity = 162_000 * 100 = 16_200_000 
        order_to_close_used_quantity = 162_000 * 36 = 5_832_000

    order_to_reduce.fill_side == buy
        order_to_reduce_filled_quantity = 5_832_000
        
    account5 receives:
        16_200_000|AAA
        6_000_000 - 5_832_000 = 168_000|BBB
    
    account4 receives:    
        5_832_000|BBB

    order_to_reduce.remaining_quantity = 78_665_230 - 16_200_000 = 62_465_230
    order_to_reduce.not_filled_quantity = 17_750_100 - 5_832_000 = 11_918_100         
```

The result order book and account balances look like:

| order_id | account  | sell_denom | buy_denom | remaining_quantity | price | fill_side | not_filled_quantity |
|----------|----------|------------|-----------|--------------------|-------|-----------|---------------------|
| order4   | account4 | AAA        | BBB       | 62_465_230 AAA     | 0.36  | buy       | 11_918_100 BBB      | 

| account  | balance                                 |
|----------|-----------------------------------------|
| account1 | 18_550_000 BBB                          | 
| account2 | 26_954_000 AAA, 66 BBB                  |
| account3 | 164_380_770 AAA, 34 BBB                 |
| account4 | 61_449_900 + 5_832_000 = 67_281_900 BBB |
| account5 | 16_200_000 AAA, 168_000 BBB             |

* order5 is executed with full match of both quantity and amount, expected to receive 6_000_000 * 2.7 = 16_200_000 AAA
  and sell 6_000_000 BBB, but received exact 16_200_000 AAA and sold less 5_832_000 BBB.
* order4 is executed partially with exact price.

* Round 5

Let's say we have the previous partially filled order and new order:

| order_id | account  | sell_denom | buy_denom | remaining_quantity | price | fill_side | not_filled_quantity |
|----------|----------|------------|-----------|--------------------|-------|-----------|---------------------|
| order4   | account4 | AAA        | BBB       | 62_465_230 AAA     | 0.36  | buy       | 11_918_100 BBB      | 
| order6   | account6 | BBB        | AAA       | 39_127_000 BBB     | 2.2   | buy       | 86_079_400 AAA      | 

Matching

```
1/2.2 > 0.36 (~0.4545 > 0.36)

buy == buy
    maker_expected_not_filled_quantity = 11_918_100 / 0.36 ~= 33105833.3333

33_106_138 < 86_079_400
    order_to_close = maker_order
    order_to_reduce = taker_order
    
    order_to_close.fill_side == buy
        execution_price = 1/0.36 = 100 / 36
        n = floor(11_918_100 / 36) = 331_058
        order_to_reduce_used_quantity = 331_058 * 36 = 11_918_088     
        order_to_close_used_quantity = 331_058 * 100 = 33_105_800
    
    order_to_reduce.fill_side == buy
        order_to_reduce_filled_quantity = 33_105_800

    account4 receives:
        11_918_088|BBB
        62_465_230 - 33_105_800 = 29_359_430|AAA
    
    account6 receives:    
        33_105_800|AAA

    order6.remaining_quantity = 39_127_000 - 11_918_088 = 27_208_912
    order6.not_filled_quantity = 86_079_400 - 33_105_800 = 52_973_600
```

The result order book and account balances look like:

| order_id | account  | sell_denom | buy_denom | remaining_quantity | price | fill_side | not_filled_quantity |
|----------|----------|------------|-----------|--------------------|-------|-----------|---------------------|
| order6   | account6 | BBB        | AAA       | 27_208_804 BBB     | 2.2   | buy       | 52_973_300 AAA      | 

| account  | balance                                                  |
|----------|----------------------------------------------------------|
| account1 | 18_550_000 BBB                                           | 
| account2 | 26_954_000 AAA, 66 BBB                                   |
| account3 | 164_380_770 AAA, 34 BBB                                  |
| account4 | 29_359_430 AAA, 67_281_900 + 11_918_088 = 79_199_988 BBB |
| account5 | 16_200_000 AAA, 168_000 BBB                              |
| account6 | 33_105_800 AAA                                           | 

* order5 is executed with close match to the expected buy quantity, expected to receive 79_200_000 BBB, but received
  79_199_988 BBB, also received the remainder of 29_359_430 AAA.
* order6 is executed partially with better price.
