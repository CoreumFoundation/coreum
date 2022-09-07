# Multisig

The doc describes the **coreum** CLI command for the multisig accounts.

# Multisig CLI sample

The sample below describes the full flow of the creation of the multisig account to the tx broadcast.
The commands should be executed using the built **cored** artifact.

* Generate 3 new keys

```
cored keys add k1
cored keys add k2
cored keys add k3
```

* Generate the multisig account with the 2 signatures threshold.

```
cored keys add k1k2k3 --multisig "k1,k2,k3" --multisig-threshold 2
```

To set up a 2-of-3 multisig, each member must supply their individual public key. In this example we already hold all keys.

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

```
cored keys show --address k1k2k3
```

* Send some coins to the multisig account

```
cored tx bank send alice $(cored keys show --address k1k2k3) 10000000000dacore --fees=300000000dacore -y
```

* Check the tx status

```
cored q tx "tx-hash"
```

The code should be 0.

* Check the multisig account balances

```
cored q bank balances $(cored keys show --address k1k2k3)
```

* Generate the json tx to send some coins back to alice

```
cored tx bank send $(cored keys show --address k1k2k3) $(cored keys show --address alice) 700dacore \
--from $(cored keys show --address k1k2k3) \
--fees=300000000dacore \
--generate-only > bank-unsigned-tx.json
```

* Show the tx content

```
cat bank-unsigned-tx.json
```

Output example

```
{"body":{"messages":[{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"devcore13purcatgmnadw3606rcyatmt60ys6e37mcnaar","to_address":"devcore1lyru5pvjymya9xq0rsg406fss45sama8e9dqrs","amount":[{"denom":"dacore","amount":"700"}]}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[{"denom":"dacore","amount":"300000000"}],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}
```

* Sign the tx from k1 account

```
cored tx sign bank-unsigned-tx.json --multisig $(cored keys show --address k1k2k3) --from k1 --output-document k1sign.json
```

* Add the signature to the json tx

```
cored tx multisign bank-unsigned-tx.json k1k2k3 k1sign.json > bank-signed-tx.json
```

* Try to send partially broadcast tx to check that it won't pass

```
cored tx broadcast bank-signed-tx.json
```

* Add one more signature

```
cored tx sign bank-unsigned-tx.json --multisig $(cored keys show --address k1k2k3) --from k2 --output-document k2sign.json
```

* Add the signature to the json tx

```
cored tx multisign bank-unsigned-tx.json k1k2k3 k1sign.json k2sign.json > bank-signed-tx.json
```

* Try to send the tx now

```
cored tx broadcast bank-signed-tx.json
```

* Check the tx status

```
cored q tx "tx-hash"
```

should be 0

* Check the alice balance

```
cored q bank balances $(cored keys show --address alice)
```
