# DEX

This document defines the concepts and mechanics of Coreum on-chain DEX. It is a work in progress.

## Order book

Coreum DEX is based on order books. Each trading pair defines its order book - which is defined by two sequences of orders, one grouping buy orders, the other one grouping sell orders.

Whenever the lowest price offered by any sell order is lower than or equal to the highest price offered by any buy order, the match exists and tokens might be exchanged.

Trading pair (and its order book) is described by two tokens: base token and quote token. The quote token is used to define the price of the base token.

Example: Let's define trading pair:
- base token: `uaaa`
- quote token: `ubbb`

It means that buy order added to the order book defined that way, specifying the price `10` means that someone wants to buy some amount of `uaaa` paying `10ubbb` for each unit of `uaaa`.

Similarly, sell order added to the order book, specifying the same price means that someone wants to sell some amount of `uaaa` charging `10ubbb` for each unit of `uaaa`.

It might be understood intuitively that in those orders `1uaaa` costs `10ubbb`.  

Order books are named by concatenating base (first) and quote (second) token. In this document order books are named using `<base-token>/<quote-token>` schema. For the above example the name is `uaaa/ubbb`

### Order book representation

The commonly used representation of an order book is the stack of two lists sorted by order price, the upper one containing sell orders and the lower one containing buy orders.

Example: This is the hypothetical `uaaa/ubbb` order book:

| Amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy price `[ubbb/uaaa]` | Explanation |                                                                                
|-----------------|--------------------------|-------------------------|-------------------------------------------------------------------------------------------------------------|
| 100             | 20                       |                         | This means that someone wants to **sell** `100uaaa` charging `20ubbb` for each `1uaaa`, `2000ubbb` in total |
| 50              | 20                       |                         |
| 300             | 15                       |                         |
| 50              |                          | 10                      | This means that someone wants to **buy** `50uaaa` paying `10ubbb` for each `1uaaa`, `500ubbb` in total |
| 150             |                          | 10                      |
| 10              |                          | 5                       |

Some records in the table have the same price, in the presentation layer (web, mobile apps, GUIs) they are compressed into one:

| Amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy price `[ubbb/uaaa]` |                                                                                
|-----------------|--------------------------|-------------------------|
| 150             | 20                       |                         |
| 300             | 15                       |                         |
| 200             |                          | 10                      |
| 10              |                          | 5                       |

Please notice, how the values inside the `Amount` column are added.

Keep in mind that internally, inside the database, each order is still kept separately for the purpose of order matching.

### Order sequencing

Order book for each trading pair contains two tables:
- containing sell orders ("sell side") - this table is sorted by price in ascending order
- containing buy orders ("buy side") - this table is sorted by price in descending order

This is how the two sorted sides of the order book, presented in the example above, look like:

Sell side:

| Amount `[uaaa]` | Price `[ubbb/uaaa]` |                                                                                
|-----------------|---------------------|
| 300             | 15                  |
| 50              | 20                  |
| 100             | 20                  |

Buy side:

| Amount `[uaaa]` | Price `[ubbb/uaaa]` |                                                                                
|-----------------|---------------------|
| 50              | 10                  |
| 150             | 10                  |
| 10              | 5                   |

Arranging orders this way, makes it possible to verify easily if match between orders is possible.
If the top record on the sell side has the price lower than or equal to the price in the top record on buy side it means
match exists and orders might be executed and cleared. If this condition is not met, nothing can be done.

This way, it is very simple to detect if order book needs to be processed. It requires only two records to be read from
the database and one comparison.

In the order book example above, there are two orders on each side, having the same price. So the question is: How should they be sorted?
It is important, because at the time of the order matching, the corresponding order might not match both orders fully, so
one of them needs to be given a priority to be cleared first. In such cases, priority is given to the order which has been
added to the order book earlier.

Putting everything together, it means there are two sort conditions:
1. price (ascending on the sell side, descending on the buy side)
2. time (ascending on both sides)

Sorting by price has a priority, time is always the less meaningful condition. It means that sell order offering lower price
is placed earlier in the selling queue than existing orders offering higher price.
The same way, buy order offering higher price is placed earlier in the buying queue than existing orders offering lower price.

