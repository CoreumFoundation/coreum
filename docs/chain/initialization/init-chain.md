# Inti chain

This doc describes the steps to initiate a chain from scratch.

*The instruction is tested on alpine3.16.*

* Prepare initial binary.

  * Pass the steps in [prepare initial binary](prepare-initial-binary.md) doc to build the initial binary.

* Run seed nodes.

  * Pass the steps in [run seed](../node/run-seed.md) doc to run the initial seeds, at least 2 seeds are recommended.

* Prepare final binary.

  * Pass the steps in [prepare final binary](prepare-final-binary.md) doc to build the final binary.

  Now the new binary is ready to be used.
  Update the variables with the new binary version in the ["cli env"](../cli-env.md) doc.

* Run initial validators

  * Pass the steps in [run genesis validator](run-genesis-validator.md) doc to run the validators set in the genesis.
