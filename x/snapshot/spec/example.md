# Example

Create fungible token:

```
coredev-00 tx asset issue-ft ALICE alice devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4 10000 --from alice --fees 100000ducore --gas 200000 --yes
```

Query for balance:

```
coredev-00 q bank balances devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4 --denom ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw
```

Create airdrop:

```
coredev-00 tx asset airdrop-ft ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw 1ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw 100 "my airdrop" --from alice --fees 100000ducore --gas 200000 --yes
```

Query for pending snapshots:

```
coredev-00 q snapshot pending devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4
```

Query for taken snapshots:

```
coredev-00 q snapshot list devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4
```

Query for airdrops:

```
coredev-00 q asset airdrops-ft ALICE-devcore1u6dycnl606n95ggeatusc3zlfd5m4xqpw66et4-KTAw
```