It is important to note, that in case order is updated (amount or price are changed), its time priority is recalculated.
Effectively the update operation is equivalent to deleting the previous order and creating a new one in atomic manner.
In this case, even if the original price is preserved, the updated/recreated order might have a lower priority if other orders
with the same price have existed. Also, if the new price is different, the order will be executed after all the preexisting orders
with the same price.

## Order matching

Order matching is the continuous process of finding the pairs of orders in the order book (one on the sell side and one on the buy side)
offering a combination of the prices triggering the token exchange (when the seller sells and the buyers buy the corresponding tokens
and the ownership of the tokens is changed).

According to what has been discussed already, this is possible if the price in the first item from the sell table is lower than
or equal to the price in the first item from the buy table.

### Equal price - Equal amount cases

Here, cases of matching orders with the exact prices and amounts are discussed. 

Let's take an example of this order book, having only one record on each side:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 300                  | 15                       | 50                  |  10                     |

Because sell price is higher than the buy price, nothing can be done. New buy order comes:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 300                  | 15                       | 300                 | 15                      |
|                      |                          | 50                  | 10                      |

Now the matching is possible, meaning that `300uaaa` might be traded between two market participants. After matching the orders,
the order book is reduced:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
|                      |                          | 50                  | 10                      |

Sell table is empty now, so nothing more can be done.

As a result of matching:
- `300uaaa` are transferred from the seller to the buyer
- `4500ubbb` (`300uaaa * 15ubbb/uaaa`) are transferred from the buyer to the seller

Similar flow is possible if sell order is added to the original order book. This is the initial situation:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 300                  | 15                       | 50                  |  10                     |

Then, new sell order is added:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 50                   | 10                       | 50                  | 10                      |
| 300                  | 15                       |                     |                         |

Then, orders are matched and the order book is reduced:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 300                  | 15                       |                     |                         |

Buy table is empty now, so nothing more can be done.

As a result of matching:
- `500uaaa` are transferred from the seller to the buyer
- `500ubbb` (`50uaaa * 10ubbb/uaaa`) are transferred from the buyer to the seller

### Equal price - Non-equal amount cases

Here, cases of matching orders with the exact prices but different amounts are discussed.

Initial order book:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 300                  | 15                       | 50                  |  10                     |

Then, new buy order comes:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 300                  | 15                       | 400                 | 15                      |
|                      |                          | 50                  | 10                      |

Then, orders are matched and the order book is reduced:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
|                      |                          | 100                 | 15                      |
|                      |                          | 50                  | 10                      |

Because amount in buy order was higher than the amount in sell order, the remaining part stays in the buy table waiting
for another seller.

As a result of matching:
- `300uaaa` are transferred from the seller to the buyer
- `4500ubbb` (`300uaaa * 15ubbb/uaaa`) are transferred from the buyer to the seller

Initial order book:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 300                  | 15                       | 50                  |  10                     |

Then, new buy order comes:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 300                  | 15                       | 200                 | 15                      |
|                      |                          | 50                  | 10                      |

Then, orders are matched and the order book is reduced:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 100                  | 15                       | 50                  | 10                      |

Because amount in buy order was lower than the amount in sell order, the remaining part stays in the sell table waiting
for another buyer.

As a result of matching:
- `200uaaa` are transferred from the seller to the buyer
- `3000ubbb` (`200uaaa * 15ubbb/uaaa`) are transferred from the buyer to the seller

Initial order book:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 300                  | 15                       | 50                  |  10                     |

Then, new sell order is added:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 100                  | 10                       | 50                  | 10                      |
| 300                  | 15                       |                     |                         |

Then, orders are matched and the order book is reduced:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 50                   | 10                       |                     |                         |
| 300                  | 15                       |                     |                         |

Because amount in sell order was higher than the amount in buy order, the remaining part stays in the sell table waiting
for another buyer.

As a result of matching:
- `50uaaa` are transferred from the seller to the buyer
- `500ubbb` (`50uaaa * 10ubbb/uaaa`) are transferred from the buyer to the seller

