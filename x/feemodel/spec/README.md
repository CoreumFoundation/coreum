<!--
order: 0
title: Fee model Overview
parent:
  title: "feemodel"
-->

# `x/feemodel`

## Abstract

This document specifies the feemodel module. The module is responsible for calculating minimum gas price required by the chain based on the [parameters](03_params.md) of fee model.

There are four regions on the fee model curve:
 - between 0 and *long average block gas* where gas price goes down exponentially from `InitialGasPrice` to gas price with maximum discount (`InitialGasPrice * (1 -MaxDiscount)`)
 - between *long average block gas* and `EscalationStartBlockGas` where we offer gas price with maximum discount all the time
 - between `EscalationStartBlockGas` and `MaxBlockGas` where price goes up rapidly (being an output of a power function) from gas price with maximum discount to `MaxGasPrice`
 - above `MaxBlockGas` (if it happens for any reason) where price is equal to `MaxGasPrice`

The input (x value) for that function is calculated by taking *short block gas average*.
Price (y value) being an output of the fee model is used as the minimum gas price for the next block.

## Contents

1. **[State](01_state.md)**
2. **[Keeper](02_keeper.md)**
5. **[Parameters](03_params.md)**
