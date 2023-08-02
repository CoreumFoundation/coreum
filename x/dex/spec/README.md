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

## Order book mirroring

So far, we discussed that each order book is characterized by the base token and quote token. In the `uaaa/ubbb` pair
`uaaa` is the base token and `ubbb` is the quote token. We might also consider the opposite pair `ubbb/uaaa` where
`ubbb` is the base token and `uaaa` is the quote token.

This way we might imagine two order books `uaaa/ubbb` and `ubbb/uaaa`. Managing both of them at the same is technically
possible but does not make any sense because liquidity is split between them. Also, it would be suboptimal because
arbitrators would create additional transactions on chain to take the opportunities caused by possible market inefficiencies.

Let's consider that there are 4 people who want to trade:
- Alice: wants to exchange `10uaaa` to `5ubbb`
- Bob: wants to exchange `4ubbb` to `1uaaa`
- Charlie: wants to exchange `2uaaa` to `8ubbb`
- Dave: wants to exchange `3ubbb` to `6uaaa`

Let's assume that trading happens using `uaaa/ubbb` order book. We must rewrite the trades in the form of that order book:
- Alice wants to sell `10uaaa` at a price of `0.5ubbb/uaaa`
- Bob wants to buy `1uaaa` at a price of `4ubbb/uaaa`
- Charlie wants to sell `2uaaa` at a price of `4ubbb/uaaa`
- Dave wants to buy `6uaaa` at a price of `0.5ubbb/uaaa`

After adding all those orders to the `uaaa/ubbb` order book it looks like this:

