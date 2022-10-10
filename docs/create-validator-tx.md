# Create Validator Tx Generation & Signing

This doc describes commands to generate and sign CreateValidator transaction type.
Similar commands could be used for another transaction types.

```bash
moniker="staker1"
mnemonic="couch swallow actual often section guitar guard wealth pig usual used provide token symptom hip novel live panel insect left moon faith argue awake"

# Add mnemonic to local keyring.
echo $mnemonic | cored keys add $moniker --recover

# Generate unsigned CreateValidator transaction.
cored tx staking create-validator \
  --amount=10000000000000ducore \
  --pubkey=$(cored keys show --pubkey $moniker) \
  --moniker=$moniker \
  --details="" \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --from=$(cored keys show --address $moniker) \
  --gas=0 \
  --generate-only > create-validator-$moniker-unsigned.json

# Sign transaction using key stored in keyring.
cored tx sign create-validator-$moniker-unsigned.json \
  --from $moniker \
  --output-document create-validator-$moniker-signed.json \
  --offline \
  --sequence=0 \
  --account-number=0 \
  --chain-id=coreum-devnet-1
```