Initial order book:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 300                  | 15                       | 50                  |  10                     |

Then, new sell order is added:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 25                   | 10                       | 50                  | 10                      |
| 300                  | 15                       |                     |                         |

Then, orders are matched and the order book is reduced:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 300                  | 15                       | 25                  | 10                      |

Because amount in sell order was lower than the amount in buy order, the remaining part stays in the buy table waiting
for another seller.

As a result of matching:
- `25uaaa` are transferred from the seller to the buyer
- `250ubbb` (`25uaaa * 10ubbb/uaaa`) are transferred from the buyer to the seller

## Non-equal price - Non-equal amount cases

Let's take a well-known initial state again:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 300                  | 15                       | 50                  |  10                     |

New buy order comes:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 300                  | 15                       | 400                 | 20                      |
|                      |                          | 50                  | 10                      |

Notice that prices in the first row are not equal. The question is: Can we match them? Let's repeat the definition
of matching possibility again:

Matching is possible if the price in the first record of sell table is **lower than** or equal to the price in the first
record of buy table.

The price in the first record of sell table is `15ubbb/uaaa`, the price in the first record of buy table is `20ubbb/uaaa`.
It means, **the required condition is met**, matching should be done. Buyer wants to buy `400uaaa` but seller offers
`300uaaa` only so, as described in previous examples, only `300uaaa` might be exchanged.

But what exchange rate should be used? There are two possibilities: `15ubbb/uaaa` offered by the seller or `20ubbb/uaaa`
offered by the buyer. The industry-wide practice is that in such situations, the price offered by the earlier-placed order
is applied, so whoever comes later takes the best possible deal (price). It means, in our example tokens are exchanged using
`15ubbb/uaaa` rate.

All of these means that as a result of the order matching:
- `300uaaa` are transferred from the seller to the buyer
- `4500ubbb` (`300uaaa * 15ubbb/uaaa`) are transferred from the buyer to the seller

This is how the order book looks like after the matching:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
|                      |                          | 100                 | 20                      |
|                      |                          | 50                  | 10                      |

Same thing works accordingly if sell order is placed later. Initial order book:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 300                  | 15                       | 50                  |  10                     |

Sell order with non-matching price is added:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 25                   | 5                        | 50                  | 10                      |
| 300                  | 15                       |                     |                         |

Matching is possible because `5ubbb/uaaa` offered by the seller is lower than or equal to `10ubbb/uaaa` offered by the buyer.
Sell order has been placed before buy order so the price `5ubbb/uaaa` is used. Sell amount (`25uaaa`) is lower than buy amount
(`50uaaa`) so only `25uaaa` is exchanged.

As a result of matching:
- `25uaaa` are transferred from the seller to the buyer
- `125ubbb` (`25uaaa * 5ubbb/uaaa`) are transferred from the buyer to the seller

Final state of the order book:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 300                  | 15                       | 25                  | 10                      |

### One-to-one matching

All the examples given so far, were about matching exactly one order taken from each side of the order book (sell and buy).
We call this situation one-to-one matching, which is the simplest possible case.

### One-to-many matching

Let's take an example of this order book:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 300                  | 15                       |                     |                         |
| 100                  | 15                       |                     |                         |
| 50                   | 20                       |                     |                         |

Now, new buy order is placed:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 300                  | 15                       | 500                 | 20                      |
| 100                  | 15                       |                     |                         |
| 50                   | 20                       |                     |                         |

We look at the top records on each side and see that matching is possible (`15ubbb/uaaa` on sell side is lower than `20ubbb/uaaa` on buy side).
It is possible to exchange `300uaaa`. This is what happens:
- `300uaaa` are transferred from the seller to the buyer
- `4500ubbb` (`300uaaa * 15ubbb/uaaa`) are transferred from the buyer to the seller

After the matching, order book looks like this:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 100                  | 15                       | 200                 | 20                      |
| 50                   | 20                       |                     |                         |