| Sell amount `[uaaa]` | Sell price `[ubbb/uaaa]` | Buy amount `[uaaa]` | Buy price `[ubbb/uaaa]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 10 (Alice)           | 0.5                      | 1 (Bob)             | 4                       |
| 2 (Charlie)          | 4                        | 6 (Dave)            | 0.5                     |

Now, let's consider the same set of desires, expressed at the beginning, but let's assume that trading happens using
`ubbb/uaaa` order book instead. It means, that we must rewrite formulas in the terms of that order book:
- Alice wants to buy `5ubbb` at a price of `2uaaa/ubbb`
- Bob wants to sell `4ubbb` at a price of `0.25uaaa/ubbb`
- Charlie wants to buy `8ubbb` at a price of `0.25uaaa/ubbb`
- Dave wants to sell `3ubbb` at a price of `2uaaa/ubbb`

The resulting `ubbb/uaaa` order book is:

| Sell amount `[ubbb]` | Sell price `[uaaa/ubbb]` | Buy amount `[ubbb]` | Buy price `[uaaa/ubbb]` |
|----------------------|--------------------------|---------------------|-------------------------|
| 4 (Bob)              | 0.25                     | 5 (Alice)           | 2                       |
| 3 (Dave)             | 2                        | 8 (Charlie)         | 0.25                    |

A few very interesting and super-important facts to note after comparing both order books:
- Each order changed the side: buy order from `uaaa/ubbb` is a sell order in `ubbb/uaaa`
- The sequence of orders on corresponding side is the same:
  Alice was first in the sell table of `uaaa/ubbb` and then she is the first in buy table of `ubbb/uaaa`.
  Bob was first in the buy table of `uaaa/ubbb` and then he is the first in sell table of `uaaa/ubbb`
- Both sides of both order books are correctly sorted: sell table is sorted by price in ascending order in both order books,
  buy table is sorted by price in descending order in both order books, meaning that order matching algorithm might be correctly
  executed without any additional steps
- For each order, the formula `orderInOrderBookA.Price = 1 / orderInOrderBookB.Price` works in both directions

All of these might be proven mathematically, but we leave this task as a home assignment to the reader :-).

Now, let's consider another thing. There are three properties of the order:
- base token amount
- execution price
- quote token amount

There is a linear relation connecting them: `quote token = execution price * base token`. This means that `base token`
and `execution price` are the inputs for the algorithm (they are provided by the user) and `quote token` is the output
(result) of the algorithm. This way of reasoning is commonly used by the exchanges in financial world, where user provides
the amount to buy/sell and the worst acceptable execution price, and then, `quote token` amount he/she pays/gets as a result
of the order execution is taken from that formula.

But we may look at the problem from different perspective. Order might be defined by having these properties instead:
- offered token amount to sell
- desired token amount to buy
- price

The formula might be defined as: `price = offered token / desired token`. Now, the `offered token` and `desired token`
are the inputs (provided by user) and price is the output.

So we have two different ways of thinking:
- order book defines the base token and quote token, user must say if his/her order is about buying or selling the base token
- there is no definition of order book, base token, nor quote token, user simply specifies what token he/she offers and what token he/she desires

We decide to implement the second type of thinking. At the same time we need to define a relation between both ways because
the industry standard is to present trading platforms in the form of order books. The intuitive definition has been already provided
at the beginning of the section, showing how the same orders of Alice, Bob, Dave and Charlie has been presented in two complementary
order books. Now we provide formal transformations.

**Case 1:**

Assumptions:
- user offers `Auaaa` but desires `Bubbb`
- `uaaa/ubbb` order book is used

Formulas:
- `base token = A`
- `quote token = B`
- `execution price = B / A`
- this is a `sell` order

**Case 2:**

Assumptions:
- user offers `Auaaa` but desires `Bubbb`
- `ubbb/uaaa` order book is used

Formulas:
- `base token = B`
- `quote token = A`
- `execution price = A / B`
- this is a `buy` order

**Case 3:**

Assumptions:
- user offers `Bubbb` but desires `Auaaa`
- `uaaa/ubbb` order book is used

Formulas:
- `base token = A`
- `quote token = B`
- `execution price = B / A`
- this is a `buy` order

**Case 4:**

Assumptions:
- user offers `Auaaa` but desires `Bubbb`
- `ubbb/uaaa` order book is used

Formulas:
- `base token = B`
- `quote token = A`
- `execution price = A / B`
- this is a `sell` order

Things to notice:
- how everything is switched between 1 and 2, and 3 and 4
- how only order type is changed between 1 and 3, and 2 and 4
- each case is unique in terms of results

The most important outcome of this discussion is that any set of orders defining `uaaa->ubbb` or `ubbb->uaaa` exchange direction
might be transformed into two order books: `uaaa/ubbb` and `ubbb/uaaa`.

It means that by defining orders by the offered and desired amounts, client application may generate both order books from
the same source of data and even let users switch between them. That's why later, in the implementation phase, concept of
order book is not defined. Here is the example state of defined system containing orders made by Alice, Bob, Charlie and Dave:

| `uaaa` -> `ubbb`             | `ubbb` -> `uaaa`          |
|------------------------------|---------------------------|
| `10uaaa` -> `5ubbb` (Alice)  | `4ubbb` -> `1uaaa` (Bob)  |
| `2uaaa` -> `8ubbb` (Charlie) | `3ubbb` -> `6uaaa` (Dave) |

So there are two sequences of orders:
- `uaaa` -> `ubbb` added by people who want to exchange `uaaa` to `ubbb`, they are sorted in ascending order by `ubbb amount / uaaa amount`
- `ubbb` -> `uaaa` added by people who want to exchange `ubbb` to `uaaa`, they are sorted in ascending order by `uaaa amount / ubbb amount`

In general case, order is added to the queue described by the tuplet `(offered token, desired token)`,
and that queue is sorted in ascending order by `desired amount / offered amount`. Depending on possible optimizations the
implemented formula might be different as long as it preserves the sequence defined here.

## Order matching between queues

As defined in the previous section we have two queues described by tuplets: `(tokenA, tokenB)` and `(tokenB, tokenA)`.
The first token in the tuplet is called `offered token`, second one is `desired token`. Now we need to redefine the
order matching rule, previously specified for order books, so it might work with queues.

For the upcoming discussion it is important to note these facts:
- function `f(x) = 1 / x` is always monotonically decreasing
- it means that if `g(x)` is monotonic (increasing or decreasing), then the monotonicity of `f(g(x))` has always the opposite direction than `g(x)`

Assumptions:
- `x` is never `0`
- `g(x)` is never `0`

Proof:
- 1st derivative of `f(x) = 1 / x` is `f'(x) = -1 / x^2`
- from the "chain rule" of derivatives: `[f(g(x))]' = f'(g(x)) * g'(x) = -1 / g(x)^2 * g'(x) = -g'(x)/g(x)^2`
- denominator `g(x)^2` is always positive, so the sign of `f'(x)` is determined only by the nominator `-g'(x)`, meaning that signs of `f'(x)` and `g'(x)` are opposite
- from the definition of derivative: if derivative's value is positive everywhere, it means that the function is monotonically increasing,
  if derivative's value is negative everywhere, it means that the function is monotonically decreasing
- by applying previous 2 points it is proven now that monotonicities of `f(x)` and `g(x)` are opposite

