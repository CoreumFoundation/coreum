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

If any two orders in the same table have the same price, they are sorted in ascending order by the time of creation.

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

The formula might be defined as: `price = desired token / offered token`. Now, the `offered token` and `desired token`
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
- order matching is possible if the first item in `(tokenA, tokenB)` queue has the `tokenB amount / tokenA amount` coefficient **lower** than
  or equal to `tokenB amount / tokenA amount` coefficient of the first item in `(tokenB, tokenA)` queue
- order matching is possible if the first item in `(tokenA, tokenB)` queue has the `tokenA amount / tokenB amount` coefficient **higher** than
  or equal to `tokenA amount / tokenB amount` coefficient of the first item in `(tokenB, tokenA)` queue

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

To limit the amount of data stored in the order books, orders should be rejected under some circumstances (**TBD**):
- order should be rejected immediately, when placed, if the offered price is too far from the currently traded one
- each order should have a maximum lifetime, to prune it from the order book if it is not executed for some time
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

This section covers details related to practical implementation of concepts described earlier.

### Order

Order properties:
- ID (`uint64`) - is the unique number identifying the order, it is taken from a sequential generator returning
  consecutive numbers starting from 1 - it is important because 0 represents the fact that ID hasn't been assigned yet.
  ID is assigned only if order must be stored in the persistent store.
- OfferedAmount (`sdk.Coin`) - is the offered coin (it contains amount and denom)
- DesiredAmount (`sdk.Coin`) - is the desired coin (it contains amount and denom)
- Type (`enum: MARKET, LIMIT`) - type of the order

### Denom prefix

In the implementation described later, it is required to use denoms as keys in the store. The denom itself is a long string.
Each order requires two denoms to be concatenated to construct the key (described in next sections). As a result,
a lot of space in the store would be used to keep those keys.

Alternatively, we may apply the same trick Cosmos SDK uses to store balances in the bank module.
Instead of using addresses as keys, sequence number is generated for an account when it receives funds for the first time
and mapping between address and number is maintained separately. This allows to store balances using 64-bit numbers in the store.

So, whenever denom is used on DEX for the first time, same mapping might be created for later use in the queues.

Wherever this doc mentions ***denom prefix*** term, it is related to the structure described above.

### Queue prefix

In the following sections there are places where it is required to organize orders in queues in a way that
store iterator returns orders belonging to the same queue one after another. For these purposes the concatenation of:
1. offered *denom prefix*
2. desired denom *denom prefix*

is used as a prefix for the key.

Wherever this doc mentions ***queue prefix*** term, it is related to the structure described above.

### Sorting by price

Order matching algorithm requires orders to be sorted by price. Both corresponding queues, `(tokenA, tokenB)` and `(tokenB, tokenA)`,
must be sorted by the same deterministic formula. Because two prices might be calculated: `token A amount / token B amount` or
`token B amount / token A amount`, it must be decided which one is used.

The algorithm is:
1. sort the denoms by their *denom prefixes* in ascending order
2. the amount corresponding to the first one goes to the **denominator**
3. the amount corresponding to the remaining one goes to the **nominator**

Assuming, the sequence is: `[tokenA, tokenB]` then the price formula is `tokenB / tokenA`. It means that from the two forms
of the order matching rule defined earlier, the first one is chosen: order matching is possible if the first item in
`(tokenA, tokenB)` queue has the `tokenB amount / tokenA amount` coefficient lower than or equal to `tokenB amount / tokenA amount`
coefficient of the first item in `(tokenB, tokenA)` queue

Important note: As a consequence, whenever someone would like to construct `tokenA/tokenB` order book, it means that queue `(tokenA, tokenB)`
is the sell side of the book and queue `(tokenB, tokenA)` corresponds to the buy side of that order book.
At the same time, whenever someone would like to construct `tokenB/tokenA` order book, it means that queue `(tokenB, tokenA)`
is the sell side of the book and queue `(tokenA, tokenB)` corresponds to the buy side of that order book.

