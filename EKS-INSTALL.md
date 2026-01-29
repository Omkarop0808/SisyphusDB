# EKS Deployment Guide

## Prerequisites
- AWS CLI configured (`aws configure`)
- eksctl installed
- kubectl installed
- Docker installed

---

## Step 1: Push Image to ECR

```bash
# Create ECR repo (one-time)
aws ecr create-repository --repository-name sisyphusdb --region ap-south-1

# Login to ECR
aws ecr get-login-password --region ap-south-1 | docker login --username AWS --password-stdin 967519196048.dkr.ecr.ap-south-1.amazonaws.com

# Build and push
docker build -t sisyphusdb:production .
docker tag sisyphusdb:production 967519196048.dkr.ecr.ap-south-1.amazonaws.com/sisyphusdb:production
docker push 967519196048.dkr.ecr.ap-south-1.amazonaws.com/sisyphusdb:production
```

---

## Step 2: Create EKS Cluster

```bash
eksctl create cluster \
  --name kv-store-cluster \
  --region ap-south-1 \
  --nodegroup-name standard-nodes \
  --node-type t3.small \
  --nodes 3

# Install EBS CSI driver
eksctl create addon --name aws-ebs-csi-driver --cluster kv-store-cluster --region ap-south-1 --force

# Attach IAM policies to node role
ROLE_NAME=$(aws iam list-roles --query 'Roles[?contains(RoleName, `eksctl`) && contains(RoleName, `NodeInstanceRole`)].RoleName' --output text)

aws iam attach-role-policy --role-name $ROLE_NAME --policy-arn arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy
aws iam attach-role-policy --role-name $ROLE_NAME --policy-arn arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly
```

---

## Step 3: Create Secrets and ConfigMaps

```bash
# ECR pull secret
kubectl create secret docker-registry ecr-secret \
  --docker-server=967519196048.dkr.ecr.ap-south-1.amazonaws.com \
  --docker-username=AWS \
  --docker-password=$(aws ecr get-login-password --region ap-south-1)

# Monitoring ConfigMaps
kubectl create configmap prometheus-config --from-file=prometheus.yml=deploy/prometheus/prometheus.yml
kubectl create configmap grafana-datasources --from-file=deploy/grafana/provisioning/datasources/
kubectl create configmap grafana-provisioning --from-file=deploy/grafana/provisioning/dashboards/
kubectl create configmap grafana-dashboards --from-file=deploy/grafana/dashboards/
```

---

## Step 4: Deploy

```bash
cd deploy/k8s
kubectl apply -f 1-services.yaml
kubectl apply -f 2-statefulset.yaml
kubectl apply -f 4-monitoring.yaml

# Watch pods
kubectl get pods -w
```

---

## Step 5: Expose Services Publicly

```bash
# Make KV-Store and Grafana public
kubectl patch svc kv-public -p '{"spec": {"type": "LoadBalancer"}}'
kubectl patch svc grafana -p '{"spec": {"type": "LoadBalancer"}}'

# Get URLs
kubectl get svc kv-public grafana
```

---

## Access

- **KV-Store**: `http://<ELB_URL>/put?key=hello&val=world`
- **Grafana**: `http://<ELB_URL>:3000` (admin/admin)

---

## Stop Cluster (Pause Billing)

**Option A: Delete cluster entirely (recommended for cost)**
```bash
eksctl delete cluster --name kv-store-cluster --region ap-south-1
```

**Option B: Scale nodes to 0 (keeps cluster, stops node billing)**
```bash
eksctl scale nodegroup --cluster kv-store-cluster --name standard-nodes --nodes 0 --region ap-south-1
```

**To resume Option B:**
```bash
eksctl scale nodegroup --cluster kv-store-cluster --name standard-nodes --nodes 3 --region ap-south-1
```

> Note: EKS control plane costs ~$0.10/hr even with 0 nodes. Full delete is cheapest.

---

## Resume After Delete

Just run Steps 2-5 again. Your ECR image persists.
