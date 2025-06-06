Feature: no match
    There are no matches

  Scenario: no_match_limit_directOB_and_invertedOB_buy_and_sell
    Given there are users with balances:
      | account | balances                    |
      | acc1    | 1000000denom1,1000000denom2 |
      | acc2    |  2659000denom1,375000denom2 |
    And there are orders:
      | id  | creator | base   | quote  | type  | price   | quantity  | side | tif |
      | id1 | acc1    | denom1 | denom2 | limit |  376e-3 | 1_000_000 | sell | gtc |
      | id2 | acc2    | denom1 | denom2 | limit |  375e-3 | 1_000_000 | sell | gtc |
      | id3 | acc1    | denom2 | denom1 | limit |  266e-2 | 1_000_000 | sell | gtc |
      | id4 | acc1    | denom2 | denom1 | limit | 2659e-3 | 1_000_000 | sell | gtc |
    Then no orders are matched
    Then acc1 sends 20denom2 to acc5
    Then acc6 sends 20denom2 to acc5
