Feature: no match
  Cases when none of the orders match with each other

  Scenario: no match limit directOB and invertedOB buy and sell
    Given there are users with balances:
      | account | balances                                  |
      | acc1    | 1000000denom1,1000000denom2,2orderReserve |
      | acc2    | 2659000denom1,375000denom2,2orderReserve  |
    And there are orders:
      | id  | creator | base   | quote  | type  | price   | quantity  | side | tif |
      | id1 | acc1    | denom1 | denom2 | limit | 376e-3  | 1_000_000 | sell | gtc |
      | id2 | acc2    | denom1 | denom2 | limit | 375e-3  | 1_000_000 | buy  | gtc |
      | id3 | acc1    | denom2 | denom1 | limit | 266e-2  | 1_000_000 | sell | gtc |
      | id4 | acc2    | denom2 | denom1 | limit | 2659e-3 | 1_000_000 | buy  | gtc |
    Then no orders are matched
    And there will be no available balances
    And there will be remaining orders:
      | id  | creator | base   | quote  | type  | price   | quantity  | side | tif | remaining quantity | remaining balance |
      | id1 | acc1    | denom1 | denom2 | limit | 376e-3  | 1_000_000 | sell | gtc | 1_000_000          | 1_000_000         |
      | id2 | acc2    | denom1 | denom2 | limit | 375e-3  | 1_000_000 | buy  | gtc | 1_000_000          | 375_000           |
      | id3 | acc1    | denom2 | denom1 | limit | 266e-2  | 1_000_000 | sell | gtc | 1_000_000          | 1_000_000         |
      | id4 | acc2    | denom2 | denom1 | limit | 2659e-3 | 1_000_000 | buy  | gtc | 1_000_000          | 2_659_000         |

  Scenario: no match market sell
    Given there are users with balances:
      | account | balances        |
      | acc1    | 0denom1,0denom2 |
    And there are orders:
      | id  | creator | base   | quote  | type   | quantity  | side | tif |
      | id1 | acc1    | denom1 | denom2 | market | 1_000_000 | sell | ioc |
    Then no orders are matched
    And there will be no available balances

  Scenario: no match market buy
    Given there are users with balances:
      | account | balances        |
      | acc1    | 0denom1,0denom2 |
    And there are orders:
      | id  | creator | base   | quote  | type   | quantity  | side | tif |
      | id1 | acc1    | denom1 | denom2 | market | 1_000_000 | buy  | ioc |
    Then no orders are matched
    And there will be no available balances

  Scenario: match limit directOB lack of balance
    Given there are users with balances:
      | account | balances                   |
      | acc1    | 999000denom1,1orderReserve |
    And there are orders:
      | id  | creator | base   | quote  | type  | price  | quantity  | side | tif |
      | id1 | acc1    | denom1 | denom2 | limit | 376e-3 | 1_000_000 | sell | gtc |
    Then expecting error that contains:
    """
    1000000denom1 is not available, available 999000denom1
    """

  Scenario: not fillable orders cancelled right after creation
    Given there are users with balances:
      | account | balances                    |
      | acc1    | 1000000denom2,1orderReserve |
      | acc2    | 1000000denom1,1orderReserve |
    And there are orders:
      | id  | creator | base   | quote  | type  | price  | quantity | side | tif |
      | id1 | acc2    | denom1 | denom2 | limit | 376e-5 | 10_000   | sell | gtc |
      | id2 | acc1    | denom1 | denom2 | limit | 333e-5 | 10_000   | buy  | gtc |
    Then no orders are matched
    And there will be users with balances:
      | account | balances                    |
      | acc1    | 1000000denom2,1orderReserve |
      | acc2    | 1000000denom1,1orderReserve |

  Scenario: partially fillable orders accepted for creation // TODO(v6): Revise this behavior.
    Given there are users with balances:
      | account | balances                  |
      | acc1    | 20000denom1,1orderReserve |
    And there are orders:
      | id  | creator | base   | quote  | type  | price  | quantity | side | tif |
      | id1 | acc1    | denom1 | denom2 | limit | 376e-5 | 20_000   | sell | gtc |
    Then no orders are matched
    And there will be remaining orders:
      | id  | creator | base   | quote  | type  | price  | quantity | side | tif | remaining quantity | remaining balance |
      | id1 | acc1    | denom1 | denom2 | limit | 376e-5 | 20_000   | sell | gtc | 20_000             | 20_000            |
    And there will be no available balances
