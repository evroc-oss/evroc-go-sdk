# k3s Kubernetes Cluster Across 3 Zones

This example demonstrates deploying a production-ready k3s Kubernetes cluster across 3 availability zones for high availability.

## What This Example Does

1. **Creates Infrastructure:**
   - Security group with k3s networking rules
   - Public IP for the control plane
   - 3 disks across zones A, B, and C

2. **Deploys k3s Cluster:**
   - 1 server (control plane) in zone A with public IP
   - 2 agents (workers) in zones B and C
   - Automatic cluster join using k3s token
   - Zone topology labels on all nodes

3. **Configuration:**
   - Disables default components (Traefik, ServiceLB, local-storage)
   - Enables kubeconfig access with proper TLS SANs
   - Sets up kubelet and Flannel networking

## Prerequisites

```bash
# Required environment variables
export EVROC_PROJECT="your-project-uuid"
export EVROC_REGION="se-sto"
export SSH_PUBLIC_KEY="ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA..."
```

## Running

```bash
cd examples/k3s-cluster

# Create the cluster (default)
go run main.go
# or explicitly:
go run main.go create

# Destroy the cluster
go run main.go destroy
```

**Timeline:**
- **VM creation**: ~2 minutes (the example waits for this)
- **Cloud-init execution**: 2-3 minutes (k3s installation - happens in background)
- **Cluster ready**: Total ~5 minutes from start

**What happens during cloud-init:**
1. VMs boot and become "ready" (example completes here)
2. Cloud-init installs packages (curl, util-linux, etc.)
3. Cloud-init downloads and installs k3s (~1-2 minutes)
4. k3s server starts and generates kubeconfig
5. k3s agents join the server
6. Cluster is fully operational

The example completes after step 1, so you must wait for steps 2-6 to finish before accessing the cluster.

## Accessing the Cluster

⚠️ **IMPORTANT**: After running the example, k3s is still installing via cloud-init. Wait 2-3 minutes before accessing the cluster.

### 1. Wait for k3s Installation

The example output provides a command to wait for k3s to be ready:

```bash
# This will wait until k3s is fully installed (replace IP with your server IP)
ssh ubuntu@<server-public-ip> 'until sudo test -f /etc/rancher/k3s/k3s.yaml; do echo "Waiting for k3s..."; sleep 5; done; echo "k3s is ready!"'
```

This command checks every 5 seconds and exits when k3s is ready.

### 2. Retrieve Kubeconfig

Once k3s is ready:

```bash
# Download kubeconfig
ssh ubuntu@<server-public-ip> sudo cat /etc/rancher/k3s/k3s.yaml > /tmp/k3s-config.yaml
```

### 3. Configure Local Access