`(tokenA, tokenB)` queue is sorted in **ascending** order by `tokenB amount / tokenA amount`. It means that the first item in that queue
offers the lowest price of **selling** `tokenA` (expressed in `tokenB/tokenA` units). At the same time, as stated in the rule above,
that queue is sorted in **descending** order by `tokenA amount / tokenB amount`. It means that the first item in the queue offers the
highest price of **buying** `tokenB` (expressed in `tokenA/tokenB` units).

`(tokenB, tokenA)` queue is sorted in **ascending** order by `tokenA amount / tokenB amount`. It means that the first item in that queue
offers the lowest price of **selling** `tokenB` (expressed in `tokenA/tokenB` units). At the same time, as stated in the rule above,
that queue is sorted in **descending** order by `tokenB amount / tokenA amount`. It means that the first item in the queue offers the
highest price of **buying** `tokenA` (expressed in `tokenB/tokenA` units).

From these two paragraphs it is possible to construct two understandings of the order matching rule:
- order matching is possible if the first item in `(tokenA, tokenB)` queue has the `tokenB amount / tokenA amount` coefficient lower than
  `tokenB amount / tokenA amount` coefficient of the first item in `(tokenB, tokenA)` queue
- order matching is possible if the first item in `(tokenA, tokenB)` queue has the `tokenA amount / tokenB amount` coefficient higher than
  `tokenA amount / tokenB amount` coefficient of the first item in `(tokenB, tokenA)` queue

Both rules mean exactly the same so only one must be chosen.

Side note: The fact that those two rules are equivalent, proves that it is possible to construct both mirrored order books
from the same two sets of orders.

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

## Implementation details

TODO: A lot of stuff here must be rewritten due to introducing mirrored order books  

This section covers details related to practical implementation of concepts described earlier.

### Order

Order properties:
- ID (`uint64`) - is the unique number identifying the order, it is taken from a sequential generator returning
  consecutive numbers starting from 1 - it is important because 0 represents the fact that ID hasn't been assigned yet.
  ID is assigned only if order must be stored in the persistent store.
- BaseToken (`string`) - is the denom of the base token
- QuoteToken (`string`) - is the denom of the quot token
- BaseTokenAmount (`sdk.Int`) - specifies the amount of the base token to be exchanged
- QuoteTokenAmount (`sdk.Int`) - specifies the amount of the quote token to be exchanged
- Direction (`enum: SELL, BUY`) - `SELL` means that user owns the base tokens and wants to sell them to get the quote tokens,
  `BUY` means that user owns the quote tokens and wants to pay them to buy the base tokens
- Type (`enum: MARKET, LIMIT`) - type of the order

### Denom in store keys

In the following sections there are places where it is required to use a denom as a part of the key in the store.
Whenever it happens, the denom must be prefixed with its length encoded using varint algorithm.

### Order book prefix

In the following sections there are places where it is required to organize orders in the store in a way that
store iterator returns orders belonging to the same order book one after another. For these purposes the concatenation of:
1. length-prefixed base denom
2. length-prefixed quote denom

is used as a prefix for the key. Wherever this doc mentions ***order book prefix*** term, it is related to the structure described above.

### Order book side prefix

When orders are stored persistently, they are grouped in the "buckets" related to a particular table (sell or buy) of the
specific order book. To process order books efficiently it is required to be able to retrieve an order by its table quickly.

For that purpose key prefix format is created being a result of concatenating:
- *order book prefix*
- byte representing the table: `0x01` - sell table and `0x02` - buy table

Wherever in the doc the term ***order book side prefix*** is referenced, it should be understood that way. 

### Execution price prefix

Order matching algorithm requires orders to be sorted by price. In the sell table they must be sorted in ascending order,
in the buy table they must be sorted in descending order. But that's not the only one sorting criterion. The other one
is defined by the incoming sequence of orders and enforced by using the order ID as a suffix in the key.

It is required to design the correct schema for storing the order execution price as a part of the key in the store to
be able to retrieve orders of interest in the form of a FIFO queue when using store iterator provided by the Cosmos SDK.

The execution price is a decimal number. In mathematical terms it is defined as `executionPrice = order.QuoteTokenAmount / order.BaseTokenAmount`.
But the floating point result (encoded as string or whatever)m is not convenient from the perspective of our goals.
Instead, to store it as a part of the key we may use two `uint64` numbers: first to encode
the whole part, the second one to encode the decimal part.

