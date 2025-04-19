### Why `trader.proto`?

It's a Protocol Buffers (.proto) file that defines:

1. The data structures (OrderSheet, OrderResponse)
2. The gRPC service interface (Trader)

The file is the contract between
* Signal Engine - Generates OrderSheets
* Trader microservice(server) - Which executes them and returns results


### Key Components in trader.proto

```proto
syntax = "proto3"
package trader;

option go_package = "cryptoquant.com/m/gen/traderpb";
```

1. Declares the proto version
2. Organizes the proto under a Go package for namespacing after generation

