# Example

Create fungible token:

```
coredev-00 tx asset issue-ft ALICE alice devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4 100000 --from alice --fees 100000ducore --gas 200000 --yes
```

Query for balance:

```
coredev-00 q bank balances devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4 --denom ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw
```

Create first airdrop:

```
coredev-00 tx airdrop create ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw 0.5ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw 30 "my airdrop" --from alice --fees 100000ducore --gas 200000 --yes
```

Wait until block 30.

Send tokens to Bob:

```
coredev-00 tx bank send devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4 devcore1qqnhyr59smec46hjh4gcrvwl4z6lsmxzdqmwzc 50000ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw --from alice --fees 100000ducore --gas 200000 --yes
```

Query for Alice's balance:

```
coredev-00 q bank balances devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4 --denom ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw
```

Balance should be 50000.

Create second airdrop:

```
coredev-00 tx airdrop create ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw 0.5ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw 60 "my airdrop" --from alice --fees 100000ducore --gas 200000 --yes
```

Query for pending snapshots:

```
coredev-00 q snapshot pending devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4
```

Wait until block 60.

Query for taken snapshots:

```
coredev-00 q snapshot list devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4
```

Query for airdrops:

```
coredev-00 q airdrop list ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw
```

Claim first airdrop from Alice:

```
coredev-00 tx airdrop claim ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw 0 --from alice --fees 100000ducore --gas 200000 --yes
```

Query for balance:

```
coredev-00 q bank balances devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4 --denom ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw
```

Balance should be 100000.

Claim first airdrop from Alice second time:

```
coredev-00 tx airdrop claim ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw 0 --from alice --fees 100000ducore --gas 200000 --yes
```

Transaction should fail.

Claim second airdrop from Alice:

```
coredev-00 tx airdrop claim ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw 1 --from alice --fees 100000ducore --gas 200000 --yes
```

Query for balance:

```
coredev-00 q bank balances devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4 --denom ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw
```

Balance should be 125000.

Claim first airdrop from Bob:

```
coredev-00 tx airdrop claim ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw 0 --from bob --fees 100000ducore --gas 200000 --yes
```

Transaction should fail.

Query for balance:

```
coredev-00 q bank balances devcore1qqnhyr59smec46hjh4gcrvwl4z6lsmxzdqmwzc --denom ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw
```

Balance should be 50000.

Claim second airdrop from Bob:

```
coredev-00 tx airdrop claim ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw 1 --from bob --fees 100000ducore --gas 200000 --yes
```

Query for balance:

```
coredev-00 q bank balances devcore1qqnhyr59smec46hjh4gcrvwl4z6lsmxzdqmwzc --denom ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw
```

Balance should be 75000.
