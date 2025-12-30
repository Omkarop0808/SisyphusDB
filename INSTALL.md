# SisyphusDB Installation & Setup Guide

This guide details how to build, deploy, and test SisyphusDB in local, Docker, and Kubernetes environments.

## 🛠️ Prerequisites
* **Go:** 1.25+
* **Docker:** 29.1+
* **Kubernetes:** Minikube, Kind, or a Cloud Provider (EKS/GKE)
* **Python 3:** (Required only for chaos testing scripts)

---

## 🚀 Option 1: Local Development (Standalone)

Run a single node for development or debugging.

### 1. Build the Binary
```
go mod download
go build -o kv-server cmd/server/main.go
2. Run a Node
Bash

# Usage: ./kv-server -id <node_id> -port <raft_port> -http <client_port> -peers <peer_list>
./kv-server -id 0 -port 5001 -http 8001
🐳 Option 2: Docker Compose (Local Cluster)
Spin up a 3-node cluster locally for integration testing.
```
1. Start the Cluster
```
docker-compose up --build -d
```
2. Verify Connectivity
The cluster exposes the following endpoints:

```
Node 0: http://localhost:8001
Node 1: http://localhost:8002
Node 2: http://localhost:8003
```
Test a write operation:

```
curl "http://localhost:8001/put?key=test&val=success"
```
☸️ Option 3: Kubernetes Deployment (Production-Like)
Deploy SisyphusDB as a StatefulSet with stable network identities.

1. Build & Load Image
If using Minikube or Kind, load the image directly to the node cache:

```
docker build -t sisyphusdb:v1 .
minikube image load sisyphusdb:v1
```
2. Apply Manifests
Deploy the Headless Service and StatefulSet:

```
kubectl apply -f deploy/service.yaml
kubectl apply -f deploy/statefulset.yaml
```
3. Verify Deployment
Wait for all 3 pods to become Ready (1/1):

```
kubectl get pods -l app=SisyphusDB -w
```

Expected Output:

```Plaintext

NAME   READY   STATUS    RESTARTS   AGE
kv-0   1/1     Running   0          45s
kv-1   1/1     Running   0          30s
kv-2   1/1     Running   0          15s
```

🧪 Testing & Verification
Running Benchmarks
To reproduce the performance metrics (Arena vs. Map):

```
go test -bench=. -benchmem ./docs/benchmarks/arena
```
Reproducing Chaos Tests
This script continuously writes data while you simulate node failures.

Deploy the Tester Pod:

```
kubectl run tester --image=python:3.9-alpine -i --tty -- sh
apk add curl
```

Run the Monitor Script (inside the pod): (Paste the contents of docs/benchmarks/measure_recovery.py here)

```Bash
python3 measure_recovery.py
```
Inject Failure (from your host terminal):

```Bash
kubectl delete pod kv-0
```
---