We check the matching condition again, and we see that matching is still possible, because `15ubbb/uaaa` offered by the seller
is again lower than `20ubbb/uaaa` offered by the buyer. It is possible to exchange `100uaaa` this time. This is what happens:
- `100uaaa` are transferred from the seller to the buyer
- `1500ubbb` (`100uaaa * 15ubbb/uaaa`) are transferred from the buyer to the seller

After the matching, order book looks like this:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 50                   | 20                       | 100                 | 20                      |

We verify the matching condition again, and it is still met because the price on both sides is equal.
We do order matching again. As a result:
- `50uaaa` are transferred from the seller to the buyer
- `1000ubbb` (`50uaaa * 20ubbb/uaaa`) are transferred from the buyer to the seller

The final state of the order book is:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
|                      |                          | 50                  | 20                      |

It is essential to understand that any single new order may trigger many matches, as far as the offered volume and price
can fulfill many orders existing on the opposite side of the order book.

Thankfully, the one-to-many case might be treated as an extension of one-to-one case:
1. add the new order to the order book
2. repeat one-to-one matching algorithm until no match is possible

### Many-to-many matching

In previous sections, all the examples covered the situations where matching condition is triggered by the exactly one
new incoming order. Depending on the implementation, it might be possible that matching conditions are triggered by many
orders existing in order book.

If order matching is done immediately after adding each order and before adding next one, then this situation cannot happen,
because each order triggers all the possible order book reductions and order executions, leading to a situation where
no more matches are possible, meaning that new match might be triggered only by the next order.

On the other hand, if order matching is not done immediately after collecting each new order, but in batches, then the
situation when there are many orders triggering order executions is possible.

Thankfully, the many-to-many matching might be treated as an extension of one-to-many case (and in turn as an extension
of one-to-one algorithm):
1. Collecting phase: each new incoming order is added to the FIFO queue
2. Execution phase: take one item at a time from FIFO queue, add it to the order book and execute one-to-many matching algorithm, 

## Fund locking

Whenever an order is placed, all the corresponding funds required to execute that order must be locked.
Otherwise, the order owner could spend them after placing the owner leading to a situation where order execution
is impossible.

In case of sell orders we must lock:
- amount of base tokens specified in the order
- all the fees possible fees (if any) related to the base token (**TBD**: how it relates to burn rate and send commission)

In case of buy orders we must lock:
- amount of quote tokens computed as a product of the base token and price specified in the order
- all the fees possible fees (if any) related to the quote token (**TBD**: how it relates to burn rate and send commission)

When orders are matched and executed, locked amounts should be utilized appropriately.
When order is canceled, corresponding amounts should be unlocked.

## Order rejection

To limit the amount of data stored in the order books, orders should be rejected under some circumstances:
- order should be rejected immediately, when placed, if the offered price is too far from the currently traded one
- (**TBD**) each order should have a maximum lifetime, to prune it from the order book if it is not executed for some time
- offered amount is too low to protect us against tones of tiny orders created by spammers (not sure about this)

## Order book creation

There is no such operation on Coreum blockchain. Anyone at any time may create an order book for any pair, just by adding
a first order (sell or buy) for that pair. Only one order book may exist for a pair. There is no concept of an order book
administrator or maintainer. All the pairs are legal, creating an order book does not require any permission or voting.

## Order types

This section describes the order types to be implemented in phase 1.

### Limit order

All the order examples presented so far are the limit orders. It means that:
- in case of sell order - the minimum acceptable price is specified in the order
- in case of buy order - the maximum acceptable price is specified in the order

To match orders and reduce the order book two conditions must be met together for the new order triggering the reduction:
- amount in the order must be greater than 0 - amount in the order is decreased after each reduction,
  so at some point the amount reaches 0, this means that reduction algorithm stops and order is discarded
- price is still acceptable - each reduction causes some order to be removed from the order book (because its amount reaches 0),
  next matching is possible, might be done using next order offering the same or worse price. At some point price
  is no longer acceptable and algorithm is stopped.

### Market order

In this case user does not specify the worst acceptable price, which means that any price is acceptable and order is matched
as long as its amount remains greater than 0, no matter what the price offered by the counterparty order is.
