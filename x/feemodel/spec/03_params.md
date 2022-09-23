<!--
order: 3
-->

# Parameters

The feemodel module contains the following parameters:

| Key                     | Type         | Example  |
|-------------------------|--------------|----------|
| InitialGasPrice         | string (dec) | "0.0625" |
| MaxGasPrice             | string (dec) | "62.5"   |
| MaxDiscount             | string (dec) | "0.5"    |
| EscalationStartBlockGas | int64        | 37500000 |
| MaxBlockGas             | int64        | 50000000 |
| ShortEmaBlockLength     | uint32       | 10       |
| LongEmaBlockLength      | uint32       | 1000     |


## InitialGasPrice

`InitialGasPrice` is the minimum gas price required when *block gas short average* is 0. It happens when there are no transactions being broadcasted. This value is also used to initialize gas price on brand-new chain.

## MaxGasPrice

`MaxGasPrice` is the minimum gas price required when *block gas short average* is greater than or equal to `MaxBlockGas`.This value is used to limit gas price escalation to avoid having possible infinity gas price value otherwise.

## MaxDiscount

`MaxDiscount` is th maximum discount we offer on top of `InitialGasPrice` if *short average block gas* is between *long average block gas* and `EscalationStartBlockGas`.

## EscalationStartBlockGas

`EscalationStartBlockGas` defines block gas usage where gas price escalation starts if *short average block gas* is higher than this value.

## MaxBlockGas

`MaxBlockGas` sets the maximum capacity of a block. This is enforced on tendermint level in genesis configuration. Once short average block gas goes above this value, gas price is a flat line equal to `MaxGasPrice`.

## ShortEmaBlockLength

`ShortEmaBlockLength` defines inertia for short average long gas in EMA model. The equation is:

`NewAverage = ((ShortAverageBlockLength - 1)*PreviousAverage + GasUsedByCurrentBlock) / ShortAverageBlockLength`

The value might be interpreted as the number of blocks which are taken to calculate the average. It would be exactly like that in SMA model, in EMA this is an approximation.

## LongEmaBlockLength

`LongEmaBlockLength` defines inertia for long average block gas in EMA model. The equation is:

`NewAverage = ((LongAverageBlockLength - 1)*PreviousAverage + GasUsedByCurrentBlock) / LongAverageBlockLength`

The value might be interpreted as the number of blocks which are taken to calculate the average. It would be exactly like that in SMA model, in EMA this is an approximation.