### Execution price prefix

It is required to design the correct schema for storing the price as a part of the key in the store to be able to retrieve
orders of interest in the form of a properly sorted FIFO queue when using store iterator provided by the Cosmos SDK.

We identified two possible ways of encoding the price for our purposes. Later on we will decide which format to use (**TBD**).

Wherever in the doc the term ***execution price prefix*** is referenced, it should be understood as defined above.

Important note: The price defined in this section should never ever be used to make any calculations during the order
matching. When orders are executed and tokens are exchanged, the exchange process should use only the offered and
desired amounts defined in the order structure. Price specified here only defines the sequence used to process
orders by the order matching algorithm.

It might happen that the result of the price calculation goes out of the scope possible to be represented using described format.
In that case:
- we may reject that order
- we may store it using maximum or minimum execution price

Technically, the second option is possible, we just cannot guarantee the proper order execution sequence. But due to
high uncertainty, especially for trading pairs presenting out-of-scope ratio consistently, I recommend rejecting those orders (**TBD**).
Due to that some pairs might be non-tradable on our DEX. As an alternative solution we may research the possibility of using `float64`
format mentioned earlier.

#### Using whole and decimal parts

To store the price as a part of the key, we may use two `uint64` numbers: first to encode the whole part, the second one to encode the decimal part.

For the whole part we may use the full scope of `uint64`, `0..18_446_744_073_709_551_615`, resulting in maximum number
greater than `10^19`.

For the decimal part we are limited to the subscope of `uint64`, `0..9_999_999_999_999_999_999`, fully covering 19 decimal places.

