# Chord DHT Implementation in Go

A distributed hash table (DHT) implementation based on the Chord protocol, built using ProtoActor-Go for actor-based concurrency.

## Features

- **Distributed Key-Value Storage**: Store and retrieve data across a distributed network of nodes
- **Fault Tolerance**: Maintains successor lists for resilience against node failures
- **Dynamic Node Joining**: New nodes can join the network and be properly integrated
- **Periodic Stabilization**: Automatic maintenance of the Chord ring structure
- **Real Hashing**: Uses SHA-1 for 160-bit node and key identifiers
- **Networking Support**: Built-in remote actor communication for distributed deployment

## Architecture

This implementation uses the actor model with ProtoActor-Go to represent each Chord node as an independent actor. Key components:

- **ChordNode**: The main actor handling all Chord protocol operations
- **Finger Tables**: For efficient key lookups (O(log N) hops)
- **Successor Lists**: For fault tolerance (maintains multiple successors)
- **Stabilization Protocol**: Periodic maintenance to keep the ring consistent

## Installation

1. Ensure you have Go 1.19+ installed
2. Clone the repository:
   ```bash
   git clone https://github.com/anayy09/Chord-DHT.git
   cd Chord-DHT
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```

## Usage

### Running the Simulation

```bash
make run
# or
go run ./cmd/chord
```

This will start a simulation with multiple nodes, perform key-value operations, and demonstrate fault tolerance by stopping a node.

### Key Operations

- **Store**: Insert key-value pairs into the DHT
- **Lookup**: Retrieve values by key
- **Join**: Add new nodes to the network
- **Stabilize**: Maintain ring structure (runs periodically)

### Configuration

- **mBits**: Set to 160 for production (currently 160)
- **Successor List Size**: Maintains up to 5 successors
- **Stabilization Interval**: Every 5 seconds

## Testing

The implementation includes comprehensive testing:

- Multiple node joins and ring formation
- Key storage and retrieval across nodes
- Fault tolerance simulation (node failures)
- Periodic stabilization verification

## Dependencies

- [ProtoActor-Go](https://github.com/asynkron/protoactor-go): Actor framework for Go
- Standard library: crypto/sha1, math/big, os, time

## References

- Stoica, I., Morris, R., Karger, D., Kaashoek, M. F., & Balakrishnan, H. (2001). Chord: A scalable peer-to-peer lookup service for internet applications. ACM SIGCOMM Computer Communication Review, 31(4), 149-160.