For the whole part we may use the full scope of `uint64`, `0..18_446_744_073_709_551_615`, resulting in maximum number
greater than `10^19`.

For the decimal part we are limited to the subscope of `uint64`, `0..9_999_999_999_999_999_999`, giving us 19 decimal places.

This system gives us the ability to store numbers in the scope `0.0000000000000000000..18446744073709551615.9999999999999999999`,
using 16 bytes. Take a note that maybe (this is huge "maybe" because I didn't do the full analysis) we could use the magic of
`float64` and store the price as a mantissa and exponent getting much higher precision and scope, using only 8 bytes at the same time.
I may figure it out, but only if the team decides that benefits are worth the time being spent, as it's not a trivial thing.

Given that format, key prefix might be defined where both numbers are concatenated this way:
- big-endian-encoded whole part
- big endian-encoded decimal part

This format is fine for the sell table as the ascending iterator will return orders in price-ascending order.
But for the buy table, the ascending iterator must return orders in **price-descending** order. Thankfully there is a trick
enabling it.

If the result of the concatenation defined above is taken and mapped in a way, where we take each byte and subtract it from
`0xFF` (255), then the mapped keys will be sorted by the ascending iterator in the opposite order, than the original ones,
which is exactly what we need.

The same mapped key might be generated in simpler way:
- for the whole part: subtract it from the maximum `uint64` value and encode using big endian
- for the decimal part: subtract it from `10^20-1` (`9_999_999_999_999_999_999`) and encode using big endian

By concatenating th result, the execution price prefix for buy order is created.

Wherever in the doc the term ***execution price prefix*** is referenced, it should be understood as defined above,
for sell and buy tables respectively.

Important note: The price defined in this section should never ever be used to make any calculations during the order
matching. When orders are executed and tokens are exchanged, the exchange process should use only `BaseTokenAmount` and
`QuoteTokenAmount` fields defined in the order structure. Price specified here only defines the sequence used to process
orders by the order matching algorithm.

It might happen that the mathematical result of the price calculation (`executionPrice = order.QuoteTokenAmount / order.BaseTokenAmount`)
goes out of the scope which might be represented using described format. In that case:
- we may reject that order
- we may store it using maximum or minimum execution price

Technically, the second option is possible, we just cannot guarantee the proper order execution defined earlier. But due to
high uncertainty, especially for trading pairs presenting out-of-scope ratio consistently, I recommend rejecting those orders (**TBD**).
Due to that some pairs might be non-tradable on our DEX. As an alternative solution we may research the possibility of using `float64`
format mentioned earlier.

### Transient order key

When orders are collected by the message handlers, they are added to the transient FIFO queue to be processed
later in the end blocker. To use that store as a FIFO efficiently, the key for the order record must be built as a concatenation
of these parts:
- *order book prefix*
- number taken from a transient sequence generator

Transient sequence generator, is the data structure and function using transient store reseted by Cosmos SDK at the beginning of each block,
returning incremented number for each call.

Wherever this doc mentions ***transient order key***, it is related to the structure described above.

### Persistent order key

When orders are saved to the persistent store, the key is constructed in a very special way, which allows to retrieve
the orders in well-defined sequence when store is iterated. The key is defined as a concatenation of:
- *order book side prefix*
- *execution price prefix*
- order ID encoded using big endian

Wherever this doc mentions ***persistent order key***, it is related to the structure described above.

If order ID hasn't been assigned yet (the value of `ID` field in the order structure is `0`) it is taken from the persistent
sequence number generator.

### Order collection

New orders come in the form of a message inside the transaction, it means that the blockchain itself, by definition, sorts
the orders - it is precisely defined which order comes first.

In Cosmos SDK it is possible to create a transient in-memory store, managed by keeper, which is automatically recreated
(pruned) at the beginning of each block.

Two types of information are stored in that transient store:
- state of the sequence number generator - generating incremented number on each request, it starts from 0 on every block
- orders collected from the transaction in the block

Considering these, this is the algorithm executed for each incoming order (by the message handler):
1. for each new incoming order generate new *transient order key*
2. store the order in the transient store under this key

This way, orders are returned by their order books, and in the scope of an order book the orders are returned from the first
to the last one.

### Executing order matching algorithm

Order matching algorithm is executed in the end blocker of the DEX module. At the moment of its execution all the orders
included in a block are present in the transient FIFO queue discussed above.

The structure of the *transient order key* causes that orders are iterated by the order books, and in the scope of each
order book they are iterated in the sequence they accepted by the blockchain.

Given that, the order matching algorithm might be finally defined. The sequence below defines the steps to be executed
in the scope of the single order book. Whenever new order is fetched from the FIFO queue, not matching the currently
processed order book, it means that processing of the previous order book has finished, results should be stored
in the persistent store and processing state for the next order book should be initialized.

To be able to read the orders from the persistent store, iterator is created for both order book tables (sell and buy) which
iterates over a prefixed store, defined by the *order book side prefix*. Algorithm is created in the way, where only one
order from persistent store might be required at a time. That order is always retrieved by taking the next value from
the right iterator.

Each time order is processed, the order from the other side of the order book is required to check if match is possible.
That order might already live in the cache (RAM) as a result of the previous processing, or it must be read from the
persistent store with one of the iterators described in the previous paragraph.
The cache is maintained separately for each side of the order book, whenever multiple records must be stored in the cache
they must be stored in the same sequence as used in the persistent store: first by price, then by sequence and whenever
they are read from the cache, it should be done in this sequence. The cache might be imagined as an in-memory working area
of the full order book.

At the end of the order book processing the diff must be stored in the persistent store, including:
- deleting cleared orders
- updating amounts in not-fully cleared orders already existing in the persistent store
- adding not-fully cleared new orders 

Each time order is read from the transient FIFO it might be a sell or buy order, the algorithm for each side
is a "mirror" of that defined for the other one. They are covered separately, starting with the sell order.

Whenever the processing of the new order book starts the initial state is defined as:
- `isSellStoreEmpty = false`
- `isBuyStoreEmpty = false`
- cache for sell table
- cache for buy table

This is how the algorithm for a sell order is defined:
1. Read the next order from FIFO
2. As mentioned above we assume that this is a sell order, so the following steps are specific to that scenario
3. Take the top buy order from the cache. If cache is empty and `isBuyStoreEmpty = true` store the order in the cache and go to (1), otherwise take the top buy order from the persistent store. 
4. If there are no buy orders (reading failed): store the order in the cache, set `isBuyStoreEmpty = true` and go to (1) 
5. Check if the order matching condition between the sell and buy order is met. If not, store both orders in cache (if not there yet) and go to (1)
6. Do the reduction (described separately below)
7. If any of the orders (at least one for sure) is completely cleared, remove it from the cache and persistent store (if it is present there) and possibly one of the order might not be completely cleared - it must be added/updated in the cache (but not from the persistent store as it still might be cleared later)
8. If the sell order has been cleared go to (1), if not go to (3)

This is how the algorithm for a buy order is defined:
1. Read the next order from FIFO
2. As mentioned above we assume that this is a buy order, so the following steps are specific to that scenario
3. Take the top sell order from the cache. If cache is empty and `isSellStoreEmpty = true` store the order in the cache and go to (1), otherwise take the top sell order from the persistent store.
4. If there are no sell orders (reading failed): store the order in the cache, set `isSellStoreEmpty = true` and go to (1)
5. Check if the order matching condition between the buy and sell order is met. If not, store both orders in cache (if not there yet) and go to (1)
6. Do the reduction (described separately below)
7. If any of the orders (at least one for sure) is completely cleared, remove it from the cache and persistent store (if it is present there) and possibly one of the order might not be completely cleared - it must be added/updated in the cache (but not from the persistent store as it still might be cleared later)
8. If the buy order has been cleared go to (1), if not go to (3)

Checking order matching condition details (5th step):

Let's quote the order matching condition again: Matching is possible if the price in the first record of sell table is lower than or equal to the price in the first
record of buy table.

Because price is not stored directly in the order it must be computed (and possibly cached). The formula is: `executionPrice = order.QuoteTokenAmount / order.BaseTokenAmount`. It means that the result must be computed for each order and compared to verify the order matching condition.

Following simple math we may transform the formula to use multiplication instead of division:

`orderA.QuoteTokenAmount / orderA.BaseTokenAmount (><=)? orderB.QuoteTokenAmount / orderB.BaseTokenAmount => orderA.QuoteTokenAmount * orderB.BaseTokenAmount (><=)? orderB.QuoteTokenAmount * orderA.BaseTokenAmount`

all the numbers are positive so inequality never changes the direction.

Reduction step details (6th step):

At this point of the algorithm it is known that there are buy and sell orders to be reduced. It is also known which order was created first.
Keep in mind that during the reduction the price offered by the earlier order must be applied.

To be continued...