Update the server address, skip TLS verification (certificate doesn't trust public IP), and save locally:

```bash
# Replace 127.0.0.1 with public IP and add insecure-skip-tls-verify
ssh ubuntu@<server-public-ip> sudo cat /etc/rancher/k3s/k3s.yaml | \
  sed 's/127.0.0.1/<server-public-ip>/g' | \
  sed '/certificate-authority-data:/d' | \
  sed 's/server: https/insecure-skip-tls-verify: true\n    server: https/' \
  > ~/.kube/sdk-k3s-config

# Set permissions
chmod 600 ~/.kube/sdk-k3s-config

# Use the config (IMPORTANT: Use absolute path, not ~)
export KUBECONFIG=$HOME/.kube/sdk-k3s-config
```

**Why `insecure-skip-tls-verify`?** The k3s certificate is only valid for 127.0.0.1 and the server's internal IP. When connecting via public IP, we need to skip certificate verification. For production, use a proper TLS setup with valid certificates.

### 3. Verify Cluster

```bash
# Check nodes
kubectl get nodes -o wide

# Expected output:
# NAME            STATUS   ROLES                  AGE   VERSION        ZONE
# sdk-k3s-server  Ready    control-plane,master   5m    v1.31.x-k3s1   a
# sdk-k3s-agent-1 Ready    <none>                 3m    v1.31.x-k3s1   b
# sdk-k3s-agent-2 Ready    <none>                 3m    v1.31.x-k3s1   c

# Verify zone labels
kubectl get nodes --show-labels | grep topology.kubernetes.io/zone

# Check cluster info
kubectl cluster-info
kubectl get pods -A
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Public Internet                      │
└────────────────────────┬────────────────────────────────┘
                         │
                         ▼
              ┌──────────────────────┐
              │   Security Group     │
              │  (SSH, API, Flannel) │
              └──────────────────────┘
                         │
        ┌────────────────┼────────────────┐
        │                │                │
┌───────▼────────┐ ┌────▼──────┐ ┌───────▼────────┐
│  Zone A        │ │  Zone B   │ │  Zone C        │
│                │ │           │ │                │
│ sdk-k3s-server │ │sdk-k3s-   │ │ sdk-k3s-       │
│ (Control Plane)│ │agent-1    │ │ agent-2        │
│                │ │(Worker)   │ │ (Worker)       │
│ c1a.m (4v/8GB) │ │c1a.s      │ │ c1a.s          │
│ Public IP      │ │(2v/4GB)   │ │ (2v/4GB)       │
│                │ │Private IP │ │ Private IP     │
└────────────────┘ └───────────┘ └────────────────┘
```

## Security Group Rules

The example creates these rules:

**Ingress:**
- Port 22 (SSH) - Management access
- Port 6443 (TCP) - Kubernetes API server
- Port 8472 (UDP) - Flannel VXLAN (internal subnet only)
- Port 10250 (TCP) - Kubelet metrics (internal subnet only)

**Egress:**
- All traffic allowed (for package installation and internet access)

## k3s Configuration

**Disabled Components:**
- Traefik (use your own ingress controller)
- ServiceLB (use MetalLB or cloud load balancer)
- Local-storage (use [Evroc CSI Driver](https://github.com/evroc-oss/evroc-csi-driver) for persistent volumes)

**Enabled Features:**
- Flannel CNI for pod networking
- CoreDNS for service discovery
- Metrics server for monitoring

## Testing Zone Distribution

Deploy a simple DaemonSet to verify each node is in a different zone:

```bash
# Deploy DaemonSet (one pod per node)
cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: zone-test
spec:
  selector:
    matchLabels:
      app: zone-test
  template:
    metadata:
      labels:
        app: zone-test
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
EOF

# Wait for pods to be ready
kubectl wait --for=condition=ready pod -l app=zone-test --timeout=60s

# Check pods are running (should be 3 - one per node)
kubectl get pods -l app=zone-test -o wide
```

**Verify zone distribution:**

```bash
# Show which zone each pod is running in
for pod in $(kubectl get pods -l app=zone-test -o name); do
  node=$(kubectl get $pod -o jsonpath='{.spec.nodeName}')
  zone=$(kubectl get node $node -o jsonpath='{.metadata.labels.topology\.kubernetes\.io/zone}')
  echo "$pod -> $node (zone: $zone)"
done
```

**Expected output:** 3 pods running on nodes in zones a, b, and c.

Example:
```
pod/zone-test-xxxxx -> sdk-k3s-server (zone: a)
pod/zone-test-yyyyy -> sdk-k3s-agent-1 (zone: b)
pod/zone-test-zzzzz -> sdk-k3s-agent-2 (zone: c)
```

**Cleanup:**
```bash
kubectl delete daemonset zone-test
```

## Next Steps

- **Install CSI Driver:** Deploy the [Evroc CSI Driver](https://github.com/evroc-oss/evroc-csi-driver) for persistent volumes with zone-aware dynamic provisioning
- **Setup Ingress:** Install nginx-ingress or Traefik
- **Add Monitoring:** Deploy Prometheus and Grafana
- **Scale Cluster:** Add more agents in each zone

## Cleanup

To delete the cluster and all its resources:

```bash
cd examples/k3s-cluster
go run main.go destroy
```

This will:
1. Delete all 3 VMs (server + 2 agents)
2. Wait for VMs to be fully deleted
3. Delete all 3 disks
4. Delete the public IP
5. Delete the security group

The destroy command handles the correct order and waits for resources to be properly deleted before proceeding.

## Cost Estimation

**Hourly cost (approximate):**
- Server (c1a.m): 4 vCPUs, 8GB RAM
- 2x Agents (c1a.s): 2 vCPUs, 4GB RAM each
- 3x Disks: 50GB each
- 1x Public IP

Check current pricing at: https://evroc.com/pricing

## Troubleshooting

**"No such file or directory" when retrieving kubeconfig:**

This means k3s hasn't finished installing yet. Check the installation progress:

```bash
# Check k3s installation log (recommended - shows detailed progress)
ssh ubuntu@<server-ip> 'tail -f /var/log/k3s-install.log'

# Check if k3s is installed and running
ssh ubuntu@<server-ip> 'systemctl status k3s'

# Watch cloud-init progress
ssh ubuntu@<server-ip> 'tail -f /var/log/cloud-init-output.log'

# Wait for k3s to be ready (automated check)
ssh ubuntu@<server-ip> 'until sudo test -f /etc/rancher/k3s/k3s.yaml; do echo "Waiting..."; sleep 5; done'
```

**k3s server not ready after 5 minutes:**
```bash
# SSH into server
ssh ubuntu@<server-ip>

# Check k3s service status
sudo systemctl status k3s

# Check k3s logs
sudo journalctl -u k3s -f

# Check cloud-init logs for errors
sudo cat /var/log/cloud-init-output.log
```

**Agents not joining:**
```bash
# SSH into agent
ssh ubuntu@<agent-private-ip>  # Via server as jump host

# Check k3s-agent logs
sudo journalctl -u k3s-agent -f

# Verify connectivity to server
nc -zv <server-private-ip> 6443
```

**No zone labels:**
```bash
# The labels are set during node registration
# If missing, they need to be set when k3s starts

# Verify on each node:
kubectl get nodes --show-labels
```

## References

- [k3s Documentation](https://docs.k3s.io/)
- [Kubernetes Topology Awareness](https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/)
- [Evroc CSI Driver](https://github.com/evroc-oss/evroc-csi-driver)
