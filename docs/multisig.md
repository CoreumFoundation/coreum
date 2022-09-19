# Multisig

The doc describes the **coreum** CLI command for the multisig accounts.

## Set up environment

### Build the cored

The [Build and Play](https://github.com/CoreumFoundation/coreum/blob/master/README.md#build-and-play) doc describes the
process of the cored binary building and installation.

### Set up the CLI environment

Set up the CLI environment following the [doc](cli-env.md).

# Multisig CLI sample

The sample below describes the full flow from the creation of the multisig account to the tx broadcast.

* Generate 3 new signing keys and recipient for testing.

```bash
cored keys add k1 $CORED_CHAIN_ID_ARGS
cored keys add k2 $CORED_CHAIN_ID_ARGS
cored keys add k3 $CORED_CHAIN_ID_ARGS
cored keys add recipient $CORED_CHAIN_ID_ARGS
```

* Generate the multisig account with the 2 signatures threshold.

```bash
cored keys add k1k2k3 --multisig "k1,k2,k3" --multisig-threshold 2 $CORED_CHAIN_ID_ARGS
```

To set up a 2-of-3 multisig, each member must supply their individual public key. In this example we already hold all
keys.

Output example:

```
- name: k1k2k3
  type: multi
  address: devcore13purcatgmnadw3606rcyatmt60ys6e37mcnaar
  pubkey: '{"@type":"/cosmos.crypto.multisig.LegacyAminoPubKey","threshold":2,"public_keys":[{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AijThzC5n3EXBouoOMe18oOxQCl8LnM150ZjAfjCFcFZ"},{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AmARIe1Ki7o1HccCyJyepCIeatmbABolZmSPYCyoSZ49"},{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A9b+BqhWA2TNeE5D1vaTGXjPhF7eGU5tEU+1P9Z2Sy/j"}]}'
  mnemonic: ""
```

The address here is the multisig account address.

* Get the address from the keystore.

```bash
cored keys show --address k1k2k3 $CORED_CHAIN_ID_ARGS
```

* Fund the multisig account from the faucet.

```bash
fund_cored_account $(cored keys show --address k1k2k3 $CORED_CHAIN_ID_ARGS)
```

* Check the multisig account balances.

```bash
cored q bank balances $(cored keys show --address k1k2k3 $CORED_CHAIN_ID_ARGS) $CORED_NODE_ARGS
```

* Generate the json tx to send some coins to recipient.

```bash
cored tx bank send $(cored keys show --address k1k2k3 $CORED_CHAIN_ID_ARGS) $(cored keys show --address recipient $CORED_CHAIN_ID_ARGS) 700$CORED_DENOM \
--from $(cored keys show --address k1k2k3 $CORED_CHAIN_ID_ARGS) \
--gas-prices 1500$CORED_DENOM \
--generate-only $CORED_CHAIN_ID_ARGS > bank-unsigned-tx.json
```

* Check the tx content.

```bash
cat bank-unsigned-tx.json
```

Output example:

```
{"body":{"messages":[{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"devcore13purcatgmnadw3606rcyatmt60ys6e37mcnaar","to_address":"devcore1lyru5pvjymya9xq0rsg406fss45sama8e9dqrs","amount":[{"denom":"dacore","amount":"700"}]}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[{"denom":"dacore","amount":"300000000"}],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}
```

* Sign the tx from k1 account.

```bash
cored tx sign bank-unsigned-tx.json --multisig $(cored keys show --address k1k2k3 $CORED_CHAIN_ID_ARGS) --from k1 --output-document k1sign.json $CORED_CHAIN_ID_ARGS
```

* Add the signature to the json tx.

```bash
cored tx multisign bank-unsigned-tx.json k1k2k3 k1sign.json $CORED_CHAIN_ID_ARGS > bank-signed-tx.json
```

* Try to send partially signed tx to check that it won't pass.

```bash
cored tx broadcast bank-signed-tx.json -y -b block $CORED_NODE_ARGS
```

* Add one more signature.

```bash
cored tx sign bank-unsigned-tx.json --multisig $(cored keys show --address k1k2k3 $CORED_CHAIN_ID_ARGS) --from k2 --output-document k2sign.json $CORED_CHAIN_ID_ARGS
```

* Add the signature to the json tx.

```bash
cored tx multisign bank-unsigned-tx.json k1k2k3 k1sign.json k2sign.json $CORED_CHAIN_ID_ARGS > bank-signed-tx.json
```

* Try to send the tx now.

```bash
cored tx broadcast bank-signed-tx.json -y -b block $CORED_NODE_ARGS
```

* Check the recipient balance.

```bash
cored q bank balances $(cored keys show --address recipient $CORED_CHAIN_ID_ARGS) $CORED_NODE_ARGS
```
