# Node prerequisites

The document contains common information for any type of node.

* Check the current "ulimit -n"
```
ulimit -n
```

If the limit < 2048, then update it. We recommend to set "65536" as limit.

**Attention!** *This setting is critical for the node, without it the node will crash as soon as it reach the limit.*

## Supported architectures and operating systems

Table below contains binary names attached to each release for supported combinations of operating system and architecture.

| \-                     | amd64             | arm64             |
|------------------------|-------------------|-------------------|
| **linux**              | cored-linux-amd64 | cored-linux-arm64 |
