Feature: direct ob limit matching
  Cases when limit orders from direct order book match with each other

  Scenario: match limit directOB maker sell taker buy close maker
    Given there are users with balances:
      | account | balances                    |
      | acc1    | 1000000denom1,1orderReserve |
      | acc2    | 3761000denom2,1orderReserve |
    And there are orders:
      | id  | creator | base   | quote  | type  | price  | quantity   | side | tif |
      | id1 | acc1    | denom1 | denom2 | limit | 375e-3 | 1_000_000  | sell | gtc |
      | id2 | acc2    | denom1 | denom2 | limit | 376e-3 | 10_000_000 | buy  | gtc |
    Then there will be remaining orders:
      | id  | creator | base   | quote  | type  | price  | quantity   | side | tif | remaining quantity | remaining balance |
      | id2 | acc2    | denom1 | denom2 | limit | 376e-3 | 10_000_000 | buy  | gtc | 9_000_000          | 3_384_000         |
    And there will be users with balances:
      | account | balances                   |
      | acc1    | 375000denom2,1orderReserve |
      | acc2    | 1000000denom1,2000denom2   |

  Scenario: match limit directOB maker sell taker buy close maker same account
    Given there are users with balances:
      | account | balances                                  |
      | acc1    | 1000000denom1,3761000denom2,2orderReserve |
    And there are orders:
      | id  | creator | base   | quote  | type  | price  | quantity   | side | tif |
      | id1 | acc1    | denom1 | denom2 | limit | 375e-3 | 1_000_000  | sell | gtc |
      | id2 | acc1    | denom1 | denom2 | limit | 376e-3 | 10_000_000 | buy  | gtc |
    Then there will be remaining orders:
      | id  | creator | base   | quote  | type  | price  | quantity   | side | tif | remaining quantity | remaining balance |
      | id2 | acc1    | denom1 | denom2 | limit | 376e-3 | 10_000_000 | buy  | gtc | 9_000_000          | 3_384_000         |
    And there will be users with balances:
      | account | balances                                 |
      | acc1    | 1000000denom1,377000denom2,1orderReserve |

  Scenario: match limit directOB maker sell taker buy insufficient funds
  """
  we fill the id1 first, so the used balance from id2 is 1_000_000 * 375e-3 = 375_000
  to fill remaining part we need (10_000_000 - 1_000_000) * 376e-3 = 3_384_000,
  so total expected to send 3_384_00 + 375_000 = 3_759_000
  """
    Given there are users with balances:
      | account | balances                    |
      | acc1    | 1000000denom1,1orderReserve |
      | acc2    | 3758000denom2,1orderReserve |
    And there are orders:
      | id  | creator | base   | quote  | type  | price  | quantity   | side | tif |
      | id1 | acc1    | denom1 | denom2 | limit | 375e-3 | 1_000_000  | sell | gtc |
      | id2 | acc2    | denom1 | denom2 | limit | 376e-3 | 10_000_000 | buy  | gtc |
    Then expecting error that contains:
    """
    3759000denom2 is not available, available 3758000denom2
    """

  Scenario: match limit directOB maker sell taker buy close taker
    Given there are users with balances:
      | account | balances                   |
      | acc1    | 100000denom1,1orderReserve |
      | acc2    | 3760denom2,1orderReserve   |
    And there are orders:
      | id  | creator | base   | quote  | type  | price  | quantity | side | tif |
      | id1 | acc1    | denom1 | denom2 | limit | 375e-3 | 100_000  | sell | gtc |
      | id2 | acc2    | denom1 | denom2 | limit | 376e-3 | 10_000   | buy  | gtc |
    Then there will be remaining orders:
      | id  | creator | base   | quote  | type  | price  | quantity | side | tif | remaining quantity | remaining balance |
      | id1 | acc1    | denom1 | denom2 | limit | 375e-3 | 100_000  | sell | gtc | 90_000             | 90_000            |
    And there will be users with balances:
      | account | balances                           |
      | acc1    | 3750denom2                         |
      | acc2    | 10000denom1,10denom2,1orderReserve |

  Scenario: match limit directOB maker sell taker buy partially fillable taker fully cancelled
  """
  not fully fillable since 20_000 * 376e-5 = 75.2, but 12_500 * 376e-5 = 47.
  However, using maker price it is fillable 20_00 * 375e-5 = 75

  Remaining base quantity = 100k - 20k
  Remaining spendable balance = 100k - 20k
  """
    Given there are users with balances:
      | account | balances                   |
      | acc1    | 100000denom1,1orderReserve |
      | acc2    | 100denom2,1orderReserve    |
    And there are orders:
      | id  | creator | base   | quote  | type  | price  | quantity | side | tif |
      | id1 | acc1    | denom1 | denom2 | limit | 374e-5 | 100_000  | sell | gtc |
      | id2 | acc2    | denom1 | denom2 | limit | 376e-5 | 20_000   | buy  | gtc |
    Then there will be remaining orders:
      | id  | creator | base   | quote  | type  | price  | quantity | side | tif | remaining quantity | remaining balance |
      | id1 | acc1    | denom1 | denom2 | limit | 374e-5 | 100_000  | sell | gtc | 100_000            | 100_000           |
    And there will be users with balances:
      | account | balances                |
      | acc2    | 100denom2,1orderReserve |

  Scenario: match limit directOB maker sell taker buy partially matchable taker filled fully
  """
  Remaining base quantity = 100k - 10k
  Remaining spendable balance = 100k - 10k
  """
    Given there are users with balances:
      | account | balances                   |
      | acc1    | 100000denom1,1orderReserve |
      | acc2    | 100denom2,1orderReserve    |
    And there are orders:
      | id  | creator | base   | quote  | type  | price  | quantity | side | tif |
      | id1 | acc1    | denom1 | denom2 | limit | 375e-5 | 100_000  | sell | gtc |
      | id2 | acc2    | denom1 | denom2 | limit | 376e-5 | 20_000   | buy  | gtc |
    Then there will be remaining orders:
      | id  | creator | base   | quote  | type  | price  | quantity | side | tif | remaining quantity | remaining balance |
      | id1 | acc1    | denom1 | denom2 | limit | 375e-5 | 100_000  | sell | gtc | 80_000             | 80_000            |
    And there will be users with balances:
      | account | balances                           |
      | acc1    | 75denom2                           |
      | acc2    | 20000denom1,25denom2,1orderReserve |

  Scenario: match limit directOB maker buy taker sell close maker
    Given there are users with balances:
      | account | balances                   |
      | acc1    | 3760denom2,1orderReserve   |
      | acc2    | 100000denom1,1orderReserve |
    And there are orders:
      | id  | creator | base   | quote  | type  | price  | quantity | side | tif |
      | id1 | acc1    | denom1 | denom2 | limit | 376e-3 | 10_000   | buy  | gtc |
      | id2 | acc2    | denom1 | denom2 | limit | 375e-3 | 100_000  | sell | gtc |
    Then there will be remaining orders:
      | id  | creator | base   | quote  | type  | price  | quantity | side | tif | remaining quantity | remaining balance |
      | id2 | acc2    | denom1 | denom2 | limit | 375e-3 | 100_000  | sell | gtc | 90_000             | 90_000            |
    And there will be users with balances:
      | account | balances                  |
      | acc1    | 10000denom1,1orderReserve |
      | acc2    | 3760denom2                |

  Scenario: match limit directOB maker buy taker sell insufficient funds
    Given there are users with balances:
      | account | balances                  |
      | acc1    | 3760denom2,1orderReserve  |
      | acc2    | 99999denom1,1orderReserve |
    And there are orders:
      | id  | creator | base   | quote  | type  | price  | quantity | side | tif |
      | id1 | acc1    | denom1 | denom2 | limit | 376e-3 | 10_000   | buy  | gtc |
      | id2 | acc2    | denom1 | denom2 | limit | 375e-3 | 100_000  | sell | gtc |
    Then expecting error that contains:
    """
    100000denom1 is not available, available 99999denom1
    """

  Scenario: match limit directOB maker buy taker sell close taker
  """
  Remaining base quantity = 100k - 10k
  Remaining spendable balance = 376e-3 * 90_000 = 33840
  """
    Given there are users with balances:
      | account | balances                  |
      | acc1    | 37600denom2,1orderReserve |
      | acc2    | 10000denom1,1orderReserve |
    And there are orders:
      | id  | creator | base   | quote  | type  | price  | quantity | side | tif |
      | id1 | acc1    | denom1 | denom2 | limit | 376e-3 | 100_000  | buy  | gtc |
      | id2 | acc2    | denom1 | denom2 | limit | 375e-3 | 10_000   | sell | gtc |
    Then there will be remaining orders:
      | id  | creator | base   | quote  | type  | price  | quantity | side | tif | remaining quantity | remaining balance |
      | id1 | acc1    | denom1 | denom2 | limit | 376e-3 | 100_000  | buy  | gtc | 90_000             | 33_840            |
    And there will be users with balances:
      | account | balances                 |
      | acc1    | 10000denom1              |
      | acc2    | 3760denom2,1orderReserve |

  Scenario: match limit directOB maker buy taker sell close taker with same price
  """
  Remaining base quantity = 100k - 10k
  Remaining spendable balance = 375e-3 * 10000 - 375e-3 * 1000 = 3375
  """
    Given there are users with balances:
      | account | balances                  |
      | acc1    | 37500denom2,1orderReserve |
      | acc2    | 10000denom1,1orderReserve |
    And there are orders:
      | id  | creator | base   | quote  | type  | price  | quantity | side | tif |
      | id1 | acc1    | denom1 | denom2 | limit | 375e-3 | 100_000  | buy  | gtc |
      | id2 | acc2    | denom1 | denom2 | limit | 375e-3 | 10_000   | sell | gtc |
    Then there will be remaining orders:
      | id  | creator | base   | quote  | type  | price  | quantity | side | tif | remaining quantity | remaining balance |
      | id1 | acc1    | denom1 | denom2 | limit | 375e-3 | 100_000  | buy  | gtc | 90_000             | 33_750            |
    And there will be users with balances:
      | account | balances                 |
      | acc1    | 10000denom1              |
      | acc2    | 3750denom2,1orderReserve |

  Scenario: match limit directOB maker sell taker buy close both
    Given there are users with balances:
      | account | balances                   |
      | acc1    | 100000denom1,1orderReserve |
      | acc2    | 50000denom2,1orderReserve  |
    And there are orders:
      | id  | creator | base   | quote  | type  | price | quantity | side | tif |
      | id1 | acc1    | denom1 | denom2 | limit | 5e-1  | 100_000  | sell | gtc |
      | id2 | acc2    | denom1 | denom2 | limit | 5e-1  | 100_000  | buy  | gtc |
    Then there will be no remaining orders
    And there will be users with balances:
      | account | balances                   |
      | acc1    | 50000denom2,1orderReserve  |
      | acc2    | 100000denom1,1orderReserve |

  Scenario: match limit directOB maker sell taker buy close both
  """
  "id1" and "id2" orders don't match
  "id3" will match the "id1" and "id2" cover them fully and the remainder will be returned to the creator's balance
  """
    Given there are users with balances:
      | account | balances                  |
      | acc1    | 50000denom1,1orderReserve |
      | acc2    | 50000denom1,1orderReserve |
      | acc3    | 60000denom2,1orderReserve |
    And there are orders:
      | id  | creator | base   | quote  | type  | price | quantity | side | tif |
      | id1 | acc1    | denom1 | denom2 | limit | 5e-1  | 50_000   | sell | gtc |
      | id2 | acc2    | denom1 | denom2 | limit | 5e-1  | 50_000   | sell | gtc |
      | id3 | acc3    | denom1 | denom2 | limit | 6e-1  | 100_000  | buy  | gtc |
    Then there will be no remaining orders
    And there will be users with balances:
      | account | balances                               |
      | acc1    | 25000denom2,1orderReserve              |
      | acc2    | 25000denom2,1orderReserve              |
      | acc3    | 100000denom1,10000denom2,1orderReserve |

  Scenario: match limit directOB close two makers buy and and taker sell
  """
  "id1" and "id2" orders don't match
  "id3" closes "id1" and "id2", with better price for the "id3", expected to receive 80, but receive 100
  """
    Given there are users with balances:
      | account | balances                   |
      | acc1    | 50000denom2,1orderReserve  |
      | acc2    | 50000denom2,1orderReserve  |
      | acc3    | 200000denom1,1orderReserve |
    And there are orders:
      | id  | creator | base   | quote  | type  | price | quantity | side | tif |
      | id1 | acc1    | denom1 | denom2 | limit | 5e-1  | 100_000  | buy  | gtc |
      | id2 | acc2    | denom1 | denom2 | limit | 5e-1  | 100_000  | buy  | gtc |
      | id3 | acc3    | denom1 | denom2 | limit | 4e-1  | 200_000  | sell | gtc |
    Then there will be no remaining orders
    And there will be users with balances:
      | account | balances                   |
      | acc1    | 100000denom1,1orderReserve |
      | acc2    | 100000denom1,1orderReserve |
      | acc3    | 100000denom2,1orderReserve |

  Scenario: match limit directOB multiple maker buy taker sell close taker with same price fifo priority
  """
  acc1 has 296800denom2 = 75_400+75_200+71_000+75_200
  order with id3 remains unmatched price is too low.
  order with id4 the part of the order should remain. Order sequence respected.
  executed id1: 200k*0.377 id2: 200k*0.376 and id4(partially): 100k*0.376
  """
    Given there are users with balances:
      | account | balances                   |
      | acc1    | 296800denom2,4orderReserve |
      | acc2    | 500000denom1,1orderReserve |
    And there are orders:
      | id  | creator | base   | quote  | type  | price  | quantity | side | tif |
      | id1 | acc1    | denom1 | denom2 | limit | 377e-3 | 200_000  | buy  | gtc |
      | id2 | acc1    | denom1 | denom2 | limit | 376e-3 | 200_000  | buy  | gtc |
      | id3 | acc1    | denom1 | denom2 | limit | 355e-3 | 200_000  | buy  | gtc |
      | id4 | acc1    | denom1 | denom2 | limit | 376e-3 | 200_000  | buy  | gtc |
      | id5 | acc2    | denom1 | denom2 | limit | 37e-2  | 500_000  | sell | gtc |
    Then there will be remaining orders:
      | id  | creator | base   | quote  | type  | price  | quantity | side | tif | remaining quantity | remaining balance |
      | id3 | acc1    | denom1 | denom2 | limit | 355e-3 | 200_000  | buy  | gtc | 200_000            | 71_000            |
      | id4 | acc1    | denom1 | denom2 | limit | 376e-3 | 200_000  | buy  | gtc | 100_000            | 37_600            |
    And there will be users with balances:
      | account | balances                   |
      | acc1    | 500000denom1,2orderReserve |
      | acc2    | 188200denom2,1orderReserve |

  Scenario: match limit directOB multiple maker sell taker buy close taker with same price fifo priority
  """
  acc1 has 700000denom1 = 200_000+200_000+100_000+200_000
  order with id3 remains unmatched price is too high.
  order with id4 the part of the order should remain.
  """
    Given there are users with balances:
      | account | balances                   |
      | acc1    | 700000denom1,4orderReserve |
      | acc2    | 189000denom2,1orderReserve |
    And there are orders:
      | id  | creator | base   | quote  | type  | price  | quantity | side | tif |
      | id1 | acc1    | denom1 | denom2 | limit | 376e-3 | 200_000  | sell | gtc |
      | id2 | acc1    | denom1 | denom2 | limit | 375e-3 | 200_000  | sell | gtc |
      | id3 | acc1    | denom1 | denom2 | limit | 399e-3 | 100_000  | sell | gtc |
      | id4 | acc1    | denom1 | denom2 | limit | 376e-3 | 200_000  | sell | gtc |
      | id5 | acc2    | denom1 | denom2 | limit | 378e-3 | 500_000  | buy  | gtc |
    Then there will be remaining orders:
      | id  | creator | base   | quote  | type  | price  | quantity | side | tif | remaining quantity | remaining balance |
      | id3 | acc1    | denom1 | denom2 | limit | 399e-3 | 100_000  | sell | gtc | 100_000            | 100_000           |
      | id4 | acc1    | denom1 | denom2 | limit | 376e-3 | 200_000  | sell | gtc | 100_000            | 100_000           |
    And there will be users with balances:
      | account | balances                              |
      | acc1    | 187800denom2,2orderReserve            |
      | acc2    | 500000denom1,1200denom2,1orderReserve |
