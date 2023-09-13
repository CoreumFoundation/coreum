# Gas estimation procedure

After every upgrade of Cosmos SDK (or whenever we decide) we should verify that our deterministic gas estimations for messages are correct.

To do it:
1. Modify crust to deploy `explorer` and `monitoring` profiles when integration tests are started
2. Run integration tests
3. Go to our Grafana dashboard in `znet`
4. Take values for deterministic gas factors reported there
5. Recalculate the deterministic gas by multiplying it by the minimum value taken from the metric.

If there is huge divergence between min and max values reported, reason should be identified.

For Bank Send message we have integration test `TestBankSendEstimation` used to estimate the gas required by each additional coin present in the message.
