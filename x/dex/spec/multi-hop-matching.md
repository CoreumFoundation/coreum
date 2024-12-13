# Multi-hop Matching

Multi-hop matching facilitates the seamless execution of an order by matching it with self/opposite entries in the order
book or through a multi-step route involving intermediary assets. This mechanism is crucial for ensuring liquidity and
enabling complex trading scenarios.

---

## Example Walkthrough

Consider the following orders in the order book:

| Order ID | Account | Base Denom | Quote Denom | Side | Remaining Quantity | Remaining Balance | Price |
|----------|---------|------------|-------------|------|--------------------|-------------------|-------|
| id1      | acc1    | denom1     | denom2      | sell | 1000000denom1      | 1000000denom1     | 0.4   |
| id2      | acc2    | denom2     | denom1      | buy  | 4000000denom2      | 9200000denom1     | 2.3   |
| id3      | acc3    | denom3     | denom2      | sell | 2000000denom3      | 2000000denom3     | 0.3   |
| id4      | acc4    | denom3     | denom1      | buy  | 1000000denom3      | 1100000denom1     | 1.1   |
| id5      | acc1    | denom1     | denom2      | buy  | 10000000denom1     | 5000000denom2     | 0.5   |

* Self OB

- Matches for `denom1/denom2` is `id1` with a price of **0.4**.

* Opposite OB

- Matches for `denom2/denom1` is `id2` with a price of **1/2.3 ~= 0.434** .

* Multi-hop

- Route: `denom2 → denom3 → denom1`.
    - Step 1 `denom2 → denom3` id3
    - Step 2: `denom3 → denom1` id4
    - Effective rate: `0.3 / 1.1 = 0.27 denom1`.
  
* Find best:
  **0.27** < **0.4** < **0.434**" so the multi-hop is the best match.

* Executing multi-hop - compute maximum quantity:
    * Iteration 1:
        * id3
            * synthetic order is: denom3/denom2, buy 2000000denom3 with price 0.3
            * spend amount 2000000 * 0.3 = 600000denom2, received 2000000denom3,
            * 600000denom2 < 5000000denom2 (we can fill it)
        * id4
            * synthetic order is: denom3/denom1, sell 1000000denom3 with price 1.1
            * spend amount 1000000denom3, received 1000000 * 1.1 = 1100000denom1
            * 2000000denom3 > 1000000denom3 (from prev round, so we need only 2000000 - 1000000 = 1000000denom3)
    * Iteration 2 (adjust id3 synthetic):
        * id3
            * synthetic order is: denom3/denom2, buy 1000000denom3 with price 0.3
            * spend amount 1000000 * 0.3 = 300000denom2, received 1000000denom3,
            * 300000denom2 < 5000000denom2 (we can fill it)
        * id4
            * synthetic order is: denom3/denom1, sell 1000000denom3 with price 1.1
            * spend amount 1000000denom3, received 1000000 * 1.1 = 1100000denom1
            * 1000000denom3 == 1000000denom3 (from prev round, so we will close this order)
        * Execution
            * acc5 send 300000denom2 to acc3
            * acc3 send 1000000denom3 to acc5
            * id3 quantity is reduced by 2000000denom3 = 2000000 - 1000000 = 1000000denom3
            * id3 balance is reduced by 2000000denom3 = 2000000 - 1000000 = 1000000denom3
            * acc5 send 1000000denom3 to acc4
            * acc4 send 1100000denom1 to acc5
            * id4 is closed
            * id5 quantity is reduced by 1100000denom1 = 10000000 - 1100000 = 8900000denom1
            * id5 balance is reduced by 2000000denom3 = 5000000 - 300000 = 4700000denom2
            * check price: 300000 / 1100000 ~= 0.27

Got order book state after multi-hop execution:

| Order ID | Account | Base Denom | Quote Denom | Side | Remaining Quantity | Remaining Balance | Price |
|----------|---------|------------|-------------|------|--------------------|-------------------|-------|
| id1      | acc1    | denom1     | denom2      | sell | 1000000denom1      | 1000000denom1     | 0.4   |
| id2      | acc2    | denom2     | denom1      | buy  | 4000000denom2      | 9200000denom1     | 2.3   |
| id3      | acc3    | denom3     | denom2      | sell | 1000000denom3      | 1000000denom3     | 0.3   |
| id5      | acc1    | denom1     | denom2      | buy  | 8900000denom1      | 4700000denom2     | 0.5   |

* Self OB

- Matches for `denom1/denom2` is `id1` with a price of **0.4**.

* Opposite OB

- Matches for `denom2/denom1` is `id2` with a price of **1/2.3 ~= 0.434** .

* Multi-hop

- Route: `denom2 → denom3 → denom1` - no available orders for the rote

* Find best:
  **0.4** < **0.434**" so the self OB is the best match and next opposite OB.
  ... execution based on the [README.md](README.md) 


### Algorithm in Pseudocode

```
until order is not filled or no matches available:
    order_denoms_mf = get_matching_finder(order.base_denom, order.quote_denom)
    order_denoms_mf_best_match = order_denoms_mf.find_best_match(order.side.inverse())
    rote_steps_best_match = []
    for i in range(len(routes) - 1):
        rote_mf = get_matching_finder(routes[i], routes[i + 1])
        rote_mf_best_match = rote_mf.find_best_match(order.side.inverse())
        step_best_match = matching_finder.find_best_order(order.side.inverse())
        rote_steps_best_match.append(step_best_match)
    best_price, rote_orders = build_rote_steps_best_match_orders(rote_steps_best_match)
    execute_plan(rote_orders) // add the results to the matching result be applied later
```

## Draft Message API

### Message Structure

```
MsgPlaceOrder {
    string sender;            
    OrderType type;           
    string id;                
    string base_denom;         
    string quote_denom;      
    string price;            
    string quantity;         
    Side side;               
    GoodTil good_til;       
    TimeInForce time_in_force; 
    ExecutionPlan execution_plan (optional);
}

ExecutionPlan {
  // if not empty 
  //    must start with the denom you spend and end with the denom you receive
  //    all items must be unique
  //    e.g. : [denom1, ucore, denom2], 
  route [] 
}

```

---

## Draft Query API

The query API is used to evaluate potential execution route without placing an order.

### Query Structure

```
QueryFindExecutionRouteRequest {
    OrderType type;                 
    string base_denom;      
    string quote_denom;     
    string price;           
    string quantity;        
    Side side;      
}
```

```
QueryFindExecutionRoteResult {
    rote[]; // array of denoms
}
```

The node will have config for the max number of denoms and default value eq to 0.

---

## Challenges and Considerations

* Liquidity Optimization

- Ensure the matching algorithm accounts for the maximum liquidity available at each step in a route.
- Adjust initial order quantity based on bottleneck steps in the route.

* Order ID Management

- Uniquely generate or manage order IDs to avoid conflicts.
- Provide meaningful identifiers for the client to interpret results.

* Route Selection

- Evaluate trade-offs between the best price and highest fill probability.
- Consider user preferences for execution strategy (e.g., partial fills, price tolerance).

* Future Enhancements

- Support for multiple user-defined routes.
- Dynamic prioritization based on market conditions.

# To discuss

1. Auto route through the core (asset FT rules issues)?
2. We support now one rote later can support multiple.
3. Discuss the release of the DEX (with or without the multi-hop-matching feature)