This system gives us the ability to store numbers in the scope `0.0000000000000000000..18446744073709551615.9999999999999999999`,
using 16 bytes. Take a note that maybe (this is huge "maybe" because I didn't do the full analysis) we could use the magic of
`float64` and store the price as a mantissa and exponent, using only 8 bytes at the same time. I may figure it out, but
only if the team decides that benefits are worth the time being spent, as it's not a trivial thing.

Given that format, key prefix might be defined where both numbers are concatenated this way:
- big-endian-encoded whole part
- big endian-encoded decimal part

#### Using fraction and exponent

The concepts described here are similar (but also significant differences are present) to the IEEE 754 standard describing floating-point number format. For more info on that, check: https://en.wikipedia.org/wiki/Double-precision_floating-point_format.

This needs to be further discussed and needs a lot of designing (**TBD**).

### Transient order key

When orders are collected by the message handlers, they are added to the transient FIFO queue to be processed
later in the end blocker. To use that store as a FIFO efficiently, the key for the order record must be built as a concatenation
of these parts:
- offered denom or desired *denom prefix* - whatever is the **first** item after sorting them
- the other *denom prefix* from the pair
- number taken from a transient sequence generator

Keep in mind that the result of concatenating two first items is the same for both corresponding queues: `(tokenA, tokenB)` and `(tokenB, tokenA)`.

Transient sequence generator, is the data structure and function using transient store reseted by Cosmos SDK at the beginning of each block,
returning incremented number for each call.

Wherever this doc mentions ***transient order key***, it is related to the structure described above.

### Persistent order key

When orders are saved to the persistent store, the key is constructed in a very special way, which allows to retrieve
the orders in well-defined sequence when store is iterated. The key is defined as a concatenation of:
- *queue prefix*
- *execution price prefix*
- order *ID* encoded using big endian

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
- orders collected from the transactions in the block

Considering these, this is the algorithm executed for each incoming order (by the message handler):
1. for each new incoming order generate new *transient order key*
2. store the order in the transient store under this key

This way, orders from two related queues (`(tokenA, tokenB)` and `(tokenB, tokenA)`) are mixed together and ordered from
the one received first to the one received last.

### Executing order matching algorithm

Order matching algorithm is executed in the end blocker of the DEX module. At the moment of its execution all the orders
included in a block are present in the transient FIFO queue discussed above.

The sequence below defines the steps to be executed in the scope of the single pair of corresponding order queues.
Whenever new order is fetched from the FIFO queue, not matching the currently processed queue pair, it means that
processing of that pair has finished, results should be stored in the persistent store and processing state for the next
queue tuple should be initialized.

To be able to read the orders from the persistent store, iterator is created for each queue, which
iterates over a prefixed store, defined by the *queue prefix*. Algorithm is designed in the way, where only one
order from persistent store might be required at a time. That order is always retrieved by taking the next value from
the right iterator.

Each time order is processed, the order from the other queue is required to check if match is possible.
That order might already live in the cache (RAM) as a result of the previous processing, or it must be read from the
persistent store with one of the iterators described in the previous paragraph.

The cache is maintained separately for each queue, whenever multiple records must be stored in the cache
they must be stored in the same sequence as used in the persistent store: first by price, then by sequence and whenever
they are read from the cache, it should be done in this sequence. The cache might be imagined as an in-memory working area
of the full queue.

At the end of the queue pair processing the diff must be stored in the persistent store, including:
- deleting cleared orders
- updating amounts in not-fully cleared orders already existing in the persistent store
- adding not-fully cleared new orders 

Each time order is read from the transient FIFO it might belong to `(tokenA, tokenB)` or `(tokenB, tokenA)` queue, the algorithm for each side
is a "mirror" of that defined for the other one. They are covered separately, starting with the ``(tokenA, tokenB)`.

Whenever the processing of the new queue pair starts the initial state is defined as:
- `isTokenATokenBStoreEmpty = false`
- `isTokenBTokenAStoreEmpty = false`

This is how the algorithm is defined for an order related to `(tokenA, tokenB)` queue:
1. Read the next order from FIFO
2. As mentioned above we assume that it belongs to `(tokenA, tokenB)` queue, so the following steps are specific to that scenario
3. Take the top order from the cache of `(tokenB, tokenA)` queue. If cache is empty and `isTokenBTokenAStoreEmpty = true` store the order in the cache and go to (1), otherwise take the top order from the persistent `(tokenB, tokenA)` queue. 
4. If there are no orders (reading failed): store the order in the cache, set `isTokenBTokenAStoreEmpty = true` and go to (1) 
5. Check if the order matching condition between the orders is met. If not, store both orders in cache (if not there yet) and go to (1)
6. Do the reduction (described separately below)
7. If any of the orders (at least one for sure) is completely cleared, remove it from the cache and persistent store (if it is present there), possibly one of the orders might not be completely cleared - it must be added/updated in the cache (but not saved to the persistent store as it still might be cleared later)
8. If the processed order (the one read from FIFO) has been cleared go to (1), if not go to (3)

This is how the algorithm is defined for an order related to `(tokenB, tokenA)` queue:
1. Read the next order from FIFO
2. As mentioned above we assume that it belongs to `(tokenB, tokenA)` queue, so the following steps are specific to that scenario
3. Take the top order from the cache of `(tokenA, tokenB)` queue. If cache is empty and `isTokenATokenBStoreEmpty = true` store the order in the cache and go to (1), otherwise take the top order from the persistent `(tokenA, tokenB)` queue.
4. If there are no orders (reading failed): store the order in the cache, set `isTokenATokenBStoreEmpty = true` and go to (1)
5. Check if the order matching condition between the orders is met. If not, store both orders in cache (if not there yet) and go to (1)
6. Do the reduction (described separately below)
7. If any of the orders (at least one for sure) is completely cleared, remove it from the cache and persistent store (if it is present there), possibly one of the orders might not be completely cleared - it must be added/updated in the cache (but not saved to the persistent store as it still might be cleared later)
8. If the processed order (the one read from FIFO) has been cleared go to (1), if not go to (3)

### Reduction step details (6th step)

At this point of the algorithm it is known that there are two orders from corresponding queue which might be reduced.
It is also known which order was created first. Keep in mind that during the reduction the price offered by the earlier
order must be applied.

Assumptions:
- `uaaaa` denom has *denom prefix* equal to `1`
- `ubbb` denom has *denom prefix* equal to `2`

It means that the formula for the execution price between those denoms is: `price = B / A`, where:
- `B` is the amount of `ubbb` token to be exchanged
- `A` is the amount of `uaaa` token to be exchanged

It is like this because of the definition in the "Sorting by price" section.

#### Trivial case

Let's discuss the obvious case first:
- Alice placed an order: `60uaaa -> 20ubbb`
- Then Bob placed an order: `20ubbb -> 60uaaa`

Reduction step for them is simple, both corresponding amounts are equal so the orders are reduced and cleared completely.
Nothing is left in the order book, even the execution price does not need to be calculated.

#### Non-trivial case

- Alice placed an `uaaa -> ubbb` order
- then Bob placed an `ubbb -> uaaa` order

Exact amounts are not specified intentionally.

Let's play a bit with the order matching rule:

```
priceAlice <= priceBob
aliceOrder.ubbbAmount / aliceOrder.uaaaAmount <= bobOrder.ubbbAmount / bobOrder.uaaaAmount
aliceOrder.ubbbAmount * bobOrder.uaaaAmount <= bobOrder.ubbbAmount * aliceOrder.uaaaAmount
```

It might be concluded that:
- if `aliceOrder.ubbbAmount <= bobOrder.ubbbAmount` and `bobOrder.uaaaAmount <= aliceOrder.uaaaAmount` then the order matching rule is true
- if `aliceOrder.ubbbAmount <= aliceOrder.uaaaAmount` and `bobOrder.uaaaAmount <= bobOrder.ubbbAmount` then the order matching rule is true

It means that for some cases, it is possible to determine the truthfulness of the order matching rule without computing the prices,
avoiding dealing with fractional numbers. But even if both formulas are false, the order matching rule still might be true. 

If the order matching rule works and amounts are not exact we know for sure that exactly one order will be cleared completely.
It means, exactly one of two things must happen:
- `minuaaaAmount = min(orderAlice.uaaaAmount, orderBob.uaaaAmount)` will be sent from Alice to Bob, or
- `minubbbAmount = min(orderAlice.ubbbAmount, orderBob.ubbbmount)` will be sent from Bob to Alice

In both cases, we need to find out the amount of the corresponding token to be sent in the opposite direction.

First case:

```
ubbbAmount = minuaaaAmount * priceAlice
ubbbAmount = minuaaaAmount * orderAlice.ubbbAmount / orderAlice.uaaaAmount
```

It might be found that in this case if `minuaaaAmount = orderAlice.uaaaAmount` then `ubbbAmount = orderAlice.ubbbAmount` without doing any calculation.
If `ubbbAmount <= minubbbAmount` it means that this is the case to apply and the following one does not need to be checked.

Otherwise, we know that the second case is true, but we need to compute the amount: 

```
uaaaAmount = minubbbAmount / priceAlice
uaaaAmount = minubbbAmount / orderAlice.ubbbAmount / orderAlice.uaaaAmount
uaaaAmount = minubbbAmount * orderAlice.uaaaAmount / orderAlice.ubbbAmount
```

It might be found that if `minubbbAmount = orderAlice.ubbbAmount` then `uaaaAmount = orderAlice.uaaaAmount` without doing any calculation.
If `uaaaAmount <= minuaaaAmount` it means that this is the case to apply.

Example:

- Alice placed an order: `60uaaa -> 10ubbb`
- Then Bob placed an order: `20ubbb -> 30uaaa`

Things to be noticed immediately:
- Order matching rule is true because `aliceOrder.ubbAmount < bobOrder.ubbAmount && bobOrder.uaaaAmount <= aliceOrder.uaaaAmount`
- `30uaaa = min(orderAlice.uaaaAmount=60uaaa, orderBob.uaaaAmount=30uaaa)` is the upper limit of `uaaa` to be exchanged
- `10ubbb = min(orderAlice.ubbbAmount=10ubbb, orderBob.ubbbAmount=20ubbb)` is the upper limit of `ubbb` to be exchanged
- Alice's order has been placed first so `priceAlice` will be used to reduce the orders, the fact that Alice's order is first,
  is determined by the sequence orders are fetched from the transient FIFO or persistent store in

Now, it must be decided how many tokens should be exchanged. We know for sure that exactly one order will be cleared completely.
It means that exactly one of the actions will be taken:
- `10ubbb` will be transferred from Bob to Alice
- `30uaaa` will be transferred from Alice to Bob

We need to find the way to determine which case is true.

Because `orderAlice.ubbbAmount <= orderBob.ubbbAmount` (it is the minimum value) then we may check if `orderAlice.uaaaAmount <= orderBob.uaaaAmount`
(if it is a minimum value). It is not so the first case is false.

We must try the other one. `orderAlice.uaaaAmount > orderBob.uaaaAmount` (it is not a minimum) so the simplified rule of getting
`ubbb` token does not apply. We compute it using the formula:

```
ubbbAmount = orderBob.uaaaAmount * orderAlice.ubbbAmount / orderAlice.uaaaAmount
ubbbAmount = 30 * 10 / 60
ubbAmount = 5
```

Important note: Using proportion to calculate the amount should be the last resort if nothing else is possible, to
eliminate roundings. If proportion is used, and the result is fractional, it should be rounded down to the floor.

Finally, we know that as a result of reducing the orders:
- `30uaaa` should be sent from Alice to Bob
- `5ubbb` should be sent from Bob to Alice
- Bob's order might be removed because it is cleared now (Bob got everything he wanted to)
- Alice's oder must be updated and wait for other counterparty order by subtracting corresponding amounts,
  so it is transformed from `60uaaa -> 10uaaa` to `(60-30uaaa) -> (10-5ubbb)` which is `30uaaa -> 5ubbb`

Important note: In the example, the prices in the Alice's order before and after the reduction is the same (`10 / 60 = 5 / 30`)
but due to the math precision it is not true in general case. It means that in the updated order price might be different,
affecting the sequence of orders in the queue. As a result, the updated order must be removed from the cache and persistent
store and added again under new key, recalculated using the algorithm described in the "Sorting by price" and "Execution price prefix"
sections.

### Precision

Let's say that there is an order A: `10uaaa -> 9ubbb`. Now, someone comes and places the order B: `1ubbb -> 1uaaa`.
The order matching condition is met because `9 / 10 <= 1 / 1`. Order B contains smaller amounts, so it will be cleared
using the price defined by order A (`0.9ubbb/uaaa`).

According to the previously defined rules:
- `minubbbAmount = min(9ubbb, 1ubbb) = 1ubbb`
- `minuaaa = min(10uaaa, 1uaaa) = 1uaaa`
- `minubbbAmount != orderA.ubbbAmount`
- `minuaaaAmount != orderA.uaaaAmount`

These mean that the corresponding amount must be calculated from the proportion.

```
ubbbAmount = 1uaaa * 9ubbb / 10uaaa = 0.9uaaa ~= 0uaaa
uaaaAmount = 1ubbb * 10ubbb / 9ubbb = 1.(1)ubbb ~= 1ubbb
```

First value is `0`, so algorithm cannot use it. Second one is positive, so algorithm could continue, but it would use
the execution price worse by more than 11%, than what the creator of the order A asked for. This is unacceptable.

Possible solutions (**TBD**):
- don't execute the orders, keep order B to the queue until someone else matches it more precisely
- require some minimal amounts in the order (like `100`) to guarantee that the price is missed by `1%` in the worst case
- put an additional fee on the order execution going to the order A creator to compensate, and at the same time discourage
  the order B creator from placing it

## To be added

- Good Till Time orders
- Immediate or Cancel orders
- Fill or Kill orders
- Indirect order matching
- Implementation details of fund locking
- Implementation details on Market orders
