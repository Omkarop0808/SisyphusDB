# SisyphusDB: Distributed Key-Value Store

SisyphusDB is a high-performance, distributed key-value store engineered for strong consistency (CP system) and high write throughput. It implements a Log-Structured Merge (LSM) Tree storage engine and utilizes the Raft consensus algorithm to manage replication across a coordinated fleet of nodes.

This project demonstrates the architectural evolution from a simple in-memory map to a fault-tolerant distributed system capable of handling production-grade workloads.

---


##
You can install locally, in dockerized containers or K8s clusters.
[A detailed guide](/INSTALL.md)

## System Architecture

The architecture is composed of three distinct layers, separating network communication, consensus logic, and physical storage.
A detailed explaination: [HERE](docs/ARCHITECHTURE.md)


## Performance Engineering & Benchmarks

The core objective of SisyphusDB is to minimize write latency and maximize throughput through low-level memory optimizations and asynchronous I/O strategies.

### 1. Memory Optimization: Custom Arena Allocator

To address the garbage collection (GC) overhead inherent in Go's standard map implementation, a custom **Bump-Pointer Arena Allocator** was engineered. By replacing standard hashing and bucket lookups with direct memory offset calculations, the system achieves O(1) storage time with zero allocations per operation in the hot path.

**Benchmark Results:**

|**Implementation**| **Latency (ns/op)** |**Throughput (Ops/sec)**| **Allocations/Op** |**Memory/Op**|
|---|---------------------|---|-------------------|---|
|**Standard Map (Baseline)**| 82.21 ns            |~12.16 M| 0                 |176 B|
|**Arena Allocator**| **23.17 ns**        |**~24.86 M**| **0**             |**64 B**|
|**Improvement**| **71.1% Faster**    |**104.4% Increase**| **50% Reduction** |**63.6% Reduction**|

Technical Insight:

Even with zero allocations, the Standard Map is limited by the CPU cost of Murmur3/AES hash functions (approx. 82ns). The Arena allocator bypasses this entirely, using pointer arithmetic to store data. This optimization reduced average write latency by ~59ns per op, which compounds significantly in high-frequency trading or telemetry contexts.

_Results verified on Intel Core i5-12450H._

### 2. Throughput Scaling: The Journey to 2,000 RPS

The system was iteratively optimized to scale write throughput by **23x** against the initial baseline.

- **Phase 1 (Baseline - 100 RPS):** The initial implementation was I/O bound due to synchronous `fsync()` calls on every Raft log entry. Throughput was physically capped by disk rotational latency.

- **Phase 2 (Asynchronous Persistence):** Persistence logic was moved to background workers. While this removed the disk bottleneck, it exposed network stack limitations, resulting in `dial tcp: address already in use` errors due to ephemeral port exhaustion.

- **Phase 3 (Optimization - 1,980 RPS):** To stabilize the system at high loads, two critical changes were implemented:

    1. **Adaptive Micro-Batching:** The replicator loop aggregates writes into 50ms windows, reducing network packet count by 95%.

    2. **TCP Connection Pooling:** Implemented `Keep-Alive` to reuse established connections, eliminating handshake overhead.


Final Load Test Results (Vegeta):

Workload: 10,000 Writes over 5 seconds.

Plaintext

```
Requests      [total, rate]          9999, 2000.31
Latencies     [mean, 99th]           32.627ms, 74.615ms
Success       [ratio]                100.00%
Status Codes  [code:count]           200:9999
```

---

## Reliability & Chaos Engineering

Fault tolerance was validated via Chaos Testing on a 3-node Kubernetes cluster to verify **sub-second leader election** and **Linearizability** under failure conditions.

### Experiment: Leader Failure during Write Load

Scenario: A client sends continuous Write (PUT) requests while the Leader node (kv-0) is forcibly deleted.

Constraint: No split-brain writes allowed; the system must failover automatically.

**Log Analysis Results:**

Plaintext

```
1767014613776,UP    System Healthy
1767014613925,DOWN  Leader Killed (Election Starts)
1767014614062,DOWN  Writes Rejected (Proxy Failed)
1767014614474,UP    New Leader Elected (Write Accepted)
```

**Recovery Metrics:**

- **Total Recovery Time:** 549ms

- **Consistency:** Zero data loss. Follower nodes correctly rejected writes during the election window, preserving strong consistency.


---

## Feature Implementation Status

The feature set targets distributed systems complexity comparable to senior-level engineering requirements.

|**Feature**|**Technical Justification**|**Status**|
|---|---|---|
|**LSM Tree Storage**|High-throughput write engine (vs. B-Trees).|✅ Done|
|**Arena Allocator**|Zero-allocation memory management.|✅ Done|
|**WAL & Crash Recovery**|Durability via `fsync` and replay logic.|✅ Done|
|**SSTables + Sparse Index**|Optimized disk I/O and binary search.|✅ Done|
|**Bloom Filters**|Probabilistic structures to minimize disk reads.|✅ Done|
|**Leveled Compaction**|Mitigation of Write/Read Amplification.|✅ Done|
|**Raft Consensus**|Distributed consistency (CAP Theorem compliance).|✅ Done|
|**gRPC & Protobuf**|Schema-strict internal communication.|✅ Done|
|**Prometheus Metrics**|System observability and telemetry.|✅ Done|
