
# Synthetic Order book
The decentralized exchange work will be based on the synthetic order book in which users can add new orders. 
An order consists of a **currency pair**, **token** that the user wants to buy, **execution price**, **size** and **additional conditions of execution and closing**. 
A user that created an order will be able to close it before it is fully executed.

Order example:
- **Pair:** USD-BTC
- **Buy:** BTC
- **Execution price**: 1.000 USD
- **Size:** 100 BTC
- **Execution conditions:** Fill or Kill, Good till time (1 hour)

On DEX users will be able to trade all the issued assets including the native asset $CORE.
Orders made in pairs of these currencies will be matched to each other through the synthetic order book. When a user creates a new order, the system will find opposite orders not only in the pair that user chose, but also in other currency pairs, simulating the order executing chain after which the user’s order will be executed.
Example:
A user created a ETH-BTC buy order, but there **are no opposite orders in this pair** through which the user’s order can be executed, 
but **there are two suitable limit orders in ETH-XPR and in XRP-BTC pairs**, executing which **the user will receive the amount of BTC that he wanted for the price that he wanted**.

## Order types

### Fill or Kill
Users can create an order that **must be executed immediately** or it will be closed. 
Orders of this type can be executed by several opposite orders and this order **can be closed even if it was not fully executed**.

### Limit and market orders
There is no technical difference between limit and market orders, 
but creating a market order the system will fill all the order parameters in the way to immediately execute this order at the lowest price.

### Stop loss
Creating an order a user will be able to **create an additional stop loss order** connected to this order with the opposite direction.

* For the "buy" order, the stop loss order must have the price lower than the original order price.
* For the "sell" order the stop loss must have the price higher than the original order price. A user can set the price and time limitation during which an order will exist.
* Stop loss orders can be filled with several opposite orders.

When a **stop loss order is executed, the original order, take profit order and stop loss order will also be closed automatically**.

### Trailing stop
Creating an order a user can set the **trailing stop automanaging order**. The user has to **set the price difference between his order and trailing stop order**.

**For "buy" orders trailing stop price will be lower than the original order.** After a user creates a trailing stop order in addition to the main order, if the price goes up, the trailing stop order price also will increase, keeping the price difference entered by the user. If a price goes down, the trailing stop order will not change and it will be executed when the price reaches the entered difference at prevailing market price.

**For "sell" orders trailing stop price will be lower than the original order.** After a user creates a trailing stop order in addition to the main order, if the price goes down, the trailing stop order price also will decrease, keeping the price difference entered by the user. If a price goes up, the trailing stop order will not change and it will be executed when the price reaches the entered difference at prevailing market price.
When a trailing stop order is executed, the original order, stop loss order and take profit order will also be closed automatically.

### Trailing stop limit
A trailing stop limit order is designed to allow an investor to **specify a limit on the maximum possible loss, without setting a limit on the maximum possible gain**.

A sell trailing stop limit **moves with the market price,** and continually **recalculates the stop trigger price at a fixed amount below the market price**, based on the user-defined "trailing" amount.
The limit order price is also continually recalculated based on the limit offset. As the market price rises, both the stop price and the limit price rise by the trail amount and limit offset respectively, but if the stock price falls, the stop price remains unchanged, and when the stop price is hit a limit order is submitted at the last calculated limit price. 

**A "Buy" trailing stop limit order is the mirror image of a sell trailing stop limit, and is generally used in falling markets.**




