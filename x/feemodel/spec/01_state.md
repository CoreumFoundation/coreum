<!--
order: 1
-->

# State

The `x/feemodel` module at the end of each block computes minimum gas price required by the chain for next block.

State managed by feemodel module:

- MinGasPrice: `0x01 | -> string(minGasPrice)`
- ShortEMAGas: `0x02 | -> int64(shortEMAGas)`
- LongEMAGasKey: `0x03 | -> int64(longEMAGas)`

## MinGasPrice

Minimum gas price required by chain

## ShortEMAGas

Short moving average of gas consumed by previous blocks

## LongEMAGasKey

Long moving average of gas consumed by previous blocks
