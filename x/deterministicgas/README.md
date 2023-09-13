# Gas estimation procedure

After every upgrade of Cosmos SDK (or whenever we decide) we should verify that our deterministic gas estimations for messages are correct.

To do it:
1. Go to our Grafana dashboard
2. Take values for deterministic gas factors reported there
3. Recalculate the deterministic gas by multiplying it by the minimum value taken from the metric.

If there is huge divergence between min and max values reported, reason should be identified.

For Bank Send message we have integration test `TestBankSendEstimation` used to estimate the gas required by each additional coin present in the message.
