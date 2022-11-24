# Node prerequisites

The document contains common information for any type of node.

* Check the current "ulimit -n"
```
ulimit -n
```

If the limit < 2048, then update it. We recommend to set "65536" as limit.

**Attention!** *This setting is critical for the node, without it the node will crash as soon as it reach the limit.*
